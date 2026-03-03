package data

import (
	"IndicoDataFusion/backend/cache"
	"IndicoDataFusion/backend/indico"
	"IndicoDataFusion/backend/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"IndicoDataFusion/backend/config"
)

// DataSourceHandler provides a high-level interface for accessing event data
// from different sources (Indico API or local test files).
type DataSourceHandler struct {
	config     *config.DataSource
	client     *indico.IndicoClient
	dataDir    string
	isTestMode bool
	cache      *cache.Cache
	// initProblems collects non-fatal initialization issues (e.g., missing API token)
	initProblems []string
	// userID is the ID of the current user, populated from the Indico client after fetching review tracks
	userID int
	// abstractsFile is the optional override path set via --abstracts-file. When non-empty,
	// GetAbstracts reads directly from this file instead of the configured data source or API.
	abstractsFile string

	// Per-handler maps (moved from package scope)
	questions    map[int]*indico.QuestionData
	contribTypes map[string]int
	mu           sync.RWMutex
}

// NewDataSourceHandler creates a new data source handler with optional cache config
// and a list of APITokenEntry to resolve api tokens referenced by name.
func NewDataSourceHandler(ds *config.DataSource, cacheConfig *config.CacheConfig, tokens []config.APITokenEntry) (*DataSourceHandler, error) {
	handler := &DataSourceHandler{
		config: ds,
	}

	// Initialize per-handler maps
	handler.questions = make(map[int]*indico.QuestionData)
	handler.contribTypes = make(map[string]int)

	// Parse cache configuration
	ttl := 24 * time.Hour               // Default: 24 hours
	maxSize := int64(100 * 1024 * 1024) // Default: 100 MB
	cacheDir := ""                      // Use default if not specified

	if cacheConfig != nil {
		if cacheConfig.TTL != "" {
			if parsedTTL, err := time.ParseDuration(cacheConfig.TTL); err == nil {
				ttl = parsedTTL
			} else {
				log.Printf("Warning: invalid cache TTL '%s', using default 24h", cacheConfig.TTL)
			}
		}
		if cacheConfig.MaxSize != "" {
			if parsedSize, err := parseSize(cacheConfig.MaxSize); err == nil {
				maxSize = parsedSize
			} else {
				log.Printf("Warning: invalid cache max_size '%s', using default 100MB", cacheConfig.MaxSize)
			}
		}
		if cacheConfig.CacheDir != "" {
			cacheDir = cacheConfig.CacheDir
		}
	}

	// Initialize cache
	cobj, err := cache.NewCache(cache.CacheOptions{
		CacheDir:       cacheDir,
		LoadOnStartup:  true,
		TTL:            ttl,
		MaxSize:        maxSize,
		DataSourceName: ds.Name,
	})
	if err != nil {
		log.Printf("Warning: failed to initialize cache: %v", err)
		// Continue without cache
	}
	handler.cache = cobj

	if ds.Indico != nil {
		// Resolve API token: require a token name and find it in provided tokens by Name only.
		apiToken := ""
		// APITokenName is required now
		if ds.Indico.APITokenName == "" {
			// Previously this was a hard error; make it a non-fatal init problem so the
			// app can continue and the UI can prompt the user to add a token.
			handler.initProblems = append(handler.initProblems, fmt.Sprintf("data source %s: missing api_token_name", ds.Name))
			log.Printf("Warning: %s", handler.initProblems[len(handler.initProblems)-1])
			// Leave client nil so API calls will return a clear error; UI can add the token
			handler.isTestMode = false
			return handler, nil
		}
		// tokens must be provided to resolve api_token_name
		if tokens == nil {
			// Record a non-fatal problem and continue; the UI can prompt the user to supply tokens
			handler.initProblems = append(handler.initProblems, fmt.Sprintf("api tokens not provided for data source %s; api_token_name %s cannot be resolved", ds.Name, ds.Indico.APITokenName))
			log.Printf("Warning: %s", handler.initProblems[len(handler.initProblems)-1])
			// Leave client nil so API calls will return a clear error; app/UI can fix tokens and reinitialize handler
			handler.isTestMode = false
			return handler, nil
		}
		// Match by Name only
		target := ds.Indico.APITokenName
		var matched *config.APITokenEntry
		for _, e := range tokens {
			if e.Name == target {
				// copy the entry so we can inspect it
				entry := e
				matched = &entry
				break
			}
		}
		if matched == nil {
			// Not found in config list - record problem and continue
			handler.initProblems = append(handler.initProblems, fmt.Sprintf("api token %q for data source %s not found in provided api-tokens", ds.Indico.APITokenName, ds.Name))
			log.Printf("Warning: %s", handler.initProblems[len(handler.initProblems)-1])
			// Leave client nil so API calls will return a clear error; UI can add the token
			handler.isTestMode = false
			return handler, nil
		}

		// Prefer token in config entry
		apiToken = matched.Token

		// If token is empty in config, try keyring
		if apiToken == "" {
			if secret, err := utils.GetAPITokenSecret(matched.Name); err == nil {
				apiToken = secret
			} else {
				// keyring lookup failed or not present - record problem but continue
				handler.initProblems = append(handler.initProblems, fmt.Sprintf("api token %q for data source %s has no token in config or keyring", ds.Indico.APITokenName, ds.Name))
				log.Printf("Warning: %s", handler.initProblems[len(handler.initProblems)-1])
				// Leave client nil so API calls will return a clear error; UI can add the token
				handler.isTestMode = false
				return handler, nil
			}
		}

		// Initialize Indico client
		client := indico.NewIndicoClient(
			ds.Indico.BaseURL,
			ds.Indico.EventID,
			apiToken,
		)
		if ds.Indico.Timeout != "" {
			if timeout, err := time.ParseDuration(ds.Indico.Timeout); err == nil {
				client.Timeout = timeout
				// Also update the HTTP client timeout to match
				client.Client = &http.Client{Timeout: timeout}
			}
		}
		handler.client = client
		handler.isTestMode = false
		log.Printf("Used Indico API token for %s: (%s) [REDACTED]", ds.Name, ds.Indico.APITokenName)
	} else if ds.Test != nil {
		// Test mode with local files
		handler.dataDir, _ = filepath.Abs(ds.Test.DataDir)
		handler.isTestMode = true
	} else {
		return nil, fmt.Errorf("data source %s has no valid configuration", ds.Name)
	}

	return handler, nil
}

