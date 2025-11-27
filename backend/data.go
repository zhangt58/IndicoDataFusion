package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// DataSourceHandler provides a high-level interface for accessing event data
// from different sources (Indico API or local test files).
type DataSourceHandler struct {
	config     *DataSource
	client     *IndicoClient
	dataDir    string
	isTestMode bool
}

// NewDataSourceHandler creates a new data source handler from a DataSource configuration.
func NewDataSourceHandler(ds *DataSource) (*DataSourceHandler, error) {
	handler := &DataSourceHandler{
		config: ds,
	}

	if ds.Indico != nil {
		// Initialize Indico client
		client := NewIndicoClient(
			ds.Indico.BaseURL,
			ds.Indico.EventID,
			ds.Indico.APIToken,
		)
		if ds.Indico.Timeout > 0 {
			client.Timeout = time.Duration(ds.Indico.Timeout)
		}
		handler.client = client
		handler.isTestMode = false
	} else if ds.Test != nil {
		// Test mode with local files
		handler.dataDir, _ = filepath.Abs(ds.Test.DataDir)
		handler.isTestMode = true
	} else {
		return nil, fmt.Errorf("data source %s has no valid configuration", ds.Name)
	}

	return handler, nil
}

// NewDataSourceHandlerFromConfig creates a handler from a full Config using the default data source.
func NewDataSourceHandlerFromConfig(cfg *Config) (*DataSourceHandler, error) {
	ds, err := cfg.GetDefaultDataSource()
	if err != nil {
		return nil, fmt.Errorf("failed to get default data source: %w", err)
	}
	return NewDataSourceHandler(ds)
}

// NewDataSourceHandlerFromConfigFile creates a handler by loading a config file.
func NewDataSourceHandlerFromConfigFile(configPath string) (*DataSourceHandler, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return NewDataSourceHandlerFromConfig(cfg)
}

// GetInfo retrieves event information from the configured data source.
func (h *DataSourceHandler) GetInfo(ctx context.Context) (*Event, error) {
	if h.isTestMode {
		return h.getInfoFromFile()
	}
	return h.getInfoFromAPI(ctx)
}

// getInfoFromFile reads event info from a local JSON file (test mode).
func (h *DataSourceHandler) getInfoFromFile() (*Event, error) {
	if h.config.Test == nil {
		return nil, fmt.Errorf("test configuration not available")
	}

	filePath := filepath.Join(h.dataDir, h.config.Test.EventInfo)
	log.Printf("Reading event info from: %v\n", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	var ev Event
	if err := json.Unmarshal(data, &ev); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	return &ev, nil
}

// getInfoFromAPI fetches event info from the Indico API.
func (h *DataSourceHandler) getInfoFromAPI(ctx context.Context) (*Event, error) {
	if h.client == nil {
		return nil, fmt.Errorf("indico client not initialized")
	}
	log.Printf("Reading event info from Indico API\n")
	return h.client.GetEventInfo()
}

// GetAbstracts retrieves abstract data from the configured data source.
func (h *DataSourceHandler) GetAbstracts(ctx context.Context) ([]AbstractData, error) {
	if h.isTestMode {
		return h.getAbstractsFromFile()
	}
	return h.getAbstractsFromAPI(ctx)
}

// getAbstractsFromFile reads abstracts from a local JSON file (test mode).
func (h *DataSourceHandler) getAbstractsFromFile() ([]AbstractData, error) {
	if h.config.Test == nil {
		return nil, fmt.Errorf("test configuration not available")
	}

	filePath := filepath.Join(h.dataDir, h.config.Test.Abstracts)
	log.Printf("Reading abstract data from: %v\n", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	var response AbstractsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	return response.Abstracts, nil
}

// getAbstractsFromAPI fetches abstracts from the Indico API.
// This is a placeholder for future implementation that would fetch from the live API.
func (h *DataSourceHandler) getAbstractsFromAPI(ctx context.Context) ([]AbstractData, error) {
	if h.client == nil {
		return nil, fmt.Errorf("indico client not initialized")
	}

	log.Printf("Reading abstract data from Indico API\n")

	// Fetch the abstracts list page to get IDs and CSRF token
	ids, csrfToken, err := h.client.GetAbstractIDsAndCSRFFromList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get abstract IDs: %w", err)
	}

	if len(ids) == 0 {
		return []AbstractData{}, nil
	}

	// Fetch the abstracts data
	rawData, err := h.client.FetchAbstractsData(ctx, ids, csrfToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch abstracts data: %w", err)
	}

	// Parse the response
	abstractsIface, ok := rawData["abstracts"]
	if !ok {
		return nil, fmt.Errorf("no abstracts field in response")
	}

	// Convert to JSON and back to properly deserialize
	jsonData, err := json.Marshal(map[string]any{"abstracts": abstractsIface})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal abstracts: %w", err)
	}

	var response AbstractsResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal abstracts: %w", err)
	}

	return response.Abstracts, nil
}

