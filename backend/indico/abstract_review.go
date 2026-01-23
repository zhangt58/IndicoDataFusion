package indico

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	xhtml "golang.org/x/net/html"
)

// ReviewTrack represents a single review track with a display name, link and numeric id.
type ReviewTrack struct {
	Name    string `json:"name"`
	Link    string `json:"link"`
	TrackID int    `json:"track_id"`
}

// ReviewTracks is a container for multiple ReviewTrack entries.
type ReviewTracks struct {
	Tracks []ReviewTrack `json:"tracks"`
}

// GetReviewTracks fetches the review statistics HTML page and extracts the
// list of review tracks. It returns a ReviewTracks struct containing each
// track's display name and relative link (href) and numeric track id.
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

	return &ReviewTracks{Tracks: tracks}, nil
}