// GetInitProblems returns any non-fatal initialization problems encountered when creating the handler.
func (h *DataSourceHandler) GetInitProblems() []string {
	if h == nil {
		return nil
	}
	return h.initProblems
}

// SetAbstractsFile sets a file path that overrides abstract data retrieval.
// When non-empty, all GetAbstracts calls read from this file instead of the
// configured data source or the Indico API.  The file must be a JSON file in
// the format returned by IndicoClient.FetchAbstractsData (i.e. an object with
// top-level "abstracts" and optional "questions" arrays).
func (h *DataSourceHandler) SetAbstractsFile(path string) {
	h.abstractsFile = path
}

// getAbstractsFromOverrideFile reads abstract data from the file specified by
// h.abstractsFile.  The file is expected to contain the raw JSON payload
// returned by FetchAbstractsData – an object with "abstracts" (array) and
// optional "questions" (array) fields.
// When an Indico client is available it also fetches the caller's review
// assignments (via GetReviewTracks) and merges IsMyReview / MyReview into
// each abstract, mirrors the post-processing done by getAbstractsFromAPI.
func (h *DataSourceHandler) getAbstractsFromOverrideFile(ctx context.Context) ([]indico.AbstractData, error) {
	filePath := h.abstractsFile
	log.Printf("Reading abstract data from override file: %v\n", filePath)

	rawBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read abstracts override file %s: %w", filePath, err)
	}

	var response indico.AbstractsResponse
	if err := json.Unmarshal(rawBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse abstracts override file %s: %w", filePath, err)
	}

	h.parseAbstractsResponse(&response)
	h.enrichAbstractsWithClient(ctx, response.Abstracts)

	// enrichAbstractsWithClient sets IsMyReview by querying review-track assignments
	// and populates MyReview from Reviews[] already in the file snapshot.  Because
	// the file is static the in-file review data may be stale, so we scrape the
	// live review page for every abstract assigned to the current user and upsert
	// the result so that MyReview always reflects the latest submitted state.
	// Up to 8 goroutines run in parallel to amortise per-abstract HTTP latency.
	if h.client != nil {
		type scrapeResult struct {
			idx    int
			review *indico.Review // nil → not yet reviewed
			err    error
		}

		const maxWorkers = 8

		// Collect indices of abstracts that need scraping.
		var targets []int
		for i := range response.Abstracts {
			if response.Abstracts[i].IsMyReview {
				targets = append(targets, i)
			}
		}

		if len(targets) > 0 {
			log.Printf("getAbstractsFromOverrideFile: scraping live reviews for %d abstracts (up to %d goroutines)", len(targets), maxWorkers)

			workCh := make(chan int, len(targets))
			for _, idx := range targets {
				workCh <- idx
			}
			close(workCh)

			resultCh := make(chan scrapeResult, len(targets))

			nWorkers := maxWorkers
			if len(targets) < nWorkers {
				nWorkers = len(targets)
			}
			var wg sync.WaitGroup
			for range nWorkers {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for idx := range workCh {
						review, err := h.scrapeMyReview(ctx, response.Abstracts[idx].ID)
						resultCh <- scrapeResult{idx: idx, review: review, err: err}
					}
				}()
			}
			wg.Wait()
			close(resultCh)

			// Apply results sequentially – no concurrent writes to response.Abstracts.
			for r := range resultCh {
				if r.err != nil {
					log.Printf("Warning: getAbstractsFromOverrideFile: failed to scrape review for abstract %d: %v", response.Abstracts[r.idx].ID, r.err)
					continue
				}
				if r.review == nil {
					// Not yet reviewed – clear any stale MyReview from the file.
					response.Abstracts[r.idx].MyReview = nil
					continue
				}
				// Upsert the live review into Reviews[] by review ID.
				upserted := false
				for j := range response.Abstracts[r.idx].Reviews {
					if response.Abstracts[r.idx].Reviews[j].ID == r.review.ID {
						response.Abstracts[r.idx].Reviews[j] = *r.review
						upserted = true
						break
					}
				}
				if !upserted {
					response.Abstracts[r.idx].Reviews = append(response.Abstracts[r.idx].Reviews, *r.review)
				}
				reviewCopy := *r.review
				response.Abstracts[r.idx].MyReview = &reviewCopy
				log.Printf("getAbstractsFromOverrideFile: scraped live review %d for abstract %d", r.review.ID, response.Abstracts[r.idx].ID)
			}
		}
	}

	return response.Abstracts, nil
}

// parseSize parses size strings like "100MB", "1GB", "512KB"
func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Check suffixes in order from longest to shortest to avoid "MB" matching "B"
	suffixes := []struct {
		suffix     string
		multiplier int64
	}{
		{"GB", 1024 * 1024 * 1024},
		{"MB", 1024 * 1024},
		{"KB", 1024},
		{"B", 1},
	}

	for _, s := range suffixes {
		if strings.HasSuffix(sizeStr, s.suffix) {
			numStr := strings.TrimSpace(strings.TrimSuffix(sizeStr, s.suffix))
			var num float64
			if _, err := fmt.Sscanf(numStr, "%f", &num); err != nil {
				return 0, fmt.Errorf("invalid size format: %s", sizeStr)
			}
			return int64(num * float64(s.multiplier)), nil
		}
	}

	return 0, fmt.Errorf("invalid size format: %s (use B, KB, MB, or GB)", sizeStr)
}

