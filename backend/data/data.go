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
}

// NewDataSourceHandler creates a new data source handler with optional cache config
// and a list of APITokenEntry to resolve api tokens referenced by name.
func NewDataSourceHandler(ds *config.DataSource, cacheConfig *config.CacheConfig, tokens []config.APITokenEntry) (*DataSourceHandler, error) {
	handler := &DataSourceHandler{
		config: ds,
	}

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
	if h.cache != nil {
		return h.cache.Shutdown(ctx)
	}
	return nil
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

	// Build question lookup map
	questionMap := make(map[int]*indico.QuestionData)
	for i := range response.Questions {
		q := &response.Questions[i]
		questionMap[q.ID] = q
	}

	// Expand question details in reviews
	for i := range response.Abstracts {
		for j := range response.Abstracts[i].Reviews {
			for k := range response.Abstracts[i].Reviews[j].Ratings {
				rating := &response.Abstracts[i].Reviews[j].Ratings[k]
				if q, ok := questionMap[rating.Question]; ok {
					rating.QuestionDetails = q
				}
			}
		}
		// Compute aggregated ratings for frontend
		response.Abstracts[i].FirstPriority = response.Abstracts[i].GetAggregatedRatingByTitle("First priority")
		response.Abstracts[i].SecondPriority = response.Abstracts[i].GetAggregatedRatingByTitle("Second priority")
	}

	return response.Abstracts, nil
}

// getAbstractsFromAPI fetches abstracts from the Indico API.
// This is a placeholder for future implementation that would fetch from the live API.
func (h *DataSourceHandler) getAbstractsFromAPI(ctx context.Context) ([]indico.AbstractData, error) {
	if h.client == nil {
		return nil, fmt.Errorf("indico client not initialized")
	}

	log.Printf("Reading abstract data from Indico API\n")

	// Fetch the abstracts list page to get all IDs and cache CSRF token internally
	ids, err := h.client.GetAbstractIDsAndCSRFFromList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get abstract IDs: %w", err)
	}

	if len(ids) == 0 {
		return []indico.AbstractData{}, nil
	}

	// Fetch the abstracts data
	rawData, err := h.client.FetchAbstractsData(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch abstracts data: %w", err)
	}

	// Parse the response
	abstractsIface, ok := rawData["abstracts"]
	if !ok {
		return nil, fmt.Errorf("no abstracts field in response")
	}

	// Also get questions if available
	questionsIface := rawData["questions"]

	// Convert to JSON and back to properly deserialize
	jsonData, err := json.Marshal(map[string]any{
		"abstracts": abstractsIface,
		"questions": questionsIface,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal abstracts: %w", err)
	}

	var response indico.AbstractsResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal abstracts: %w", err)
	}

	// Build question lookup map
	questionMap := make(map[int]*indico.QuestionData)
	for i := range response.Questions {
		q := &response.Questions[i]
		questionMap[q.ID] = q
	}

	// Expand question details in reviews
	for i := range response.Abstracts {
		for j := range response.Abstracts[i].Reviews {
			for k := range response.Abstracts[i].Reviews[j].Ratings {
				rating := &response.Abstracts[i].Reviews[j].Ratings[k]
				if q, ok := questionMap[rating.Question]; ok {
					rating.QuestionDetails = q
				}
			}
		}
		// Compute aggregated ratings for frontend
		response.Abstracts[i].FirstPriority = response.Abstracts[i].GetAggregatedRatingByTitle("First priority")
		response.Abstracts[i].SecondPriority = response.Abstracts[i].GetAggregatedRatingByTitle("Second priority")
	}

	myReviewIDsSet := getReviewIDsSet(h, ctx)
	if myReviewIDsSet != nil {
		// Mark abstracts that are in my review list
		for i := range response.Abstracts {
			response.Abstracts[i].IsMyReview = myReviewIDsSet[response.Abstracts[i].FriendlyID]
			response.Abstracts[i].ReviewURL = fmt.Sprintf("%s/event/%d/abstracts/%d", h.client.BaseURL, h.client.EventID, response.Abstracts[i].ID)
		}
	}

	// Set IndicoURL for each abstract
	for i := range response.Abstracts {
		response.Abstracts[i].IndicoURL = fmt.Sprintf("%s/event/%d/abstracts/%d", h.client.BaseURL, h.client.EventID, response.Abstracts[i].ID)
	}

	// Ensure reviewer avatar URLs are absolute by prefixing the client's BaseURL
	for i := range response.Abstracts {
		for j := range response.Abstracts[i].Reviews {
			av := response.Abstracts[i].Reviews[j].User.AvatarURL
			if av != "" && !strings.HasPrefix(av, "http") {
				response.Abstracts[i].Reviews[j].User.AvatarURL = h.client.BaseURL + av
			}
		}
		// Submitter avatar
		if response.Abstracts[i].Submitter != nil {
			sav := response.Abstracts[i].Submitter.AvatarURL
			if sav != "" && !strings.HasPrefix(sav, "http") {
				response.Abstracts[i].Submitter.AvatarURL = h.client.BaseURL + sav
			}
		}
		// Judge avatar
		if response.Abstracts[i].Judge != nil {
			jav := response.Abstracts[i].Judge.AvatarURL
			if jav != "" && !strings.HasPrefix(jav, "http") {
				response.Abstracts[i].Judge.AvatarURL = h.client.BaseURL + jav
			}
		}
	}

	return response.Abstracts, nil
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

