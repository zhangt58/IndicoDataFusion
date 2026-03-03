package indico

import (
	"context"
	"fmt"
	"io"
	"log"
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

	// required if the abstract data is provided through --abstracts-file, otherwise
	// it is set when GetAbstracts is called.
	if _, csrf := parseAbstractIDsAndCSRFFromRoot(doc); csrf != "" {
		c.csrfToken = csrf
	}

	// Extract user ID from body tag's data-user-id attribute
	var findBody func(*xhtml.Node) *xhtml.Node
	findBody = func(n *xhtml.Node) *xhtml.Node {
		if n.Type == xhtml.ElementNode && strings.EqualFold(n.Data, "body") {
			return n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if result := findBody(c); result != nil {
				return result
			}
		}
		return nil
	}

	if bodyNode := findBody(doc); bodyNode != nil {
		if userIDStr := getAttr(bodyNode.Attr, "data-user-id"); userIDStr != "" {
			if userID, err := strconv.Atoi(userIDStr); err == nil {
				c.UserID = userID
			}
		}
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

// ParseReviewFromHTML extracts the current user's review from an already-parsed
// abstract display page (HTML document node). The page is the one served at
// /event/{event-id}/abstracts/{abstract-id} and contains a "proposal-review-{id}"
// div for each review left by the current user.
//
// questionsByTitle maps a lower-cased question title (e.g. "first priority") to its
// numeric question ID so that ratings can be populated with the correct IDs.
// Pass nil or an empty map if question ID resolution is not needed.
//
// abstractsByDBID maps an abstract's database ID (as it appears in the Indico URL,
// e.g. /event/37/abstracts/143/) to a RelatedAbstract that carries the friendly_id
// and title. Used to resolve proposed_related_abstract for mark_as_duplicate/merge.
// Pass nil if this information is not available.
//
// tracksByCode maps a track code (e.g. "MC8.1") to a Track carrying the numeric ID.
// Used to resolve proposed_tracks for change_tracks reviews.
// Pass nil if this information is not available.
//
// contribTypesByName maps a contribution type name (e.g. "Contributed Oral
// Presentation") to its numeric ID.  Used to resolve ProposedContribType for the
// accept action when the badge reads "Proposed to accept as <strong>Name</strong>".
// Pass nil if this information is not available.
//
// The function returns nil, nil when no review block is found on the page.
func ParseReviewFromHTML(doc *xhtml.Node, questionsByTitle map[string]int, abstractsByDBID map[int]*RelatedAbstract, tracksByCode map[string]*Track, contribTypesByName map[string]int) (*Review, error) {
	// ── helpers ────────────────────────────────────────────────────────────────

	getAttr := func(attrs []xhtml.Attribute, name string) string {
		for _, a := range attrs {
			if strings.EqualFold(a.Key, name) {
				return a.Val
			}
		}
		return ""
	}

	hasClassToken := func(attrs []xhtml.Attribute, token string) bool {
		cls := getAttr(attrs, "class")
		for _, t := range strings.Fields(cls) {
			if t == token {
				return true
			}
		}
		return false
	}

	// textContent returns all text-node descendants concatenated and normalised.
	var textContent func(*xhtml.Node) string
	textContent = func(n *xhtml.Node) string {
		var b strings.Builder
		var walk func(*xhtml.Node)
		walk = func(cur *xhtml.Node) {
			if cur.Type == xhtml.TextNode {
				b.WriteString(cur.Data)
				b.WriteByte(' ')
			}
			for c := cur.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
		walk(n)
		return strings.Join(strings.Fields(b.String()), " ")
	}

	// findFirst performs a depth-first search and returns the first node that
	// satisfies the predicate.
	var findFirst func(*xhtml.Node, func(*xhtml.Node) bool) *xhtml.Node
	findFirst = func(n *xhtml.Node, pred func(*xhtml.Node) bool) *xhtml.Node {
		if pred(n) {
			return n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if r := findFirst(c, pred); r != nil {
				return r
			}
		}
		return nil
	}

	// findAll returns all nodes matching the predicate (depth-first).
	var findAll func(*xhtml.Node, func(*xhtml.Node) bool) []*xhtml.Node
	findAll = func(n *xhtml.Node, pred func(*xhtml.Node) bool) []*xhtml.Node {
		var res []*xhtml.Node
		var walk func(*xhtml.Node)
		walk = func(cur *xhtml.Node) {
			if pred(cur) {
				res = append(res, cur)
			}
			for c := cur.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
		walk(n)
		return res
	}

	isElem := func(n *xhtml.Node, tag string) bool {
		return n.Type == xhtml.ElementNode && strings.EqualFold(n.Data, tag)
	}

	// ── locate the proposal-review-{id} div ────────────────────────────────────
	// The div has id="proposal-review-{reviewID}".
	reviewDiv := findFirst(doc, func(n *xhtml.Node) bool {
		if !isElem(n, "div") {
			return false
		}
		id := getAttr(n.Attr, "id")
		return strings.HasPrefix(id, "proposal-review-")
	})
	if reviewDiv == nil {
		// No review found on this page.
		return nil, nil
	}

	// ── extract review ID ──────────────────────────────────────────────────────
	reviewIDStr := strings.TrimPrefix(getAttr(reviewDiv.Attr, "id"), "proposal-review-")
	reviewID, _ := strconv.Atoi(reviewIDStr)

	// ── extract created_dt from <time datetime="..."> ──────────────────────────
	var createdDT string
	timeNode := findFirst(reviewDiv, func(n *xhtml.Node) bool {
		return isElem(n, "time") && getAttr(n.Attr, "datetime") != ""
	})
	if timeNode != nil {
		createdDT = getAttr(timeNode.Attr, "datetime")
	}

	// ── extract reviewer name from <strong> ────────────────────────────────────
	var reviewerFullName string
	strongNode := findFirst(reviewDiv, func(n *xhtml.Node) bool {
		return isElem(n, "strong")
	})
	if strongNode != nil {
		reviewerFullName = textContent(strongNode)
	}

	// ── extract reviewer user ID from avatar img src (/user/{id}/...) ─────────
	var reviewerUserID int
	var reviewerAvatarURL string
	imgNode := findFirst(reviewDiv, func(n *xhtml.Node) bool {
		return isElem(n, "img") && strings.Contains(getAttr(n.Attr, "class"), "profile-picture")
	})
	if imgNode != nil {
		reviewerAvatarURL = getAttr(imgNode.Attr, "src")
		// Path pattern: /user/{id}/picture-...
		parts := strings.Split(strings.Trim(reviewerAvatarURL, "/"), "/")
		if len(parts) >= 2 && parts[0] == "user" {
			reviewerUserID, _ = strconv.Atoi(parts[1])
		}
	}

	// ── extract review-group (track) link and title ────────────────────────────
	var track Track
	reviewGroupDiv := findFirst(reviewDiv, func(n *xhtml.Node) bool {
		return isElem(n, "div") && hasClassToken(n.Attr, "review-group")
	})
	if reviewGroupDiv != nil {
		aNode := findFirst(reviewGroupDiv, func(n *xhtml.Node) bool {
			return isElem(n, "a")
		})
		if aNode != nil {
			fullTitle := getAttr(aNode.Attr, "title") // long title e.g. "MC7 ...: MC7.1 - MC7.1 First Track..."
			href := getAttr(aNode.Attr, "href")       // e.g. /event/37/abstracts/reviewing/88/
			// Extract track ID from href (last numeric path segment)
			if href != "" {
				if pu, err := url.Parse(href); err == nil {
					seg := strings.Trim(pu.Path, "/")
					parts := strings.Split(seg, "/")
					last := parts[len(parts)-1]
					track.ID, _ = strconv.Atoi(last)
				}
			}
			// Parse "GroupName: TrackCode - TrackTitle" or just display text
			// title attr: "MC7 Accelerator Technology Main Systems: MC7.1 - MC7.1 First Track..."
			// We store the full title in track.Title and try to extract code.
			if fullTitle != "" {
				// Try to split on " - " to get "GroupName: Code" and "Title"
				parts := strings.SplitN(fullTitle, " - ", 2)
				if len(parts) == 2 {
					// parts[0] = "MC7 ...: MC7.1", parts[1] = "MC7.1 First Track..."
					colonParts := strings.SplitN(parts[0], ": ", 2)
					if len(colonParts) == 2 {
						track.Code = strings.TrimSpace(colonParts[1])
					}
					track.Title = strings.TrimSpace(parts[1])
				} else {
					track.Title = fullTitle
				}
			} else {
				track.Title = textContent(aNode)
			}
		}
	}

	// ── extract proposed action, contrib type & related abstract ──────────────
	var proposedAction string
	var proposedContribType *ContribType
	var proposedRelatedAbstract *RelatedAbstract
	reviewBadgesDiv := findFirst(reviewDiv, func(n *xhtml.Node) bool {
		return isElem(n, "div") && hasClassToken(n.Attr, "review-badges")
	})
	if reviewBadgesDiv != nil {
		// Look for a <span> inside review-badges that contains the action word.
		// HTML patterns:
		//   accept (no contrib type): Proposed to <span class="bold ...">accept</span>
		//   accept (with contrib):    Proposed to <span class="bold ...">accept</span> as <strong>Contributed Oral Presentation</strong>
		//   reject:           Proposed to <span class="bold ...">reject</span>
		//   change tracks:    Proposed to <span class="bold ...">change tracks</span>
		//   mark_as_duplicate: Proposed as <span class="bold ...">duplicate</span> of <a href="/event/37/abstracts/143/">Title</a>
		//   merge:            Proposed to <span class="bold ...">merge</span> into <a href="/event/37/abstracts/149/">Title</a>
		actionSpan := findFirst(reviewBadgesDiv, func(n *xhtml.Node) bool {
			return isElem(n, "span") && hasClassToken(n.Attr, "bold")
		})
		if actionSpan != nil {
			proposedAction = strings.ToLower(strings.TrimSpace(textContent(actionSpan)))
		}
		switch proposedAction {
		case "duplicate":
			proposedAction = "mark_as_duplicate"
		case "change tracks":
			proposedAction = "change_tracks"
		}

		// For accept: extract the optional contribution type from a <strong>
		// sibling that follows the action span in review-badges.
		// Template emits: … accept</span> as <strong>Name</strong>
		// We scan review-badges direct children for a <strong> element.
		if proposedAction == "accept" {
			strongNode := findFirst(reviewBadgesDiv, func(n *xhtml.Node) bool {
				return isElem(n, "strong")
			})
			if strongNode != nil {
				ctName := strings.TrimSpace(textContent(strongNode))
				if ctName != "" {
					ct := &ContribType{Name: ctName}
					// Resolve numeric ID from the lookup map if available.
					if contribTypesByName != nil {
						if id, ok := contribTypesByName[ctName]; ok {
							ct.ID = id
						}
					}
					proposedContribType = ct
				}
			}
		}

		// For mark_as_duplicate and merge, extract the target abstract from the
		// sibling <a href="/event/.../abstracts/{dbID}/"> that follows the action span.
		if proposedAction == "mark_as_duplicate" || proposedAction == "merge" {
			// Find an <a> in review-badges whose href matches /abstracts/{digits}/
			// (but NOT /abstracts/reviewing/ which is the track link).
			relatedAnchor := findFirst(reviewBadgesDiv, func(n *xhtml.Node) bool {
				if !isElem(n, "a") {
					return false
				}
				href := getAttr(n.Attr, "href")
				// Must contain /abstracts/ but must NOT be a reviewing or review-edit link.
				return strings.Contains(href, "/abstracts/") &&
					!strings.Contains(href, "/reviewing/") &&
					!strings.Contains(href, "/reviews/")
			})
			if relatedAnchor != nil {
				href := getAttr(relatedAnchor.Attr, "href")
				// Extract the trailing numeric segment: /event/37/abstracts/143/
				if pu, err := url.Parse(href); err == nil {
					seg := strings.Trim(pu.Path, "/")
					parts := strings.Split(seg, "/")
					// last segment is the DB id
					dbID, parseErr := strconv.Atoi(parts[len(parts)-1])
					if parseErr == nil {
						// Try to resolve friendly_id + title from the lookup map.
						if abstractsByDBID != nil {
							if ra, ok := abstractsByDBID[dbID]; ok {
								proposedRelatedAbstract = ra
							}
						}
						// Fallback: populate from what the HTML gives us (title text, no friendly_id).
						if proposedRelatedAbstract == nil {
							title := strings.TrimSpace(textContent(relatedAnchor))
							proposedRelatedAbstract = &RelatedAbstract{
								ID:    dbID,
								Title: title,
							}
						}
					}
				}
			}
		}
	}

	// ── extract ratings ────────────────────────────────────────────────────────
	var ratings []Rating
	reviewQuestionsUL := findFirst(reviewDiv, func(n *xhtml.Node) bool {
		return isElem(n, "ul") && hasClassToken(n.Attr, "review-questions")
	})
	if reviewQuestionsUL != nil {
		// Each <li class="flexrow"> contains question index, question text, and value
		liNodes := findAll(reviewQuestionsUL, func(n *xhtml.Node) bool {
			return isElem(n, "li") && hasClassToken(n.Attr, "flexrow")
		})
		for _, li := range liNodes {
			// question text
			qtextDiv := findFirst(li, func(n *xhtml.Node) bool {
				return isElem(n, "div") && hasClassToken(n.Attr, "question-text")
			})
			var questionTitle string
			if qtextDiv != nil {
				questionTitle = textContent(qtextDiv)
			}

			// value: the last <div> child (after index and question-text divs) holds "Yes" or "No"
			// We walk direct div children and pick the third one.
			var valueDivs []*xhtml.Node
			for c := li.FirstChild; c != nil; c = c.NextSibling {
				if isElem(c, "div") {
					valueDivs = append(valueDivs, c)
				}
			}
			var valueStr string
			if len(valueDivs) >= 3 {
				valueStr = strings.TrimSpace(textContent(valueDivs[len(valueDivs)-1]))
			}

			// Resolve value to bool
			var ratingValue interface{}
			switch strings.ToLower(valueStr) {
			case "yes":
				ratingValue = true
			case "no":
				ratingValue = false
			default:
				// Try numeric
				if n, err := strconv.Atoi(valueStr); err == nil {
					ratingValue = n
				} else {
					ratingValue = valueStr
				}
			}

			// Resolve question ID from title
			questionID := 0
			if questionsByTitle != nil {
				if id, ok := questionsByTitle[strings.ToLower(questionTitle)]; ok {
					questionID = id
				}
			}

			var qDetails *QuestionData
			if questionTitle != "" {
				qDetails = &QuestionData{
					ID:    questionID,
					Title: questionTitle,
				}
			}

			ratings = append(ratings, Rating{
				Question:        questionID,
				Value:           ratingValue,
				QuestionDetails: qDetails,
			})
		}
	}

	// ── extract proposed_tracks (change_tracks action only) ───────────────────
	// HTML structure (inside i-box-content, after ratings-details):
	//   Possible destination tracks:
	//     <div class="review-group truncate-text">
	//       <span title="GroupName: TrackCode - TrackTitle">display text</span>
	//     </div>, ...
	// These review-group divs hold a <span title="..."> (no anchor), which
	// distinguishes them from the reviewer's own track review-group (which uses
	// <a href="...reviewing/{id}/">).
	var proposedTracks []Track
	if proposedAction == "change_tracks" {
		// Search root: prefer i-box-content so we avoid the reviewer-track
		// review-group in i-timeline-item-metadata, but fall back to the full
		// reviewDiv if i-box-content is not found.
		searchRoot := findFirst(reviewDiv, func(n *xhtml.Node) bool {
			return isElem(n, "div") && hasClassToken(n.Attr, "i-box-content")
		})
		if searchRoot == nil {
			searchRoot = reviewDiv
		}

		reviewGroupDivs := findAll(searchRoot, func(n *xhtml.Node) bool {
			return isElem(n, "div") && hasClassToken(n.Attr, "review-group")
		})
		for _, rgDiv := range reviewGroupDivs {
			// Skip any review-group that contains an <a> element — those are the
			// reviewer's own track group, not proposed destination tracks.
			if findFirst(rgDiv, func(n *xhtml.Node) bool { return isElem(n, "a") }) != nil {
				continue
			}
			// Find the <span title="..."> that carries the full track description.
			spanNode := findFirst(rgDiv, func(n *xhtml.Node) bool {
				return isElem(n, "span") && getAttr(n.Attr, "title") != ""
			})
			if spanNode == nil {
				continue
			}
			fullTitle := getAttr(spanNode.Attr, "title")
			// title format: "GroupName: TrackCode - TrackTitle"
			// e.g. "MC8 Applications...: MC8.1 - MC8.1 First Track in MC8..."
			var tCode, tTitle string
			dashParts := strings.SplitN(fullTitle, " - ", 2)
			if len(dashParts) == 2 {
				colonParts := strings.SplitN(dashParts[0], ": ", 2)
				if len(colonParts) == 2 {
					tCode = strings.TrimSpace(colonParts[1])
				}
				tTitle = strings.TrimSpace(dashParts[1])
			} else {
				tTitle = fullTitle
			}

			t := Track{Code: tCode, Title: tTitle}
			// Resolve track ID from lookup map if available.
			if tracksByCode != nil && tCode != "" {
				if lt, ok := tracksByCode[tCode]; ok {
					t.ID = lt.ID
					if t.Title == "" {
						t.Title = lt.Title
					}
				}
			}
			proposedTracks = append(proposedTracks, t)
		}
	}
	if proposedTracks == nil {
		proposedTracks = []Track{}
	}

	// ── extract comment ────────────────────────────────────────────────────────
	var comment string
	markdownDiv := findFirst(reviewDiv, func(n *xhtml.Node) bool {
		return isElem(n, "div") && hasClassToken(n.Attr, "markdown-text")
	})
	if markdownDiv != nil {
		comment = strings.TrimSpace(textContent(markdownDiv))
	}

	// ── assemble the Review struct ─────────────────────────────────────────────
	review := &Review{
		ID:                      reviewID,
		CreatedDT:               createdDT,
		ProposedAction:          proposedAction,
		ProposedContribType:     proposedContribType,
		Ratings:                 ratings,
		Comment:                 comment,
		Track:                   track,
		ProposedTracks:          proposedTracks,
		ProposedRelatedAbstract: proposedRelatedAbstract,
		User: Reviewer{
			ID:        reviewerUserID,
			FullName:  reviewerFullName,
			AvatarURL: reviewerAvatarURL,
		},
	}

	return review, nil
}

// GetReviewFromAbstractPage fetches the abstract display page for the given
// abstract database ID and parses the current user's review from it.
// questionsByTitle maps lower-cased question titles to their numeric IDs so that
// Rating.Question fields can be populated correctly.
// abstractsByDBID, tracksByCode, and contribTypesByName are optional lookup maps
// forwarded to ParseReviewFromHTML to resolve proposed_related_abstract,
// proposed_tracks, and proposed_contrib_type respectively.
// Returns nil, nil when the page exists but contains no review by this user.
func (c *IndicoClient) GetReviewFromAbstractPage(ctx context.Context, abstractID int, questionsByTitle map[string]int, abstractsByDBID map[int]*RelatedAbstract, tracksByCode map[string]*Track, contribTypesByName map[string]int) (*Review, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/event/%d/abstracts/%d", c.EventID, abstractID)
	u.Path = joinPaths(u.Path, path)

	ctxReq, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxReq, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

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

	// Cache CSRF token if available
	if _, csrf := parseAbstractIDsAndCSRFFromRoot(doc); csrf != "" {
		c.csrfToken = csrf
	}

	return ParseReviewFromHTML(doc, questionsByTitle, abstractsByDBID, tracksByCode, contribTypesByName)
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

// ReviewSubmissionRequest represents the parameters for submitting a review.
type ReviewSubmissionRequest struct {
	ReviewID                 *int        // Review ID for editing existing review (nil for new review)
	TrackID                  int         // The review track ID
	QuestionRatings          map[int]int // Question ID -> Rating value (0 or 1)
	ProposedAction           string      // accept, reject, change_tracks, mark_as_duplicate, merge
	ProposedContributionType *int        // Required for accept/reject actions (can be None)
	ProposedTracks           []int       // Required for change_tracks action
	ProposedRelatedAbstract  *int        // Required for mark_as_duplicate/merge actions
	Comment                  string      // Review comment
}

// SubmitAbstractReview submits a review for an abstract.
// The abstractID parameter is the database ID of the abstract being reviewed.
// If ReviewID is provided in the request, it will edit an existing review, track id is not required in this case,
// Otherwise, it will create a new review for the specified track.
func (c *IndicoClient) SubmitAbstractReview(ctx context.Context, abstractID int, req *ReviewSubmissionRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate required fields
	if req.TrackID <= 0 {
		return fmt.Errorf("track_id is required")
	}
	if req.ProposedAction == "" {
		return fmt.Errorf("proposed_action is required")
	}

	// Validate action-specific requirements
	switch req.ProposedAction {
	case "accept", "reject":
		// proposed_contribution_type is required (can be None)
		// We allow nil pointer to represent "None"
	case "change_tracks":
		if len(req.ProposedTracks) == 0 {
			return fmt.Errorf("proposed_tracks is required for change_tracks action")
		}
	case "mark_as_duplicate", "merge":
		if req.ProposedRelatedAbstract == nil {
			return fmt.Errorf("proposed_related_abstract is required for %s action", req.ProposedAction)
		}
	default:
		return fmt.Errorf("invalid proposed_action: %s", req.ProposedAction)
	}

	// Ensure we have a CSRF token
	if c.csrfToken == "" {
		return fmt.Errorf("csrf_token is required and not cached")
	}

	// Build the URL based on whether this is an edit or new review
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}

	var path string
	if req.ReviewID != nil {
		// Edit existing review: /event/{event-id}/abstracts/{abstractID}/reviews/{review-id}/edit
		path = fmt.Sprintf("/event/%d/abstracts/%d/reviews/%d/edit", c.EventID, abstractID, *req.ReviewID)
	} else {
		// Create new review: /event/{event-id}/abstracts/{abstractID}/review/track/{track-id}
		path = fmt.Sprintf("/event/%d/abstracts/%d/review/track/%d", c.EventID, abstractID, req.TrackID)
	}
	u.Path = joinPaths(u.Path, path)

	// Build form data
	formData := url.Values{}
	trackPrefix := fmt.Sprintf("track-%d-", req.TrackID)

	// Add CSRF token
	formData.Set(trackPrefix+"csrf_token", c.csrfToken)

	// Add question ratings
	for questionID, rating := range req.QuestionRatings {
		key := fmt.Sprintf("%squestion_%d", trackPrefix, questionID)
		formData.Set(key, strconv.Itoa(rating))
	}

	// Add proposed action
	formData.Set(trackPrefix+"proposed_action", req.ProposedAction)

	// Add action-specific fields
	switch req.ProposedAction {
	case "accept":
		if req.ProposedContributionType != nil {
			formData.Set(trackPrefix+"proposed_contribution_type", strconv.Itoa(*req.ProposedContributionType))
		} else {
			// Use '__None' for None value
			formData.Set(trackPrefix+"proposed_contribution_type", "__None")
		}
	case "reject":
		// No additional fields required for reject
	case "change_tracks":
		for _, trackID := range req.ProposedTracks {
			formData.Add(trackPrefix+"proposed_tracks", strconv.Itoa(trackID))
		}
	case "mark_as_duplicate", "merge":
		if req.ProposedRelatedAbstract != nil {
			formData.Set(trackPrefix+"proposed_related_abstract", strconv.Itoa(*req.ProposedRelatedAbstract))
		}
	}

	// Add comment
	if req.Comment != "" {
		formData.Set(trackPrefix+"comment", req.Comment)
	}

	// Create request with context
	ctxReq, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctxReq, http.MethodPost, u.String(), strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")
	if c.APIToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIToken)
	}

	// Execute request
	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
		return fmt.Errorf("api error: status %d: %s", resp.StatusCode, string(b))
	}

	// Compact log message for review submission (single-line)
	log.Printf("review_submission url=%s payload=%s\n", u.String(), formData.Encode())

	return nil
}
