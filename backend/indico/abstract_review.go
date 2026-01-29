package indico

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	xhtml "golang.org/x/net/html"
)

// ReviewTrack represents a single review track with a display name, link and numeric id.
type ReviewTrack struct {
	Name          string `json:"name"`
	Link          string `json:"link"`
	TrackID       int    `json:"track_id"`
	AbstractCount int    `json:"abstract_count"`
}

// ReviewTracks is a container for multiple ReviewTrack entries.
type ReviewTracks struct {
	Tracks []ReviewTrack `json:"tracks"`
}

// GetReviewTracks fetches the review statistics HTML page and extracts the
// list of review tracks. It returns a ReviewTracks struct containing each
// track's display name and relative link (href) and numeric track id.
// this method returns the list of the assigned review tracks per user.
func (c *IndicoClient) GetReviewTracks(ctx context.Context) (*ReviewTracks, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/event/%d/abstracts/reviewing/statistics", c.EventID)
	u.Path = joinPaths(u.Path, path)

	ctxReq, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxReq, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Accept HTML
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	if c.APIToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIToken)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
		return nil, fmt.Errorf("api error: status %d: %s", resp.StatusCode, string(b))
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	doc, err := xhtml.Parse(strings.NewReader(string(b)))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	// helper: get attribute value by name (case-insensitive)
	getAttr := func(attrs []xhtml.Attribute, name string) string {
		for _, a := range attrs {
			if strings.EqualFold(a.Key, name) {
				return a.Val
			}
		}
		return ""
	}

	// findNodesByClass searches the subtree rooted at n for elements with the
	// optional tag (empty means any tag) and a class token matching classToken.
	findNodesByClass := func(n *xhtml.Node, tag, classToken string) []*xhtml.Node {
		var res []*xhtml.Node
		var walk func(*xhtml.Node)
		walk = func(cur *xhtml.Node) {
			if cur.Type == xhtml.ElementNode {
				if tag == "" || strings.EqualFold(cur.Data, tag) {
					cls := getAttr(cur.Attr, "class")
					if cls != "" {
						for _, tok := range strings.Fields(cls) {
							if tok == classToken {
								res = append(res, cur)
								break
							}
						}
					}
				}
			}
			for c := cur.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
		walk(n)
		return res
	}

	// textContent concatenates all text node descendants of n, normalized.
	textContent := func(n *xhtml.Node) string {
		var bld strings.Builder
		var walk func(*xhtml.Node)
		walk = func(cur *xhtml.Node) {
			if cur.Type == xhtml.TextNode {
				bld.WriteString(cur.Data)
				bld.WriteByte(' ')
			}
			for c := cur.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
		walk(n)
		return strings.Join(strings.Fields(bld.String()), " ")
	}

	var tracks []ReviewTrack

	// Find the track-review-list containers
	lists := findNodesByClass(doc, "div", "track-review-list")
	for _, list := range lists {
		// For each title node inside the list
		titles := findNodesByClass(list, "div", "title")
		for _, t := range titles {
			// find first <a> descendant
			var aNode *xhtml.Node
			var findA func(*xhtml.Node)
			findA = func(cur *xhtml.Node) {
				if aNode != nil {
					return
				}
				if cur.Type == xhtml.ElementNode && strings.EqualFold(cur.Data, "a") {
					aNode = cur
					return
				}
				for c := cur.FirstChild; c != nil; c = c.NextSibling {
					findA(c)
				}
			}
			findA(t)

			var name string
			var link string
			if aNode != nil {
				link = getAttr(aNode.Attr, "href")
				name = textContent(aNode)
			} else {
				name = textContent(t)
			}
			// prefer relative URL: if href is absolute but same host as base, use requestURI (path + query)
			if link != "" {
				if parsed, err := url.Parse(link); err == nil {
					if parsed.IsAbs() {
						// same host -> convert to relative
						if strings.EqualFold(parsed.Host, u.Host) {
							link = parsed.RequestURI()
						}
					}
				}
			}
			if name == "" {
				continue
			}
			// extract numeric id from link path (last path segment)
			trackID := 0
			if link != "" {
				if parsed, err := url.Parse(link); err == nil {
					seg := strings.Trim(parsed.Path, "/")
					if seg != "" {
						parts := strings.Split(seg, "/")
						last := parts[len(parts)-1]
						if id, err := strconv.Atoi(last); err == nil {
							trackID = id
						}
					}
				}
			}
			tracks = append(tracks, ReviewTrack{Name: name, Link: link, TrackID: trackID})
		}
	}

	// Populate AbstractCount for each track concurrently to avoid N+1 query problem
	var wg sync.WaitGroup
	for i := range tracks {
		if tracks[i].TrackID > 0 {
			wg.Add(1)
			go func(index int, trackID int) {
				defer wg.Done()
				ids, err := c.GetReviewAbstractIDs(ctx, trackID)
				if err != nil {
					// Log error but continue - don't fail the whole request
					// Set count to 0 on error
					tracks[index].AbstractCount = 0
				} else {
					tracks[index].AbstractCount = len(ids)
				}
			}(i, tracks[i].TrackID)
		}
	}
	wg.Wait()

	return &ReviewTracks{Tracks: tracks}, nil
}

// GetReviewAbstractIDs fetches the review-track page for the given reviewTrackID
// and returns the list of abstract IDs (friendly_id) found in table rows with class
// "abstract-row" which include a `data-friendly-id` attribute.
// this method returns the list of the abstracts under an assigned review track per user.
func (c *IndicoClient) GetReviewAbstractIDs(ctx context.Context, reviewTrackID int) ([]int, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/event/%d/abstracts/reviewing/%d", c.EventID, reviewTrackID)
	u.Path = joinPaths(u.Path, path)

	ctxReq, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxReq, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Accept HTML
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	if c.APIToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIToken)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
		return nil, fmt.Errorf("api error: status %d: %s", resp.StatusCode, string(b))
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	doc, err := xhtml.Parse(strings.NewReader(string(b)))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	// helper: get attribute value by name (case-insensitive)
	getAttr := func(attrs []xhtml.Attribute, name string) string {
		for _, a := range attrs {
			if strings.EqualFold(a.Key, name) {
				return a.Val
			}
		}
		return ""
	}

	var ids []int
	var walk func(*xhtml.Node)
	walk = func(n *xhtml.Node) {
		if n.Type == xhtml.ElementNode && strings.EqualFold(n.Data, "tr") {
			// check class contains "abstract-row"
			cls := getAttr(n.Attr, "class")
			if cls != "" {
				for _, tok := range strings.Fields(cls) {
					if tok == "abstract-row" {
						idStr := getAttr(n.Attr, "data-friendly-id")
						if idStr != "" {
							if id, err := strconv.Atoi(idStr); err == nil {
								ids = append(ids, id)
							}
						}
						break
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	return ids, nil
}

// AggRatings aggregates ratings from a single review by question ID.
func (r *Review) AggRatings() map[int]float64 {
	agg := make(map[int]float64)
	for _, rating := range r.Ratings {
		value := convertRatingValue(rating.Value)
		agg[rating.Question] = value
	}
	return agg
}

// AggregateAllRatings aggregates ratings across all reviews for an abstract.
// Returns a map of question ID to total aggregated value.
func (a *AbstractData) AggregateAllRatings() map[int]float64 {
	agg := make(map[int]float64)
	for _, review := range a.Reviews {
		reviewAgg := review.AggRatings()
		for qID, value := range reviewAgg {
			agg[qID] += value
		}
	}
	return agg
}

// GetAggregatedRatingByTitle gets the aggregated rating for a question by its title (case-insensitive).
// Returns 0 if the question is not found or has no ratings.
func (a *AbstractData) GetAggregatedRatingByTitle(questionTitle string) float64 {
	agg := a.AggregateAllRatings()

	// Find question ID by title
	for _, review := range a.Reviews {
		for _, rating := range review.Ratings {
			if rating.QuestionDetails != nil {
				if equalsCaseInsensitive(rating.QuestionDetails.Title, questionTitle) {
					if val, ok := agg[rating.Question]; ok {
						return val
					}
				}
			}
		}
	}
	return 0
}

// convertRatingValue converts a rating value to float64.
// Handles int, float64, bool, and string types.
func convertRatingValue(value interface{}) float64 {
	switch v := value.(type) {
	case int:
		return float64(v)
	case float64:
		return v
	case bool:
		if v {
			return 1.0
		}
		return 0.0
	case string:
		// Handle string representations of boolean
		lower := strings.ToLower(v)
		if lower == "true" || lower == "yes" || lower == "1" {
			return 1.0
		}
		return 0.0
	default:
		return 0.0
	}
}

// Helper function for case-insensitive comparison
func equalsCaseInsensitive(a, b string) bool {
	return strings.ToLower(a) == strings.ToLower(b)
}
