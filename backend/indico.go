package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"time"
)

// HTTPClient is the subset of http.Client used so the client can be mocked in tests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// IndicoClient performs REST calls to the backend API.
type IndicoClient struct {
	BaseURL  string
	EventID  int
	APIToken string
	Client   HTTPClient
	Timeout  time.Duration
}

// NewIndicoClient constructs a client with a sensible default http.Client.
func NewIndicoClient(baseURL string, eventID int, apiToken string) *IndicoClient {
	return &IndicoClient{
		BaseURL:  stringsTrimRightSlash(baseURL),
		EventID:  eventID,
		APIToken: apiToken,
		Client:   &http.Client{Timeout: 10 * time.Second},
		Timeout:  10 * time.Second,
	}
}

// GetEventInfo retrieves the raw JSON export for the configured event, decodes
// it and returns an Event populated from the first element of the "results"
// key. If detail is non-empty, the query parameter `detail=<value>` will be sent.
func (c *IndicoClient) GetEventInfo(ctx context.Context, detail string) (*Event, error) {
	// Build path and query.
	path := fmt.Sprintf("/export/event/%d.json", c.EventID)
	q := url.Values{}
	if detail != "" {
		q.Set("detail", detail)
	}

	// Decode the whole response into a generic map. The doGet helper will
	// stream-decode the JSON body into this map for us.
	var resp map[string]any
	if err := c.doGet(ctx, path, q, &resp); err != nil {
		return nil, err
	}

	// Extract "results" from the response map. We expect an array of objects
	// (type []any). Take the first element (a map) as the source of event fields.
	v, ok := resp["results"]
	if !ok || v == nil {
		return nil, fmt.Errorf("missing results in response")
	}

	arr, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("results is not an array, got %T", v)
	}
	if len(arr) == 0 {
		return nil, fmt.Errorf("no results in response")
	}
	firstMap, ok := arr[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("first result element is not an object: %T", arr[0])
	}

	// helper to fetch string values from the map
	get := func(key string) string {
		if x, ok := firstMap[key]; ok && x != nil {
			switch s := x.(type) {
			case string:
				return s
			default:
				return fmt.Sprintf("%v", s)
			}
		}
		return ""
	}

	desc := get("description")
	// Unescape any HTML entities so Description contains original HTML tags.
	desc = html.UnescapeString(desc)

	ev := &Event{
		ID:          get("id"),
		Title:       get("title"),
		Description: desc,
		Location:    get("location"),
		Address:     get("address"),
		Category:    get("category"),
	}

	// parse startDate and endDate which may be either strings or maps
	if raw, ok := firstMap["startDate"]; ok && raw != nil {
		if t, err := parseDateField(raw); err == nil {
			ev.StartDate = t
		}
	}
	if raw, ok := firstMap["endDate"]; ok && raw != nil {
		if t, err := parseDateField(raw); err == nil {
			ev.EndDate = t
		}
	}

	return ev, nil
}

func (c *IndicoClient) doGet(ctx context.Context, path string, query url.Values, out interface{}) error {
	// prepare URL
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}
	u.Path = joinPaths(u.Path, path)
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	// context with timeout
	ctxReq, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxReq, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if c.APIToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIToken)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024))
		return fmt.Errorf("api error: status %d: %s", resp.StatusCode, string(b))
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode error: %w", err)
	}
	return nil
}

// Event holds the essential information about an Indico event.
type Event struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"startDate,omitempty"`
	EndDate     time.Time `json:"endDate,omitempty"`
	Location    string    `json:"location,omitempty"`
	Address     string    `json:"address,omitempty"`
	Category    string    `json:"category,omitempty"`
}

// helper: joinPaths ensures no duplicate slashes.
func joinPaths(base, add string) string {
	if len(base) == 0 {
		return add
	}
	if base[len(base)-1] == '/' && len(add) > 0 && add[0] == '/' {
		return base + add[1:]
	}
	if base[len(base)-1] != '/' && (len(add) == 0 || add[0] != '/') {
		return base + "/" + add
	}
	return base + add
}

// helper: trim trailing slash from base URL
func stringsTrimRightSlash(s string) string {
	for len(s) > 1 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

// parseDateField accepts either a string date or a map with date/time/tz and returns time.Time
func parseDateField(v any) (time.Time, error) {
	switch t := v.(type) {
	case map[string]any:
		// expected keys: date, time, tz
		dateStr := ""
		timeStr := ""
		tzStr := ""
		if d, ok := t["date"]; ok && d != nil {
			dateStr = fmt.Sprintf("%v", d)
		}
		if tm, ok := t["time"]; ok && tm != nil {
			timeStr = fmt.Sprintf("%v", tm)
		}
		if tz, ok := t["tz"]; ok && tz != nil {
			tzStr = fmt.Sprintf("%v", tz)
		}
		if dateStr == "" {
			return time.Time{}, fmt.Errorf("missing date field")
		}
		loc := time.UTC
		if tzStr != "" {
			if l, err := time.LoadLocation(tzStr); err == nil {
				loc = l
			}
		}
		if timeStr == "" {
			// parse date only
			if tt, err := time.ParseInLocation("2006-01-02", dateStr, loc); err == nil {
				return tt, nil
			} else {
				return time.Time{}, err
			}
		}
		// combine
		combined := dateStr + " " + timeStr
		// primary layout with seconds
		if tt, err := time.ParseInLocation("2006-01-02 15:04:05", combined, loc); err == nil {
			return tt, nil
		}
		// try without seconds
		if tt, err := time.ParseInLocation("2006-01-02 15:04", combined, loc); err == nil {
			return tt, nil
		}
		return time.Time{}, fmt.Errorf("unrecognized combined datetime: %s", combined)
	default:
		return time.Time{}, fmt.Errorf("unsupported date field type: %T", v)
	}
}