// getReviewIDsSet returns a map of abstract friendly IDs that are on my review list or not.
func getReviewIDsSet(h *DataSourceHandler, ctx context.Context) map[int]bool {

	// Fetch review tracks and mark abstracts that are on my review list
	reviewTracks, err := h.client.GetReviewTracks(ctx)
	if err != nil {
		log.Printf("Warning: failed to fetch review tracks for refresh: %v", err)
		return nil
	}

	if reviewTracks != nil && len(reviewTracks.Tracks) > 0 {
		myReviewIDsSet := make(map[int]bool)
		for _, track := range reviewTracks.Tracks {
			if track.Link == "" {
				continue
			}
			abstractIDs, err := h.client.GetReviewAbstractIDs(ctx, track.TrackID)
			if err != nil {
				log.Printf("Warning: failed to get abstract IDs for track %d: %v", track.TrackID, err)
				continue
			}
			for _, aid := range abstractIDs {
				myReviewIDsSet[aid] = true
			}
		}
		return myReviewIDsSet
	}
	return nil
}

// RefreshAbstractByID fetches fresh data for a single abstract from the API.
// This bypasses the cache and always fetches from the live API.
func (h *DataSourceHandler) RefreshAbstractByID(ctx context.Context, id int) (*indico.AbstractData, error) {
	if h.client == nil {
		return nil, fmt.Errorf("indico client not initialized (test mode or no API access)")
	}

	log.Printf("Refreshing abstract %d from Indico API\n", id)

	// Fetch the single abstract data
	rawData, err := h.client.FetchAbstractsData(ctx, []string{fmt.Sprintf("%d", id)})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch abstract %d: %w", id, err)
	}

	// Parse the response
	abstractsIface, ok := rawData["abstracts"]
	if !ok {
		return nil, fmt.Errorf("no abstracts field in response")
	}

	// Also get questions if available
	questionsIface := rawData["questions"]

	// Convert to JSON and back to properly deserialize
	jsonData, err := json.Marshal(map[string]any{
		"abstracts": abstractsIface,
		"questions": questionsIface,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal abstract: %w", err)
	}

	var response indico.AbstractsResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal abstract: %w", err)
	}

	if len(response.Abstracts) == 0 {
		return nil, fmt.Errorf("abstract with ID %d not found in API response", id)
	}

	// Build question lookup map
	questionMap := make(map[int]*indico.QuestionData)
	for i := range response.Questions {
		q := &response.Questions[i]
		questionMap[q.ID] = q
	}

	// Expand question details in reviews for the single abstract
	abstract := &response.Abstracts[0]
	for j := range abstract.Reviews {
		for k := range abstract.Reviews[j].Ratings {
			rating := &abstract.Reviews[j].Ratings[k]
			if q, ok := questionMap[rating.Question]; ok {
				rating.QuestionDetails = q
			}
		}
	}

	// Compute aggregated ratings for frontend
	abstract.FirstPriority = abstract.GetAggregatedRatingByTitle("First priority")
	abstract.SecondPriority = abstract.GetAggregatedRatingByTitle("Second priority")

	myReviewIDsSet := getReviewIDsSet(h, ctx)
	if myReviewIDsSet != nil {
		abstract.IsMyReview = myReviewIDsSet[abstract.FriendlyID]
		abstract.ReviewURL = fmt.Sprintf("%s/event/%d/abstracts/%d", h.client.BaseURL, h.client.EventID, abstract.ID)
	}

	// Ensure avatar URLs are absolute for single-abstract refresh as well
	for j := range abstract.Reviews {
		av := abstract.Reviews[j].User.AvatarURL
		if av != "" && !strings.HasPrefix(av, "http") {
			abstract.Reviews[j].User.AvatarURL = h.client.BaseURL + av
		}
	}
	if abstract.Submitter != nil {
		sav := abstract.Submitter.AvatarURL
		if sav != "" && !strings.HasPrefix(sav, "http") {
			abstract.Submitter.AvatarURL = h.client.BaseURL + sav
		}
	}
	if abstract.Judge != nil {
		jav := abstract.Judge.AvatarURL
		if jav != "" && !strings.HasPrefix(jav, "http") {
			abstract.Judge.AvatarURL = h.client.BaseURL + jav
		}
	}

	// Update the cache if available
	if h.cache != nil {
		// Get current cached abstracts
		if cached, found := h.cache.Get("abstracts"); found {
			var abstracts []indico.AbstractData

			// Try direct type assertion first
			if cachedAbstracts, ok := cached.([]indico.AbstractData); ok {
				abstracts = cachedAbstracts
			} else {
				// Try JSON conversion
				jsonData, err := json.Marshal(cached)
				if err == nil {
					_ = json.Unmarshal(jsonData, &abstracts)
				}
			}

			// Update the abstract in the cache
			updated := false
			for i := range abstracts {
				if abstracts[i].ID == id {
					abstracts[i] = *abstract
					updated = true
					break
				}
			}

			// If not found in cache, append it
			if !updated {
				abstracts = append(abstracts, *abstract)
			}

			// Save back to cache
			h.cache.Set("abstracts", abstracts)
			log.Printf("Updated abstract %d in cache\n", id)
		}
	}

	log.Printf("Successfully refreshed abstract %d from API: %v\n", id, abstract)
	return abstract, nil
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
		abstracts, err := h.getAbstractsFromAPI(ctx)
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

// End of DataSourceHandler methods