// GetContributions retrieves contribution data from the configured data source.
func (h *DataSourceHandler) GetContributions(ctx context.Context) ([]ContributionData, error) {
	if h.isTestMode {
		return h.getContributionsFromFile()
	}
	return h.getContributionsFromAPI(ctx)
}

// getContributionsFromFile reads contributions from a local JSON file (test mode).
func (h *DataSourceHandler) getContributionsFromFile() ([]ContributionData, error) {
	if h.config.Test == nil {
		return nil, fmt.Errorf("test configuration not available")
	}

	filePath := filepath.Join(h.dataDir, h.config.Test.Contribs)
	log.Printf("Reading contribution data from: %v\n", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	var response ContributionsAPIResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	if len(response.Results) == 0 {
		return []ContributionData{}, nil
	}

	return response.Results[0].Contributions, nil
}

// getContributionsFromAPI fetches contributions from the Indico API.
func (h *DataSourceHandler) getContributionsFromAPI(ctx context.Context) ([]ContributionData, error) {
	if h.client == nil {
		return nil, fmt.Errorf("indico client not initialized")
	}

	log.Printf("Reading contribution data from Indico API\n")

	// Construct the API path for contributions
	path := fmt.Sprintf("/export/event/%d.json", h.client.EventID)

	queryValues := url.Values{}
	queryValues.Set("detail", "contributions")

	// Fetch the contribution data
	var response ContributionsAPIResponse
	if err := h.client.doGet(ctx, path, queryValues, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch contributions: %w", err)
	}

	if len(response.Results) == 0 {
		return []ContributionData{}, nil
	}

	return response.Results[0].Contributions, nil
}

// GetContributionByID retrieves a specific contribution by ID.
func (h *DataSourceHandler) GetContributionByID(ctx context.Context, id string) (*ContributionData, error) {
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
func (h *DataSourceHandler) GetAbstractByID(ctx context.Context, id int) (*AbstractData, error) {
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

// GetAbstractsByState filters abstracts by their state.
func (h *DataSourceHandler) GetAbstractsByState(ctx context.Context, state string) ([]AbstractData, error) {
	abstracts, err := h.GetAbstracts(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []AbstractData
	for _, abstract := range abstracts {
		if abstract.State == state {
			filtered = append(filtered, abstract)
		}
	}

	return filtered, nil
}

// GetContributionsBySession filters contributions by session.
func (h *DataSourceHandler) GetContributionsBySession(ctx context.Context, session string) ([]ContributionData, error) {
	contributions, err := h.GetContributions(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []ContributionData
	for _, contrib := range contributions {
		if contrib.Session == session {
			filtered = append(filtered, contrib)
		}
	}

	return filtered, nil
}

// GetContributionsByTrack filters contributions by track.
func (h *DataSourceHandler) GetContributionsByTrack(ctx context.Context, track string) ([]ContributionData, error) {
	contributions, err := h.GetContributions(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []ContributionData
	for _, contrib := range contributions {
		if contrib.Track == track {
			filtered = append(filtered, contrib)
		}
	}

	return filtered, nil
}