// NewDataSourceHandlerFromConfig creates a handler from a full Config using the default data source.
func NewDataSourceHandlerFromConfig(cfg *config.Config) (*DataSourceHandler, error) {
	// Validate config before using it
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	ds, err := cfg.GetActiveDataSource()
	if err != nil {
		return nil, fmt.Errorf("failed to get active data source: %w", err)
	}
	// Forward cfg.APITokens so token resolution uses tokens loaded from the config
	return NewDataSourceHandler(ds, cfg.Cache, cfg.APITokens)
}

// NewDataSourceHandlerFromConfigFile creates a handler by loading a config file.
func NewDataSourceHandlerFromConfigFile(configPath string) (*DataSourceHandler, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return NewDataSourceHandlerFromConfig(cfg)
}

// Shutdown gracefully shuts down the handler, saving cache to disk
func (h *DataSourceHandler) Shutdown(ctx context.Context) error {
	// If receiver is nil, nothing to do
	if h == nil {
		return nil
	}

	// Clear per-handler maps to release memory and avoid cross-handler contamination
	h.mu.Lock()
	h.questions = nil
	h.contribTypes = nil
	h.mu.Unlock()

	if h.cache != nil {
		return h.cache.Shutdown(ctx)
	}
	return nil
}

// GetClient returns the Indico client for direct API access.
// Returns nil if in test mode or client is not initialized.
func (h *DataSourceHandler) GetClient() *indico.IndicoClient {
	return h.client
}

