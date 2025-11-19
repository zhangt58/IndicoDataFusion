package backend

import (
	"context"
	"encoding/json"
	"fmt"
	stdhtml "html"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	xhtml "golang.org/x/net/html"
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
	desc = stdhtml.UnescapeString(desc)

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

// ExtractAbstractIDsAndCSRFFromFile parses an HTML file and returns two values:
//   - a slice of `value` attributes for any <input> element found within a
//     <tr> element whose class contains the token "abstract-row";
//   - the CSRF token (if found) from a <meta name="csrf_token"|"csrf-token"|"csrf" content="..."> tag.
//
// The function uses golang.org/x/net/html to robustly parse HTML and is defensive
// about missing elements.
func (c *IndicoClient) ExtractAbstractIDsAndCSRFFromFile(htmlPath string) ([]string, string, error) {
	f, err := os.Open(htmlPath)
	if err != nil {
		return nil, "", fmt.Errorf("open html file: %w", err)
	}
	defer f.Close()

	doc, err := xhtml.Parse(f)
	if err != nil {
		return nil, "", fmt.Errorf("parse html: %w", err)
	}

	var ids []string
	var csrf string

	// helper: check if an attribute list contains a key=value pair (case-insensitive key)
	getAttr := func(attrs []xhtml.Attribute, name string) string {
		for _, a := range attrs {
			if strings.EqualFold(a.Key, name) {
				return a.Val
			}
		}
		return ""
	}

	// find meta csrf token anywhere in the document
	var walk func(*xhtml.Node)
	walk = func(n *xhtml.Node) {
		if n.Type == xhtml.ElementNode {
			if strings.EqualFold(n.Data, "meta") && csrf == "" {
				name := strings.ToLower(getAttr(n.Attr, "name"))
				if name == "csrf_token" || name == "csrf-token" || name == "csrf" {
					csrf = getAttr(n.Attr, "content")
				}
			}

			// tr with class token
			if strings.EqualFold(n.Data, "tr") {
				cls := getAttr(n.Attr, "class")
				if cls != "" {
					// tokenized by whitespace
					for _, tok := range strings.Fields(cls) {
						if tok == "abstract-row" {
							// find any <input> descendants and collect value attributes
							var findInputs func(*xhtml.Node)
							findInputs = func(m *xhtml.Node) {
								if m.Type == xhtml.ElementNode && strings.EqualFold(m.Data, "input") {
									val := getAttr(m.Attr, "value")
									if val != "" {
										ids = append(ids, val)
									}
								}
								for c := m.FirstChild; c != nil; c = c.NextSibling {
									findInputs(c)
								}
							}
							// run on this tr node
							findInputs(n)
							break
						}
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)

	return ids, csrf, nil
}

// parseAbstractIDsAndCSRFFromRoot walks an HTML node tree and returns the
// list of input `value` attributes found under any <tr class="abstract-row"> and
// the first CSRF token found in a <meta name=... content=...> tag.
func parseAbstractIDsAndCSRFFromRoot(doc *xhtml.Node) ([]string, string) {
	var ids []string
	var csrf string

	getAttr := func(attrs []xhtml.Attribute, name string) string {
		for _, a := range attrs {
			if strings.EqualFold(a.Key, name) {
				return a.Val
			}
		}
		return ""
	}

	var walk func(*xhtml.Node)
	walk = func(n *xhtml.Node) {
		if n.Type == xhtml.ElementNode {
			if strings.EqualFold(n.Data, "meta") && csrf == "" {
				name := strings.ToLower(getAttr(n.Attr, "name"))
				if name == "csrf_token" || name == "csrf-token" || name == "csrf" {
					csrf = getAttr(n.Attr, "content")
				}
			}

			if strings.EqualFold(n.Data, "tr") {
				cls := getAttr(n.Attr, "class")
				if cls != "" {
					for _, tok := range strings.Fields(cls) {
						if tok == "abstract-row" {
							var findInputs func(*xhtml.Node)
							findInputs = func(m *xhtml.Node) {
								if m.Type == xhtml.ElementNode && strings.EqualFold(m.Data, "input") {
									val := getAttr(m.Attr, "value")
									if val != "" {
										ids = append(ids, val)
									}
								}
								for c := m.FirstChild; c != nil; c = c.NextSibling {
									findInputs(c)
								}
							}
							findInputs(n)
							break
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)
	return ids, csrf
}

// ExtractAbstractIDsAndCSRFFromHTML parses HTML content provided as a string and
// returns the abstract input values and csrf token (same semantics as the
// file-based version).
func (c *IndicoClient) ExtractAbstractIDsAndCSRFFromHTML(htmlContent string) ([]string, string, error) {
	doc, err := xhtml.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, "", fmt.Errorf("parse html: %w", err)
	}
	ids, csrf := parseAbstractIDsAndCSRFFromRoot(doc)
	return ids, csrf, nil
}

// FetchAbstractsData posts to the Indico manage abstracts endpoint and returns the decoded JSON
// response as a map[string]any. It posts form-encoded data with csrf_token and repeated abstract_id
// form fields. The caller must provide a non-empty csrfToken and at least one id.
func (c *IndicoClient) FetchAbstractsData(ctx context.Context, ids []string, csrfToken string) (map[string]any, error) {
	if csrfToken == "" {
		return nil, fmt.Errorf("empty csrf token")
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no ids provided")
	}

	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/event/%d/manage/abstracts/abstracts.json", c.EventID)
	u.Path = joinPaths(u.Path, path)

	v := url.Values{}
	v.Set("csrf_token", csrfToken)
	for _, id := range ids {
		v.Add("abstract_id", id)
	}

	ctxReq, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxReq, http.MethodPost, u.String(), strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if c.APIToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIToken)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024))
		return nil, fmt.Errorf("api error: status %d: %s", resp.StatusCode, string(b))
	}

	var out map[string]any
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}
	return out, nil
}

// ListAbstracts fetches the HTML page for the abstracts list at
// /event/<event-id>/manage/abstracts/list/.
func (c *IndicoClient) ListAbstracts(ctx context.Context) (string, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("/event/%d/manage/abstracts/list/", c.EventID)
	u.Path = joinPaths(u.Path, path)

	ctxReq, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxReq, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}

	// Accept HTML
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	if c.APIToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIToken)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
		return "", fmt.Errorf("api error: status %d: %s", resp.StatusCode, string(b))
	}

	// Read the HTML body (limit to reasonable size to avoid pathological responses)
	b, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	return string(b), nil
}

// GetAbstractIDsAndCSRFFromList fetches the abstracts list page using the
// provided token and returns the parsed abstract ids and csrf token. It is a
// higher-level helper that combines ListAbstracts + parsing.
func (c *IndicoClient) GetAbstractIDsAndCSRFFromList(ctx context.Context) ([]string, string, error) {
	htmlBody, err := c.ListAbstracts(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("fetch list html: %w", err)
	}
	return c.ExtractAbstractIDsAndCSRFFromHTML(htmlBody)
}