// GetInfo retrieves event information from the configured data source.
func (h *DataSourceHandler) GetInfo(ctx context.Context) (*indico.Event, error) {
	if h.isTestMode {
		return h.getInfoFromFile()
	}

	// Check cache first in API mode
	if h.cache != nil {
		if cached, found := h.cache.Get("event_info"); found {
			log.Printf("Using cached event info (type: %T)", cached)
			// Try direct type assertion first
			if event, ok := cached.(*indico.Event); ok {
				log.Printf("Successfully retrieved event info from cache")
				return event, nil
			}
			// Type assertion failed - likely due to JSON unmarshaling
			// Try to convert via JSON re-marshaling
			log.Printf("Direct type assertion failed, attempting JSON conversion...")
			jsonData, err := json.Marshal(cached)
			if err == nil {
				var event indico.Event
				if err := json.Unmarshal(jsonData, &event); err == nil {
					log.Printf("Successfully converted event info from cache via JSON")
					return &event, nil
				}
			}
			// Conversion failed - cache data corrupted, delete and refetch
			log.Printf("Warning: cached event_info has wrong type and conversion failed (expected *Event, got %T), deleting and refetching", cached)
			h.cache.Delete("event_info")
		} else {
			log.Printf("No cached event_info found for this data source")
		}
	}

	// Fetch from API
	event, err := h.getInfoFromAPI(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if h.cache != nil {
		h.cache.Set("event_info", event)
	}

	return event, nil
}

// getInfoFromFile reads event info from a local JSON file (test mode).
func (h *DataSourceHandler) getInfoFromFile() (*indico.Event, error) {
	if h.config.Test == nil {
		return nil, fmt.Errorf("test configuration not available")
	}

	filePath := filepath.Join(h.dataDir, h.config.Test.EventInfo)
	log.Printf("Reading event info from: %v\n", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	var ev indico.EventAPIResponse
	if err := json.Unmarshal(data, &ev); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	return &ev.Results[0], nil
}

// getInfoFromAPI fetches event info from the Indico API.
func (h *DataSourceHandler) getInfoFromAPI(ctx context.Context) (*indico.Event, error) {
	_ = ctx // referenced to avoid unused parameter warning
	if h.client == nil {
		return nil, fmt.Errorf("indico client not initialized")
	}
	log.Printf("Reading event info from Indico API\n")
	return h.client.GetEventInfo()
}

// GetAbstracts retrieves abstract data from the configured data source.
func (h *DataSourceHandler) GetAbstracts(ctx context.Context) ([]indico.AbstractData, error) {
	// --abstracts-file override: read from the local file and enrich with live review
	// data, then cache.  The cache is populated the same way as API mode so that
	// RefreshCache (triggered by the Refresh button) can invalidate it and force a
	// re-read of both the file and the live review-assignment API.
	if h.abstractsFile != "" {
		if h.cache != nil {
			if cached, found := h.cache.Get("abstracts"); found {
				if abstracts, ok := cached.([]indico.AbstractData); ok {
					log.Printf("Using cached abstracts (abstracts-file mode): %d entries", len(abstracts))
					return abstracts, nil
				}
				// Stale/corrupt entry – discard and re-read.
				h.cache.Delete("abstracts")
			}
		}
		abstracts, err := h.getAbstractsFromOverrideFile(ctx)
		if err != nil {
			return nil, err
		}
		if h.cache != nil {
			h.cache.Set("abstracts", abstracts)
		}
		return abstracts, nil
	}

	if h.isTestMode {
		return h.getAbstractsFromFile()
	}

	// Check cache first in API mode
	if h.cache != nil {
		if cached, found := h.cache.Get("abstracts"); found {
			log.Printf("Using cached abstracts (type: %T)", cached)
			// Try direct type assertion first
			if abstracts, ok := cached.([]indico.AbstractData); ok {
				log.Printf("Successfully retrieved %d abstracts from cache", len(abstracts))
				return abstracts, nil
			}
			// Type assertion failed - likely due to JSON unmarshaling
			// Try to convert via JSON re-marshaling
			log.Printf("Direct type assertion failed, attempting JSON conversion...")
			jsonData, err := json.Marshal(cached)
			if err == nil {
				var abstracts []indico.AbstractData
				if err := json.Unmarshal(jsonData, &abstracts); err == nil {
					log.Printf("Successfully converted %d abstracts from cache via JSON", len(abstracts))
					return abstracts, nil
				}
			}
			// Conversion failed - cache data corrupted, delete and refetch
			log.Printf("Warning: cached abstracts has wrong type and conversion failed (expected []AbstractData, got %T), deleting and refetching", cached)
			h.cache.Delete("abstracts")
		} else {
			log.Printf("No cached abstracts found for this data source")
		}
	}

	// Fetch from API
	abstracts, err := h.getAbstractsFromAPI(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if h.cache != nil {
		h.cache.Set("abstracts", abstracts)
	}

	return abstracts, nil
}

// getAbstractsFromFile reads abstracts from a local JSON file (test mode).
func (h *DataSourceHandler) getAbstractsFromFile() ([]indico.AbstractData, error) {
	if h.config.Test == nil {
		return nil, fmt.Errorf("test configuration not available")
	}

	filePath := filepath.Join(h.dataDir, h.config.Test.Abstracts)
	log.Printf("Reading abstract data from: %v\n", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	var response indico.AbstractsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	h.parseAbstractsResponse(&response)
	return response.Abstracts, nil
}

// getAbstractsFromAPI fetches abstracts from the Indico API.
func (h *DataSourceHandler) getAbstractsFromAPI(ctx context.Context) ([]indico.AbstractData, error) {
	if h.client == nil {
		return nil, fmt.Errorf("indico client not initialized")
	}

	log.Printf("Reading abstract data from Indico API\n")

	ids, err := h.client.GetAbstractIDsAndCSRFFromList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get abstract IDs: %w", err)
	}

	if len(ids) == 0 {
		return []indico.AbstractData{}, nil
	}

	rawData, err := h.client.FetchAbstractsData(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch abstracts data: %w", err)
	}

	response, err := h.unmarshalAbstractsRaw(rawData)
	if err != nil {
		return nil, err
	}

	h.parseAbstractsResponse(response)
	h.enrichAbstractsWithClient(ctx, response.Abstracts)
	return response.Abstracts, nil
}

// unmarshalAbstractsRaw converts the raw map[string]any returned by FetchAbstractsData
// into an AbstractsResponse via a JSON round-trip.
func (h *DataSourceHandler) unmarshalAbstractsRaw(rawData map[string]any) (*indico.AbstractsResponse, error) {
	abstractsIface, ok := rawData["abstracts"]
	if !ok {
		return nil, fmt.Errorf("no abstracts field in response")
	}
	jsonData, err := json.Marshal(map[string]any{
		"abstracts": abstractsIface,
		"questions": rawData["questions"],
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal abstracts: %w", err)
	}
	var response indico.AbstractsResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal abstracts: %w", err)
	}
	return &response, nil
}

// parseAbstractsResponse populates the handler's question and contrib-type lookup maps
// from the response, then expands question details into every review rating and
// computes the aggregated priority ratings for each abstract.
// It is safe to call from any fetch path (file, API, override).
func (h *DataSourceHandler) parseAbstractsResponse(response *indico.AbstractsResponse) {
	// Build question lookup map.
	if len(response.Questions) > 0 {
		h.mu.Lock()
		for i := range response.Questions {
			q := &response.Questions[i]
			h.questions[q.ID] = q
		}
		h.mu.Unlock()
	}

	// Build contribution types map.
	if len(response.Abstracts) > 0 {
		h.mu.Lock()
		for i := range response.Abstracts {
			if ct := response.Abstracts[i].AcceptedContribType; ct != nil {
				h.contribTypes[ct.Name] = ct.ID
			}
			if ct := response.Abstracts[i].SubmittedContribType; ct != nil {
				h.contribTypes[ct.Name] = ct.ID
			}
			for j := range response.Abstracts[i].Reviews {
				if ct := response.Abstracts[i].Reviews[j].ProposedContribType; ct != nil {
					h.contribTypes[ct.Name] = ct.ID
				}
			}
		}
		h.mu.Unlock()
	}

	// Expand question details into ratings and attach shared maps.
	for i := range response.Abstracts {
		h.mu.RLock()
		response.Abstracts[i].Questions = h.questions
		response.Abstracts[i].ContribTypesMap = &h.contribTypes
		h.mu.RUnlock()

		for j := range response.Abstracts[i].Reviews {
			for k := range response.Abstracts[i].Reviews[j].Ratings {
				rating := &response.Abstracts[i].Reviews[j].Ratings[k]
				h.mu.RLock()
				if q, ok := h.questions[rating.Question]; ok {
					rating.QuestionDetails = q
				}
				h.mu.RUnlock()
			}
		}
		response.Abstracts[i].FirstPriority = response.Abstracts[i].GetAggregatedRatingByTitle("First priority")
		response.Abstracts[i].SecondPriority = response.Abstracts[i].GetAggregatedRatingByTitle("Second priority")
	}
}

// enrichAbstractsWithClient merges review-assignment data (IsMyReview / MyReview) and,
// when a live client is available, also sets IndicoURL and absolutises avatar URLs.
// It also overrides ReviewedForTracks from the live review track assignments so
// that abstracts loaded from a static file (--abstracts-file) carry up-to-date
// track info.
// It is a no-op when h.client is nil (test / offline mode).
func (h *DataSourceHandler) enrichAbstractsWithClient(ctx context.Context, abstracts []indico.AbstractData) {
	ra := getReviewAssignments(h, ctx)
	if ra != nil {
		for i := range abstracts {
			abstracts[i].IsMyReview = ra.isMyReview[abstracts[i].FriendlyID]
			populateMyReview(h, &abstracts[i])

			// Override ReviewedForTracks with tracks from the live review
			// assignments.
			if assignedTracks, ok := ra.tracksByAbstract[abstracts[i].FriendlyID]; ok {
				abstracts[i].ReviewedForTracks = make([]indico.Track, len(assignedTracks))
				copy(abstracts[i].ReviewedForTracks, assignedTracks)
			}
		}
	}

	if h.client == nil {
		return
	}

	for i := range abstracts {
		abstracts[i].IndicoURL = fmt.Sprintf("%s/event/%d/abstracts/%d",
			h.client.BaseURL, h.client.EventID, abstracts[i].ID)

		for j := range abstracts[i].Reviews {
			av := abstracts[i].Reviews[j].User.AvatarURL
			if av != "" && !strings.HasPrefix(av, "http") {
				abstracts[i].Reviews[j].User.AvatarURL = h.client.BaseURL + av
			}
		}
		if abstracts[i].Submitter != nil {
			if sav := abstracts[i].Submitter.AvatarURL; sav != "" && !strings.HasPrefix(sav, "http") {
				abstracts[i].Submitter.AvatarURL = h.client.BaseURL + sav
			}
		}
		if abstracts[i].Judge != nil {
			if jav := abstracts[i].Judge.AvatarURL; jav != "" && !strings.HasPrefix(jav, "http") {
				abstracts[i].Judge.AvatarURL = h.client.BaseURL + jav
			}
		}
	}
}

// GetContributions retrieves contribution data from the configured data source.
func (h *DataSourceHandler) GetContributions(ctx context.Context) ([]indico.ContributionData, error) {
	if h.isTestMode {
		return h.getContributionsFromFile()
	}

	// Check cache first in API mode
	if h.cache != nil {
		if cached, found := h.cache.Get("contributions"); found {
			log.Printf("Using cached contributions (type: %T)", cached)
			// Try direct type assertion first
			if contribs, ok := cached.([]indico.ContributionData); ok {
				log.Printf("Successfully retrieved %d contributions from cache", len(contribs))
				return contribs, nil
			}
			// Type assertion failed - likely due to JSON unmarshaling
			// Try to convert via JSON re-marshaling
			log.Printf("Direct type assertion failed, attempting JSON conversion...")
			jsonData, err := json.Marshal(cached)
			if err == nil {
				var contribs []indico.ContributionData
				if err := json.Unmarshal(jsonData, &contribs); err == nil {
					log.Printf("Successfully converted %d contributions from cache via JSON", len(contribs))
					return contribs, nil
				}
			}
			// Conversion failed - cache data corrupted, delete and refetch
			log.Printf("Warning: cached contributions has wrong type and conversion failed (expected []ContributionData, got %T), deleting and refetching", cached)
			h.cache.Delete("contributions")
		} else {
			log.Printf("No cached contributions found for this data source")
		}
	}

	// Fetch from API
	contribs, err := h.getContributionsFromAPI(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if h.cache != nil {
		h.cache.Set("contributions", contribs)
	}

	return contribs, nil
}

// getContributionsFromFile reads contributions from a local JSON file (test mode).
func (h *DataSourceHandler) getContributionsFromFile() ([]indico.ContributionData, error) {
	if h.config.Test == nil {
		return nil, fmt.Errorf("test configuration not available")
	}

	filePath := filepath.Join(h.dataDir, h.config.Test.Contribs)
	log.Printf("Reading contribution data from: %v\n", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	var response indico.ContributionsAPIResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	if len(response.Results) == 0 {
		return []indico.ContributionData{}, nil
	}

	return response.Results[0].Contributions, nil
}

// getContributionsFromAPI fetches contributions from the Indico API.
func (h *DataSourceHandler) getContributionsFromAPI(ctx context.Context) ([]indico.ContributionData, error) {
	if h.client == nil {
		return nil, fmt.Errorf("indico client not initialized")
	}

	log.Printf("Reading contribution data from Indico API\n")

	// Construct the API path for contributions
	path := fmt.Sprintf("/export/event/%d.json", h.client.EventID)

	queryValues := url.Values{}
	queryValues.Set("detail", "contributions")

	// Fetch the contribution data
	var response indico.ContributionsAPIResponse
	if err := h.client.DoGet(ctx, path, queryValues, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch contributions: %w", err)
	}

	if len(response.Results) == 0 {
		return []indico.ContributionData{}, nil
	}

	return response.Results[0].Contributions, nil
}

// GetContributionByID retrieves a specific contribution by ID.
func (h *DataSourceHandler) GetContributionByID(ctx context.Context, id string) (*indico.ContributionData, error) {
	contributions, err := h.GetContributions(ctx)
	if err != nil {
		return nil, err
	}

	for _, contrib := range contributions {
		if contrib.ID == id {
			return &contrib, nil
		}
	}

	return nil, fmt.Errorf("contribution with ID %s not found", id)
}

// GetAbstractByID retrieves a specific abstract by ID.
func (h *DataSourceHandler) GetAbstractByID(ctx context.Context, id int) (*indico.AbstractData, error) {
	abstracts, err := h.GetAbstracts(ctx)
	if err != nil {
		return nil, err
	}

	for _, abstract := range abstracts {
		if abstract.ID == id {
			return &abstract, nil
		}
	}

	return nil, fmt.Errorf("abstract with ID %d not found", id)
}

// reviewAssignments holds the result of fetching review track assignments:
// a set of abstract friendly IDs on the review list, and a mapping from
// each friendly ID to the review tracks it belongs to.
type reviewAssignments struct {
	// isMyReview maps abstract friendly_id → true for abstracts on the review list.
	isMyReview map[int]bool
	// tracksByAbstract maps abstract friendly_id → review Track(s) the abstract
	// is assigned to for this reviewer.  Used to update ReviewedForTracks when
	// the abstract data comes from a static file (--abstracts-file mode).
	tracksByAbstract map[int][]indico.Track
}

// getReviewAssignments fetches the reviewer's track assignments via
// GetReviewTracks + GetReviewAbstractIDs and returns both the review-ID set
// and per-abstract track mappings.  Returns nil when h.client is nil or when
// GetReviewTracks fails.
func getReviewAssignments(h *DataSourceHandler, ctx context.Context) *reviewAssignments {
	if h.client == nil {
		return nil
	}

	// Fetch review tracks and mark abstracts that are on my review list
	reviewTracks, err := h.client.GetReviewTracks(ctx)
	if err != nil {
		log.Printf("Warning: failed to fetch review tracks for refresh: %v", err)
		return nil
	}

	if reviewTracks == nil || len(reviewTracks.Tracks) == 0 {
		return nil
	}

	ra := &reviewAssignments{
		isMyReview:       make(map[int]bool),
		tracksByAbstract: make(map[int][]indico.Track),
	}
	for _, rt := range reviewTracks.Tracks {
		if rt.Link == "" {
			continue
		}
		abstractIDs, err := h.client.GetReviewAbstractIDs(ctx, rt.TrackID)
		if err != nil {
			log.Printf("Warning: failed to get abstract IDs for track %d: %v", rt.TrackID, err)
			continue
		}
		// Build a Track from the ReviewTrack info.
		// ReviewTrack.Name is e.g. "MC7 Accelerator Technology Main Systems: MC7.2"
		// Extract the track code (the part after the last ": ").
		trackCode := ""
		if idx := strings.LastIndex(rt.Name, ": "); idx >= 0 {
			trackCode = strings.TrimSpace(rt.Name[idx+2:])
		}
		track := indico.Track{
			ID:    rt.TrackID,
			Code:  trackCode,
			Title: rt.Name,
		}
		for _, aid := range abstractIDs {
			ra.isMyReview[aid] = true
			ra.tracksByAbstract[aid] = append(ra.tracksByAbstract[aid], track)
		}
	}
	return ra
}

// populateMyReview finds and sets the current user's review in the abstract.
// It matches reviews by comparing the review user's ID with the current user's ID.
func populateMyReview(h *DataSourceHandler, abstract *indico.AbstractData) {
	// Sync user ID from client if available
	if h.client != nil && h.client.UserID > 0 {
		h.userID = h.client.UserID
	}

	if h.userID == 0 {
		// No user ID available, cannot match reviews
		return
	}

	// Match reviews by user ID
	for i := range abstract.Reviews {
		if abstract.Reviews[i].User.ID == h.userID {
			// Create a copy of the review
			reviewCopy := abstract.Reviews[i]
			abstract.MyReview = &reviewCopy
			return
		}
	}
}

// scrapeMyReview fetches the current user's review for a single abstract by
// visiting its display page, expands question details from the handler's
// question map, and returns the populated Review (or nil when not yet reviewed).
// It is a no-op when h.client is nil.
func (h *DataSourceHandler) scrapeMyReview(ctx context.Context, abstractID int) (*indico.Review, error) {
	if h.client == nil {
		return nil, fmt.Errorf("indico client not initialized")
	}

	h.mu.RLock()
	questionsByTitle := make(map[string]int, len(h.questions))
	for qid, q := range h.questions {
		questionsByTitle[strings.ToLower(q.Title)] = qid
	}
	// Build contribTypesByName: name → ID, for resolving ProposedContribType in
	// accept reviews that carry "as <strong>Name</strong>" in the HTML badge.
	contribTypesByName := make(map[string]int, len(h.contribTypes))
	for name, id := range h.contribTypes {
		contribTypesByName[name] = id
	}
	h.mu.RUnlock()

	// Build lookup maps for ParseReviewFromHTML from cached abstracts.
	// abstractsByDBID maps abstract DB ID → RelatedAbstract (for mark_as_duplicate/merge).
	// tracksByCode maps track code → Track (for change_tracks proposed tracks).
	var abstractsByDBID map[int]*indico.RelatedAbstract
	var tracksByCode map[string]*indico.Track
	if h.cache != nil {
		if cached, found := h.cache.Get("abstracts"); found {
			// Mirror the same type-assertion + JSON-fallback logic used in
			// GetAbstracts so friendly_id is resolved correctly even after a
			// cache JSON round-trip (where the value may be []interface{}).
			var abstracts []indico.AbstractData
			if typed, ok := cached.([]indico.AbstractData); ok {
				abstracts = typed
			} else {
				if jsonData, err := json.Marshal(cached); err == nil {
					_ = json.Unmarshal(jsonData, &abstracts)
				}
			}
			if len(abstracts) > 0 {
				abstractsByDBID = make(map[int]*indico.RelatedAbstract, len(abstracts))
				tracksByCode = make(map[string]*indico.Track)
				for i := range abstracts {
					a := &abstracts[i]
					abstractsByDBID[a.ID] = &indico.RelatedAbstract{
						ID:         a.ID,
						FriendlyID: a.FriendlyID,
						Title:      a.Title,
					}
					// Collect tracks from submitted_for_tracks and reviewed_for_tracks.
					for j := range a.SubmittedForTracks {
						t := &a.SubmittedForTracks[j]
						if t.Code != "" {
							tracksByCode[t.Code] = t
						}
					}
					for j := range a.ReviewedForTracks {
						t := &a.ReviewedForTracks[j]
						if t.Code != "" {
							tracksByCode[t.Code] = t
						}
					}
				}
			}
		}
	}

	review, err := h.client.GetReviewFromAbstractPage(ctx, abstractID, questionsByTitle, abstractsByDBID, tracksByCode, contribTypesByName)
	if err != nil {
		return nil, err
	}
	if review == nil {
		return nil, nil
	}

	// Expand question details from the handler's question map.
	h.mu.RLock()
	for k := range review.Ratings {
		if qd, ok := h.questions[review.Ratings[k].Question]; ok {
			review.Ratings[k].QuestionDetails = qd
		}
	}
	h.mu.RUnlock()

	return review, nil
}

// upsertAbstractInCache updates or appends an abstract in the cached abstracts
// slice.  It is a no-op when no cache is configured or the "abstracts" cache
// entry does not exist yet.
func (h *DataSourceHandler) upsertAbstractInCache(abstract indico.AbstractData) {
	if h.cache == nil {
		return
	}
	cached, found := h.cache.Get("abstracts")
	if !found {
		return
	}
	var abstracts []indico.AbstractData
	if cachedAbstracts, ok := cached.([]indico.AbstractData); ok {
		abstracts = cachedAbstracts
	} else {
		b, err := json.Marshal(cached)
		if err == nil {
			_ = json.Unmarshal(b, &abstracts)
		}
	}
	for i := range abstracts {
		if abstracts[i].ID == abstract.ID {
			abstracts[i] = abstract
			h.cache.Set("abstracts", abstracts)
			log.Printf("Updated abstract %d in cache", abstract.ID)
			return
		}
	}
	abstracts = append(abstracts, abstract)
	h.cache.Set("abstracts", abstracts)
	log.Printf("Inserted abstract %d into cache", abstract.ID)
}

// RefreshAbstractByID fetches fresh data for a single abstract from the API.
// This bypasses the cache and always fetches from the live API.
//
// Special case – abstractsFile mode (--abstracts-file flag):
// When abstract data is served from a local file the abstract fields themselves
// are not re-fetched from the API; instead the existing abstract is taken from
// the current data set.  The review page is still fetched so that MyReview is
// updated after a submit / update, allowing the UI to show "Submit review" vs
// "Update review" correctly.
func (h *DataSourceHandler) RefreshAbstractByID(ctx context.Context, id int) (*indico.AbstractData, error) {
	// ── abstractsFile mode: keep abstract data from file, refresh review only ─
	if h.abstractsFile != "" {
		if h.client == nil {
			return nil, fmt.Errorf("indico client not initialized (cannot refresh review in abstracts-file mode)")
		}

		log.Printf("RefreshAbstractByID (abstracts-file mode): refreshing review for abstract %d", id)

		// Retrieve existing abstract from the cached/file data set.
		existing, err := h.GetAbstractByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("abstracts-file mode: abstract %d not found: %w", id, err)
		}
		// Work on a copy so we don't mutate the cached slice in place.
		abstract := *existing

		// Override ReviewedForTracks and IsMyReview from the live review track
		// assignments so the abstract carries the same enrichment as when
		// loaded via getAbstractsFromOverrideFile.
		ra := getReviewAssignments(h, ctx)
		if ra != nil {
			abstract.IsMyReview = ra.isMyReview[abstract.FriendlyID]
			if assignedTracks, ok := ra.tracksByAbstract[abstract.FriendlyID]; ok {
				// override ReviewedForTracks with the latest assignments (not append)
				abstract.ReviewedForTracks = make([]indico.Track, len(assignedTracks))
				copy(abstract.ReviewedForTracks, assignedTracks)
			}
		}

		review, err := h.scrapeMyReview(ctx, abstract.ID)
		if err != nil {
			log.Printf("Warning: RefreshAbstractByID (abstracts-file mode): failed to fetch review page for abstract %d: %v", abstract.ID, err)
		} else if review != nil {
			// Upsert into abstract.Reviews by review ID.
			upserted := false
			for j := range abstract.Reviews {
				if abstract.Reviews[j].ID == review.ID {
					abstract.Reviews[j] = *review
					upserted = true
					break
				}
			}
			if !upserted {
				abstract.Reviews = append(abstract.Reviews, *review)
			}
			reviewCopy := *review
			abstract.MyReview = &reviewCopy
			abstract.IsMyReview = true
			log.Printf("RefreshAbstractByID (abstracts-file mode): updated MyReview (review %d) for abstract %d", review.ID, abstract.ID)
		} else {
			abstract.MyReview = nil
		}

		h.upsertAbstractInCache(abstract)
		return &abstract, nil
	}

	// ── normal API mode ────────────────────────────────────────────────────────
	if h.client == nil {
		return nil, fmt.Errorf("indico client not initialized (test mode or no API access)")
	}

	log.Printf("Refreshing abstract %d from Indico API\n", id)

	rawData, err := h.client.FetchAbstractsData(ctx, []string{fmt.Sprintf("%d", id)})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch abstract %d: %w", id, err)
	}

	response, err := h.unmarshalAbstractsRaw(rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse abstract %d: %w", id, err)
	}

	if len(response.Abstracts) == 0 {
		return nil, fmt.Errorf("abstract with ID %d not found in API response", id)
	}

	h.parseAbstractsResponse(response)
	h.enrichAbstractsWithClient(ctx, response.Abstracts)

	abstract := response.Abstracts[0]
	h.upsertAbstractInCache(abstract)

	log.Printf("Successfully refreshed abstract %d from API\n", id)
	return &abstract, nil
}

// GetAbstractsByState filters abstracts by their state.
func (h *DataSourceHandler) GetAbstractsByState(ctx context.Context, state string) ([]indico.AbstractData, error) {
	abstracts, err := h.GetAbstracts(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []indico.AbstractData
	for _, abstract := range abstracts {
		if abstract.State == state {
			filtered = append(filtered, abstract)
		}
	}

	return filtered, nil
}

// GetContributionsBySession filters contributions by session.
func (h *DataSourceHandler) GetContributionsBySession(ctx context.Context, session string) ([]indico.ContributionData, error) {
	contributions, err := h.GetContributions(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []indico.ContributionData
	for _, contrib := range contributions {
		if contrib.Session == session {
			filtered = append(filtered, contrib)
		}
	}

	return filtered, nil
}

// GetContributionsByTrack filters contributions by track.
func (h *DataSourceHandler) GetContributionsByTrack(ctx context.Context, track string) ([]indico.ContributionData, error) {
	contributions, err := h.GetContributions(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []indico.ContributionData
	for _, contrib := range contributions {
		if contrib.Track == track {
			filtered = append(filtered, contrib)
		}
	}

	return filtered, nil
}

// GetReviewTracks returns the list of review tracks for the configured data source.
// In API mode this queries the Indico client; in test mode it returns data from a test file if configured.
func (h *DataSourceHandler) GetReviewTracks(ctx context.Context) (*indico.ReviewTracks, error) {
	if h.isTestMode {
		// No test fixture currently configured for review tracks; return empty set to keep behavior predictable
		return &indico.ReviewTracks{Tracks: []indico.ReviewTrack{}}, nil
	}

	if h.client == nil {
		return nil, fmt.Errorf("indico client not initialized")
	}

	return h.client.GetReviewTracks(ctx)
}

// GetMyReviewForAbstract fetches the current user's review for a single abstract
// by database ID. It scrapes the abstract display page and populates
// Rating.QuestionDetails from the handler's question map.
// Returns nil (no error) when the abstract has not been reviewed yet.
func (h *DataSourceHandler) GetMyReviewForAbstract(ctx context.Context, abstractID int) (*indico.Review, error) {
	if h.isTestMode {
		return nil, fmt.Errorf("cannot fetch live review in test mode")
	}
	return h.scrapeMyReview(ctx, abstractID)
}

// GetAssignedReviewCount returns the number of unique abstracts assigned to the current user
// across all review tracks. Returns 0 if none or if review track information is not available.
func (h *DataSourceHandler) GetAssignedReviewCount(ctx context.Context) (int, error) {
	if h.isTestMode {
		// In test mode no review-assignment fixtures are available; return 0 to keep behavior predictable
		return 0, nil
	}

	if h.client == nil {
		return 0, fmt.Errorf("indico client not initialized")
	}

	ra := getReviewAssignments(h, ctx)
	if ra == nil {
		return 0, nil
	}

	return len(ra.isMyReview), nil
}

// GetReviewAbstractIDs returns the list of abstract IDs (friendly_id) for a given review track ID.
// In API mode this queries the Indico client; in test mode it returns an empty list.
func (h *DataSourceHandler) GetReviewAbstractIDs(ctx context.Context, reviewTrackID int) ([]int, error) {
	if h.isTestMode {
		return []int{}, nil
	}

	if h.client == nil {
		return nil, fmt.Errorf("indico client not initialized")
	}

	return h.client.GetReviewAbstractIDs(ctx, reviewTrackID)
}

// RefreshCache invalidates and refreshes a specific cache entry
func (h *DataSourceHandler) RefreshCache(ctx context.Context, key string) error {
	if h.cache == nil {
		return fmt.Errorf("cache not initialized")
	}

	// Invalidate the cache entry
	h.cache.Delete(key)

	// Immediately fetch fresh data based on key
	switch key {
	case "event_info":
		event, err := h.getInfoFromAPI(ctx)
		if err != nil {
			return fmt.Errorf("failed to refresh event info: %w", err)
		}
		h.cache.Set(key, event)
	case "abstracts":
		// Route through GetAbstracts so that the --abstracts-file override is
		// respected; it already handles file, test, and API modes.
		abstracts, err := h.GetAbstracts(ctx)
		if err != nil {
			return fmt.Errorf("failed to refresh abstracts: %w", err)
		}
		h.cache.Set(key, abstracts)
	case "contributions":
		contribs, err := h.getContributionsFromAPI(ctx)
		if err != nil {
			return fmt.Errorf("failed to refresh contributions: %w", err)
		}
		h.cache.Set(key, contribs)
	default:
		return fmt.Errorf("unknown cache key: %s", key)
	}

	log.Printf("Cache refreshed for key: %s", key)
	return nil
}

// DeleteCacheEntry removes a specific entry from cache
func (h *DataSourceHandler) DeleteCacheEntry(key string) error {
	if h.cache == nil {
		return fmt.Errorf("cache not initialized")
	}

	h.cache.Delete(key)
	log.Printf("Cache entry deleted: %s", key)
	return nil
}

// ClearCache removes all entries from cache and deletes the cache file
func (h *DataSourceHandler) ClearCache() error {
	if h.cache == nil {
		return fmt.Errorf("cache not initialized")
	}

	// Clear in-memory cache
	h.cache.Clear()

	// Delete the cache file
	if err := h.cache.DeleteCacheFile(); err != nil {
		log.Printf("Warning: failed to delete cache file: %v", err)
		// Don't fail - the in-memory cache is cleared
	}

	log.Printf("Cache cleared and file removed")
	return nil
}

// GetCacheStats returns cache statistics
func (h *DataSourceHandler) GetCacheStats() map[string]interface{} {
	if h.cache == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}
	stats := h.cache.GetStats()
	stats["enabled"] = true
	return stats
}

// GetCacheKeys returns all available cache keys
func (h *DataSourceHandler) GetCacheKeys() []string {
	if h.cache == nil {
		return []string{}
	}
	return h.cache.Keys()
}

// IsTestMode returns true if the data source is in test mode (local files)
func (h *DataSourceHandler) IsTestMode() bool {
	return h.isTestMode
}

// IsAbstractsFileMode returns true when abstract data is served from the
// --abstracts-file override rather than from the Indico API or test fixtures.
// In this mode the refresh button should be hidden because there is no live
// API to refresh from.
func (h *DataSourceHandler) IsAbstractsFileMode() bool {
	return h.abstractsFile != ""
}

// GetDataSourceName returns the name of the data source
func (h *DataSourceHandler) GetDataSourceName() string {
	if h.config == nil {
		return ""
	}
	return h.config.Name
}

// GetCacheEntries returns all cache entries with metadata grouped by data source
func (h *DataSourceHandler) GetCacheEntries() map[string][]*cache.CacheEntry {
	if h.cache == nil {
		return make(map[string][]*cache.CacheEntry)
	}
	return h.cache.GetAllEntriesWithMetadata()
}

// GetCacheEntryMetadata retrieves metadata for a specific cache entry
func (h *DataSourceHandler) GetCacheEntryMetadata(key string) (*cache.CacheEntry, bool) {
	if h.cache == nil {
		return nil, false
	}
	return h.cache.GetWithMetadata(key)
}

// SetCacheOnExpiry allows higher-level code (e.g., App) to register a callback when cache entries expire.
// Note: This does NOT delete the entry, it only notifies about expiry.
func (h *DataSourceHandler) SetCacheOnExpiry(cb func(fullKey string)) {
	if h.cache == nil {
		return
	}
	h.cache.SetOnExpiry(cb)
}

// SetCacheOnEvict allows higher-level code to register a callback when cache entries are evicted due to size.
func (h *DataSourceHandler) SetCacheOnEvict(cb func(fullKey string)) {
	if h.cache == nil {
		return
	}
	h.cache.SetOnEvict(cb)
}

// UpdateCacheTTL updates the TTL for the cache and recalculates ExpiresAt for all existing entries.
// This is used when the cache configuration is updated without recreating the handler.
func (h *DataSourceHandler) UpdateCacheTTL(newTTL time.Duration) {
	if h.cache == nil {
		return
	}
	h.cache.UpdateTTL(newTTL)
}

// End of DataSourceHandler methods
