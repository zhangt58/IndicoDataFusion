package indico

import (
	"os"
	"strings"
	"testing"

	xhtml "golang.org/x/net/html"
)

// sharedQuestionsByTitle is the question map used across all response-fixture tests.
// IDs match those in the JSON fixtures (19 = First priority, 20 = Second Priority).
var sharedQuestionsByTitle = map[string]int{
	"first priority":  19,
	"second priority": 20,
}

// parseFixtureHTML is a helper that reads an HTML file, parses it, and calls
// ParseReviewFromHTML with the supplied lookup maps.
func parseFixtureHTML(t *testing.T, path string, abstractsByDBID map[int]*RelatedAbstract, tracksByTitle map[string]*Track) *Review {
	t.Helper()
	return parseFixtureHTMLWithContrib(t, path, abstractsByDBID, tracksByTitle, nil)
}

// parseFixtureHTMLWithContrib is like parseFixtureHTML but also accepts a
// contribTypesByName map so tests can verify ProposedContribType resolution.
func parseFixtureHTMLWithContrib(t *testing.T, path string, abstractsByDBID map[int]*RelatedAbstract, tracksByTitle map[string]*Track, contribTypesByName map[string]int) *Review {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("fixture file not found at %s, skipping: %v", path, err)
	}
	doc, err := xhtml.Parse(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}
	review, _, err := ParseReviewFromHTML(doc, sharedQuestionsByTitle, abstractsByDBID, tracksByTitle, contribTypesByName)
	if err != nil {
		t.Fatalf("ParseReviewFromHTML returned error: %v", err)
	}
	if review == nil {
		t.Fatal("ParseReviewFromHTML returned nil review, expected a review")
	}
	return review
}

// assertCommonFields validates the fields that are identical across all
// response-fixture reviews (ID=18796, reviewer Tong Zhang, track MC7.2/99,
// created_dt 2026-03-01).
func assertCommonFields(t *testing.T, review *Review) {
	t.Helper()

	if review.ID != 18796 {
		t.Errorf("ID: got %d, want 18796", review.ID)
	}
	wantDT := "2026-03-01T02:45:41.324684+00:00"
	if review.CreatedDT != wantDT {
		t.Errorf("CreatedDT: got %q, want %q", review.CreatedDT, wantDT)
	}
	if review.User.ID != 5071 {
		t.Errorf("User.ID: got %d, want 5071", review.User.ID)
	}
	if review.User.FullName != "Tong Zhang" {
		t.Errorf("User.FullName: got %q, want %q", review.User.FullName, "Tong Zhang")
	}
	if !strings.Contains(review.User.AvatarURL, "/user/5071/") {
		t.Errorf("User.AvatarURL: got %q, expected to contain /user/5071/", review.User.AvatarURL)
	}
	if review.Track.ID != 99 {
		t.Errorf("Track.ID: got %d, want 99", review.Track.ID)
	}
	if review.Track.Code != "MC7.2" {
		t.Errorf("Track.Code: got %q, want MC7.2", review.Track.Code)
	}
	if !strings.Contains(review.Track.Title, "MC7.2") {
		t.Errorf("Track.Title: got %q, expected to contain MC7.2", review.Track.Title)
	}
}

// assertRatings validates that ratings match the expected first/second priority
// boolean values.
func assertRatings(t *testing.T, review *Review, wantFirst, wantSecond bool) {
	t.Helper()
	if len(review.Ratings) != 2 {
		t.Fatalf("Ratings: got %d, want 2", len(review.Ratings))
	}
	r0 := review.Ratings[0]
	if r0.Question != 19 {
		t.Errorf("Ratings[0].Question: got %d, want 19", r0.Question)
	}
	if r0.Value != wantFirst {
		t.Errorf("Ratings[0].Value (First priority): got %v, want %v", r0.Value, wantFirst)
	}
	if r0.QuestionDetails == nil || !strings.EqualFold(r0.QuestionDetails.Title, "First priority") {
		t.Errorf("Ratings[0].QuestionDetails: got %v", r0.QuestionDetails)
	}
	r1 := review.Ratings[1]
	if r1.Question != 20 {
		t.Errorf("Ratings[1].Question: got %d, want 20", r1.Question)
	}
	if r1.Value != wantSecond {
		t.Errorf("Ratings[1].Value (Second Priority): got %v, want %v", r1.Value, wantSecond)
	}
	if r1.QuestionDetails == nil || !strings.EqualFold(r1.QuestionDetails.Title, "Second Priority") {
		t.Errorf("Ratings[1].QuestionDetails: got %v", r1.QuestionDetails)
	}
}

// ── accept ─────────────────────────────────────────────────────────────────────

// TestParseReviewFromHTML_Accept verifies the accept action response.
// Expected JSON: proposed_action=accept, ratings N/Y (false/true), no related abstract,
// no proposed tracks, and no contrib type (fixture has none).
func TestParseReviewFromHTML_Accept(t *testing.T) {
	review := parseFixtureHTML(t, "review-accept_response.html", nil, nil)
	assertCommonFields(t, review)

	if review.ProposedAction != "accept" {
		t.Errorf("ProposedAction: got %q, want accept", review.ProposedAction)
	}
	// JSON: ratings[0].value=false (No), ratings[1].value=true (Yes) → N/Y
	assertRatings(t, review, false, true)

	if review.Comment != "accept with N/Y" {
		t.Errorf("Comment: got %q, want %q", review.Comment, "accept with N/Y")
	}
	if review.ProposedRelatedAbstract != nil {
		t.Errorf("ProposedRelatedAbstract: expected nil, got %+v", review.ProposedRelatedAbstract)
	}
	if len(review.ProposedTracks) != 0 {
		t.Errorf("ProposedTracks: expected empty, got %v", review.ProposedTracks)
	}
	// Fixture has no contribution type in badge → must be nil.
	if review.ProposedContribType != nil {
		t.Errorf("ProposedContribType: expected nil for fixture without contrib type, got %+v", review.ProposedContribType)
	}
}

// TestParseReviewFromHTML_Accept_WithContribType verifies contrib type extraction
// when the badge reads "Proposed to accept as <strong>Contributed Oral Presentation</strong>".
// Uses a synthetic HTML that reproduces the exact Indico template output.
func TestParseReviewFromHTML_Accept_WithContribType(t *testing.T) {
	// Minimal synthetic HTML reproducing the accept+contrib-type badge structure.
	const syntheticHTML = `<!DOCTYPE html><html><body>
<div id="proposal-review-7001" class="i-timeline-item">
  <div class="avatar-div">
    <img src="/user/5071/picture-default/abc" class="ui image circular profile-picture">
  </div>
  <div class="flexrow i-timeline-item-content">
    <div class="i-timeline-item-metadata">
      <div class="f-self-stretch"><strong>Test User</strong> left a review
        <time datetime="2026-03-01T10:00:00+00:00">Mar 1, 2026</time>
      </div>
      <div class="review-group truncate-text">
        <a href="/event/37/abstracts/reviewing/99/" title="GroupX: MC7.2 - MC7.2 Track">GroupX: MC7.2</a>
      </div>
    </div>
    <div class="i-timeline-item-box">
      <div class="i-box-header flexrow">
        <div class="review-badges">
          Proposed to <span class="bold underline semantic-text success">accept</span> as <strong>Contributed Oral Presentation</strong>
        </div>
      </div>
      <div class="i-box-content js-form-container">
        <div class="markdown-text"><p>looks good</p></div>
      </div>
    </div>
  </div>
</div>
</body></html>`

	contribTypesByName := map[string]int{
		"Contributed Oral Presentation": 42,
		"Poster Presentation":           17,
	}

	doc, err := xhtml.Parse(strings.NewReader(syntheticHTML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	review, _, err := ParseReviewFromHTML(doc, sharedQuestionsByTitle, nil, nil, contribTypesByName)
	if err != nil {
		t.Fatalf("ParseReviewFromHTML: %v", err)
	}
	if review == nil {
		t.Fatal("expected review, got nil")
	}

	if review.ProposedAction != "accept" {
		t.Errorf("ProposedAction: got %q, want accept", review.ProposedAction)
	}
	if review.ProposedContribType == nil {
		t.Fatal("ProposedContribType: expected non-nil")
	}
	if review.ProposedContribType.Name != "Contributed Oral Presentation" {
		t.Errorf("ProposedContribType.Name: got %q, want %q", review.ProposedContribType.Name, "Contributed Oral Presentation")
	}
	if review.ProposedContribType.ID != 42 {
		t.Errorf("ProposedContribType.ID: got %d, want 42", review.ProposedContribType.ID)
	}
	if review.Comment != "looks good" {
		t.Errorf("Comment: got %q, want %q", review.Comment, "looks good")
	}
	if len(review.ProposedTracks) != 0 {
		t.Errorf("ProposedTracks: expected empty for accept action, got %v", review.ProposedTracks)
	}
}

// TestParseReviewFromHTML_Accept_WithContribType_NoLookup verifies that when no
// contribTypesByName map is provided, the contrib type name is still extracted
// but the ID is 0.
func TestParseReviewFromHTML_Accept_WithContribType_NoLookup(t *testing.T) {
	const syntheticHTML = `<!DOCTYPE html><html><body>
<div id="proposal-review-7002" class="i-timeline-item">
  <div class="avatar-div">
    <img src="/user/5071/picture-default/abc" class="ui image circular profile-picture">
  </div>
  <div class="flexrow i-timeline-item-content">
    <div class="i-timeline-item-metadata">
      <div class="f-self-stretch"><strong>Test User</strong> left a review
        <time datetime="2026-03-01T10:00:00+00:00">Mar 1, 2026</time>
      </div>
    </div>
    <div class="i-timeline-item-box">
      <div class="i-box-header flexrow">
        <div class="review-badges">
          Proposed to <span class="bold underline semantic-text success">accept</span> as <strong>Poster Presentation</strong>
        </div>
      </div>
      <div class="i-box-content js-form-container">
        <div class="markdown-text"><p>poster is fine</p></div>
      </div>
    </div>
  </div>
</div>
</body></html>`

	doc, err := xhtml.Parse(strings.NewReader(syntheticHTML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// Pass nil contribTypesByName — name should still be extracted, ID=0.
	review, _, err := ParseReviewFromHTML(doc, sharedQuestionsByTitle, nil, nil, nil)
	if err != nil {
		t.Fatalf("ParseReviewFromHTML: %v", err)
	}
	if review == nil {
		t.Fatal("expected review, got nil")
	}
	if review.ProposedContribType == nil {
		t.Fatal("ProposedContribType: expected non-nil even without lookup map")
	}
	if review.ProposedContribType.Name != "Poster Presentation" {
		t.Errorf("ProposedContribType.Name: got %q, want %q", review.ProposedContribType.Name, "Poster Presentation")
	}
	if review.ProposedContribType.ID != 0 {
		t.Errorf("ProposedContribType.ID: got %d, want 0 (no lookup map)", review.ProposedContribType.ID)
	}
}

// ── reject ─────────────────────────────────────────────────────────────────────

// TestParseReviewFromHTML_Reject verifies the reject action response.
// Expected JSON: proposed_action=reject, ratings N/N (false/false), no related abstract, no proposed tracks.
func TestParseReviewFromHTML_Reject(t *testing.T) {
	review := parseFixtureHTML(t, "review-reject_response.html", nil, nil)
	assertCommonFields(t, review)

	if review.ProposedAction != "reject" {
		t.Errorf("ProposedAction: got %q, want reject", review.ProposedAction)
	}
	// JSON: ratings[0].value=false (No), ratings[1].value=false (No) → N/N
	assertRatings(t, review, false, false)

	if review.Comment != "reject with N/N" {
		t.Errorf("Comment: got %q, want %q", review.Comment, "reject with N/N")
	}
	if review.ProposedRelatedAbstract != nil {
		t.Errorf("ProposedRelatedAbstract: expected nil, got %+v", review.ProposedRelatedAbstract)
	}
	if len(review.ProposedTracks) != 0 {
		t.Errorf("ProposedTracks: expected empty, got %v", review.ProposedTracks)
	}
}

// ── mark_as_duplicate ──────────────────────────────────────────────────────────

// TestParseReviewFromHTML_MarkAsDuplicate verifies the mark_as_duplicate action response.
// Expected JSON: proposed_action=mark_as_duplicate, ratings Y/N (true/false),
//
//	proposed_related_abstract={id:143, friendly_id:50, title:"Scaling Up..."}.
func TestParseReviewFromHTML_MarkAsDuplicate(t *testing.T) {
	// Build abstractsByDBID lookup: DB ID 143 → friendly_id 50.
	abstractsByDBID := map[int]*RelatedAbstract{
		143: {
			ID:         143,
			FriendlyID: 50,
			Title:      "Scaling Up the Alba Cabling Database and Plans to Turn Into an Asset Management System (fake)",
		},
	}
	review := parseFixtureHTML(t, "review-mark_as_duplicate_response.html", abstractsByDBID, nil)
	assertCommonFields(t, review)

	if review.ProposedAction != "mark_as_duplicate" {
		t.Errorf("ProposedAction: got %q, want mark_as_duplicate", review.ProposedAction)
	}
	// JSON: ratings[0].value=true (Yes), ratings[1].value=false (No) → Y/N
	assertRatings(t, review, true, false)

	if review.Comment != "mark as duplicate with #50, with N/Y from Y/N" {
		t.Errorf("Comment: got %q, want %q", review.Comment, "mark as duplicate with #50, with N/Y from Y/N")
	}

	ra := review.ProposedRelatedAbstract
	if ra == nil {
		t.Fatal("ProposedRelatedAbstract: expected non-nil")
	}
	if ra.ID != 143 {
		t.Errorf("ProposedRelatedAbstract.ID: got %d, want 143", ra.ID)
	}
	if ra.FriendlyID != 50 {
		t.Errorf("ProposedRelatedAbstract.FriendlyID: got %d, want 50", ra.FriendlyID)
	}
	if !strings.Contains(ra.Title, "Scaling Up") {
		t.Errorf("ProposedRelatedAbstract.Title: got %q, expected to contain 'Scaling Up'", ra.Title)
	}
	if len(review.ProposedTracks) != 0 {
		t.Errorf("ProposedTracks: expected empty, got %v", review.ProposedTracks)
	}
}

// TestParseReviewFromHTML_MarkAsDuplicate_Fallback verifies that when no
// abstractsByDBID map is provided, the parser still populates the DB ID and
// title from the HTML anchor text.
func TestParseReviewFromHTML_MarkAsDuplicate_Fallback(t *testing.T) {
	review := parseFixtureHTML(t, "review-mark_as_duplicate_response.html", nil, nil)

	if review.ProposedAction != "mark_as_duplicate" {
		t.Errorf("ProposedAction: got %q, want mark_as_duplicate", review.ProposedAction)
	}
	ra := review.ProposedRelatedAbstract
	if ra == nil {
		t.Fatal("ProposedRelatedAbstract: expected non-nil even without lookup map")
	}
	if ra.ID != 143 {
		t.Errorf("ProposedRelatedAbstract.ID (fallback): got %d, want 143", ra.ID)
	}
	// FriendlyID will be 0 without the lookup map — that is expected.
	if ra.FriendlyID != 0 {
		t.Errorf("ProposedRelatedAbstract.FriendlyID (fallback, no map): got %d, want 0", ra.FriendlyID)
	}
	if !strings.Contains(ra.Title, "Scaling Up") {
		t.Errorf("ProposedRelatedAbstract.Title (fallback): got %q, expected to contain 'Scaling Up'", ra.Title)
	}
}

// ── merge ──────────────────────────────────────────────────────────────────────

// TestParseReviewFromHTML_Merge verifies the merge action response.
// Expected JSON: proposed_action=merge, ratings Y/N (true/false),
//
//	proposed_related_abstract={id:149, friendly_id:56, title:"Wow.. congratulations!"}.
func TestParseReviewFromHTML_Merge(t *testing.T) {
	abstractsByDBID := map[int]*RelatedAbstract{
		149: {
			ID:         149,
			FriendlyID: 56,
			Title:      "Wow.. congratulations!",
		},
	}
	review := parseFixtureHTML(t, "review-merge_response.html", abstractsByDBID, nil)
	assertCommonFields(t, review)

	if review.ProposedAction != "merge" {
		t.Errorf("ProposedAction: got %q, want merge", review.ProposedAction)
	}
	// JSON: ratings[0].value=true (Yes), ratings[1].value=false (No) → Y/N
	assertRatings(t, review, true, false)

	if review.Comment != "merge to #56, with N/Y from Y/N" {
		t.Errorf("Comment: got %q, want %q", review.Comment, "merge to #56, with N/Y from Y/N")
	}

	ra := review.ProposedRelatedAbstract
	if ra == nil {
		t.Fatal("ProposedRelatedAbstract: expected non-nil")
	}
	if ra.ID != 149 {
		t.Errorf("ProposedRelatedAbstract.ID: got %d, want 149", ra.ID)
	}
	if ra.FriendlyID != 56 {
		t.Errorf("ProposedRelatedAbstract.FriendlyID: got %d, want 56", ra.FriendlyID)
	}
	if ra.Title != "Wow.. congratulations!" {
		t.Errorf("ProposedRelatedAbstract.Title: got %q, want %q", ra.Title, "Wow.. congratulations!")
	}
	if len(review.ProposedTracks) != 0 {
		t.Errorf("ProposedTracks: expected empty, got %v", review.ProposedTracks)
	}
}

// ── change_tracks ──────────────────────────────────────────────────────────────

// TestParseReviewFromHTML_ChangeTracks verifies the change_tracks action response.
// Expected: proposed_action=change_tracks, proposed_tracks=[MC8.1, MC8.2] with
// IDs resolved from the tracksByTitle map, ratings Y/N (true/false).
// Note: there is no JSON fixture for change_tracks; the expected values are
// derived directly from the HTML content.
func TestParseReviewFromHTML_ChangeTracks(t *testing.T) {
	// Build tracksByTitle lookup keyed by the Track.Title value, which is the
	// part after " - " in the span's title attribute
	// ("GroupName: Code - Title").  Code is often empty in real Indico data so
	// Title is the reliable lookup key.
	tracksByTitle := map[string]*Track{
		"MC8.1 First Track in MC8 Track Group Applications of Accelerators, Technology Transfer and Industrial Relations": {
			ID: 101, Code: "MC8.1", Title: "MC8.1 First Track in MC8 Track Group Applications of Accelerators, Technology Transfer and Industrial Relations",
		},
		"MC8.2 Second Track in MC8 Track Group": {
			ID: 102, Code: "MC8.2", Title: "MC8.2 Second Track in MC8 Track Group",
		},
	}
	review := parseFixtureHTML(t, "review-change_tracks_response.html", nil, tracksByTitle)
	assertCommonFields(t, review)

	if review.ProposedAction != "change_tracks" {
		t.Errorf("ProposedAction: got %q, want change_tracks", review.ProposedAction)
	}
	// HTML ratings: First priority=Yes (true), Second Priority=No (false) → Y/N
	assertRatings(t, review, true, false)

	if review.Comment != "to track MC8.2, 8.1, with Y/N from N/Y" {
		t.Errorf("Comment: got %q, want %q", review.Comment, "to track MC8.2, 8.1, with Y/N from N/Y")
	}
	if review.ProposedRelatedAbstract != nil {
		t.Errorf("ProposedRelatedAbstract: expected nil, got %+v", review.ProposedRelatedAbstract)
	}

	// Expect exactly two proposed tracks: MC8.1 and MC8.2.
	if len(review.ProposedTracks) != 2 {
		t.Fatalf("ProposedTracks: got %d, want 2; tracks=%v", len(review.ProposedTracks), review.ProposedTracks)
	}
	codes := map[string]bool{}
	ids := map[int]bool{}
	for _, pt := range review.ProposedTracks {
		codes[pt.Code] = true
		ids[pt.ID] = true
	}
	for _, wantCode := range []string{"MC8.1", "MC8.2"} {
		if !codes[wantCode] {
			t.Errorf("ProposedTracks: missing track code %q; got %v", wantCode, review.ProposedTracks)
		}
	}
	for _, wantID := range []int{101, 102} {
		if !ids[wantID] {
			t.Errorf("ProposedTracks: missing track ID %d; got %v", wantID, review.ProposedTracks)
		}
	}
	// The reviewer's own track (MC7.2) must NOT appear in proposed tracks.
	if codes["MC7.2"] {
		t.Errorf("ProposedTracks: reviewer's own track MC7.2 must not appear in proposed tracks")
	}
}

// TestParseReviewFromHTML_ChangeTracks_NoLookup verifies that when no
// tracksByTitle map is provided, proposed tracks are still extracted with
// code+title from the HTML but ID=0.
func TestParseReviewFromHTML_ChangeTracks_NoLookup(t *testing.T) {
	review := parseFixtureHTML(t, "review-change_tracks_response.html", nil, nil)

	if review.ProposedAction != "change_tracks" {
		t.Errorf("ProposedAction: got %q, want change_tracks", review.ProposedAction)
	}
	if len(review.ProposedTracks) != 2 {
		t.Fatalf("ProposedTracks: got %d, want 2 (even without lookup map)", len(review.ProposedTracks))
	}
	for _, pt := range review.ProposedTracks {
		// Code is extracted from the HTML title attribute even without a lookup map.
		if pt.Code == "" {
			t.Errorf("ProposedTracks entry missing Code: %+v", pt)
		}
		// Title is always extracted from the span title attribute.
		if pt.Title == "" {
			t.Errorf("ProposedTracks entry missing Title: %+v", pt)
		}
		// ID should be 0 when no lookup map is provided.
		if pt.ID != 0 {
			t.Errorf("ProposedTracks entry ID should be 0 without lookup map: %+v", pt)
		}
	}
}

// TestParseReviewFromHTML_ChangeTracks_Synthetic verifies the <a>-skipping logic
// using a minimal synthetic HTML that mirrors the exact Indico DOM structure.
// It also exercises the fallback path where i-box-content is absent, to ensure
// the parser still finds proposed tracks directly from the reviewDiv.
func TestParseReviewFromHTML_ChangeTracks_Synthetic(t *testing.T) {
	// Minimal HTML that reproduces the Indico review DOM:
	//  - i-timeline-item-metadata contains review-group with <a> (reviewer's track — must be skipped)
	//  - i-box-content contains review-group with <span title="..."> (proposed tracks)
	const syntheticHTML = `<!DOCTYPE html><html><body>
<div id="proposal-review-99" class="i-timeline-item">
  <div class="avatar-div">
    <img src="/user/5071/picture-default/abc" class="ui image circular profile-picture">
  </div>
  <div class="flexrow i-timeline-item-content">
    <div class="i-timeline-item-metadata">
      <div class="f-self-stretch"><strong>Test User</strong> left a review
        <time datetime="2026-01-01T00:00:00+00:00">Jan 1, 2026</time>
      </div>
      <div class="review-group truncate-text">
        <a href="/event/37/abstracts/reviewing/10/" title="GroupA: T1 - T1 Reviewer Track">GroupA: T1</a>
      </div>
    </div>
    <div class="i-timeline-item-box">
      <div class="i-box-header flexrow">
        <div class="review-badges">
          Proposed to <span class="bold underline semantic-text warning">change tracks</span>
        </div>
      </div>
      <div class="i-box-content js-form-container">
        <div>
          Possible destination tracks:
          <div class="review-group truncate-text">
            <span title="GroupB: T2 - T2 First Destination Track">GroupB: T2</span>
          </div>,
          <div class="review-group truncate-text">
            <span title="GroupB: T3 - T3 Second Destination Track">GroupB: T3</span>
          </div>
        </div>
        <div class="markdown-text"><p>synthetic comment</p></div>
      </div>
    </div>
  </div>
</div>
</body></html>`

	tracksByTitle := map[string]*Track{
		"T2 First Destination Track":  {ID: 200, Code: "T2", Title: "T2 First Destination Track"},
		"T3 Second Destination Track": {ID: 300, Code: "T3", Title: "T3 Second Destination Track"},
	}

	doc, err := xhtml.Parse(strings.NewReader(syntheticHTML))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	review, _, err := ParseReviewFromHTML(doc, sharedQuestionsByTitle, nil, tracksByTitle, nil)
	if err != nil {
		t.Fatalf("ParseReviewFromHTML: %v", err)
	}
	if review == nil {
		t.Fatal("expected review, got nil")
	}

	if review.ProposedAction != "change_tracks" {
		t.Errorf("ProposedAction: got %q, want change_tracks", review.ProposedAction)
	}
	if len(review.ProposedTracks) != 2 {
		t.Fatalf("ProposedTracks: got %d, want 2; tracks=%v", len(review.ProposedTracks), review.ProposedTracks)
	}
	codes := map[string]bool{}
	for _, pt := range review.ProposedTracks {
		codes[pt.Code] = true
	}
	if !codes["T2"] || !codes["T3"] {
		t.Errorf("ProposedTracks: expected T2 and T3; got %v", review.ProposedTracks)
	}
	// Reviewer's own track T1 must not appear.
	if codes["T1"] {
		t.Errorf("ProposedTracks: reviewer's own track T1 must not appear in proposed tracks")
	}
	// IDs must be resolved from the lookup map.
	for _, pt := range review.ProposedTracks {
		if pt.ID == 0 {
			t.Errorf("ProposedTracks: ID not resolved for track %q", pt.Code)
		}
	}
	if review.Comment != "synthetic comment" {
		t.Errorf("Comment: got %q, want %q", review.Comment, "synthetic comment")
	}
}

// ── legacy 1004.html fixture ───────────────────────────────────────────────────

// TestParseReviewFromHTML verifies that ParseReviewFromHTML correctly
// extracts the review data from 1004.html and matches the expected values from
// 1004.json.
func TestParseReviewFromHTML(t *testing.T) {
	htmlPath := "../../1004.html"
	data, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Skipf("fixture file not found at %s, skipping: %v", htmlPath, err)
	}

	doc, err := xhtml.Parse(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	questionsByTitle := map[string]int{
		"first priority":  19,
		"second priority": 20,
	}

	review, _, err := ParseReviewFromHTML(doc, questionsByTitle, nil, nil, nil)
	if err != nil {
		t.Fatalf("ParseReviewFromHTML returned error: %v", err)
	}
	if review == nil {
		t.Fatal("ParseReviewFromHTML returned nil review, expected a review")
	}

	if review.ID != 18795 {
		t.Errorf("ID: got %d, want 18795", review.ID)
	}
	wantDT := "2026-02-28T22:38:52.294861+00:00"
	if review.CreatedDT != wantDT {
		t.Errorf("CreatedDT: got %q, want %q", review.CreatedDT, wantDT)
	}
	if review.ProposedAction != "reject" {
		t.Errorf("ProposedAction: got %q, want %q", review.ProposedAction, "reject")
	}
	if review.Comment != "from a regular reviewer" {
		t.Errorf("Comment: got %q, want %q", review.Comment, "from a regular reviewer")
	}
	if review.Track.ID != 88 {
		t.Errorf("Track.ID: got %d, want 88", review.Track.ID)
	}
	if review.Track.Code != "MC7.1" {
		t.Errorf("Track.Code: got %q, want %q", review.Track.Code, "MC7.1")
	}
	if !strings.Contains(review.Track.Title, "MC7.1") {
		t.Errorf("Track.Title: got %q, expected it to contain %q", review.Track.Title, "MC7.1")
	}
	if review.User.ID != 5071 {
		t.Errorf("User.ID: got %d, want 5071", review.User.ID)
	}
	if review.User.FullName != "Tong Zhang" {
		t.Errorf("User.FullName: got %q, want %q", review.User.FullName, "Tong Zhang")
	}
	if !strings.Contains(review.User.AvatarURL, "/user/5071/") {
		t.Errorf("User.AvatarURL: got %q, expected to contain %q", review.User.AvatarURL, "/user/5071/")
	}
	if len(review.Ratings) != 2 {
		t.Fatalf("Ratings: got %d, want 2", len(review.Ratings))
	}
	r0 := review.Ratings[0]
	if r0.Question != 19 {
		t.Errorf("Ratings[0].Question: got %d, want 19", r0.Question)
	}
	if r0.Value != true {
		t.Errorf("Ratings[0].Value: got %v, want true", r0.Value)
	}
	if r0.QuestionDetails == nil || r0.QuestionDetails.Title != "First priority" {
		t.Errorf("Ratings[0].QuestionDetails: got %v", r0.QuestionDetails)
	}
	r1 := review.Ratings[1]
	if r1.Question != 20 {
		t.Errorf("Ratings[1].Question: got %d, want 20", r1.Question)
	}
	if r1.Value != false {
		t.Errorf("Ratings[1].Value: got %v, want false", r1.Value)
	}
	if r1.QuestionDetails == nil || !strings.EqualFold(r1.QuestionDetails.Title, "Second Priority") {
		t.Errorf("Ratings[1].QuestionDetails: got %v", r1.QuestionDetails)
	}
	if review.ProposedTracks == nil {
		t.Errorf("ProposedTracks should be non-nil (empty slice)")
	}
	if review.ModifiedDT != nil {
		t.Errorf("ModifiedDT: expected nil from HTML parse, got %v", *review.ModifiedDT)
	}
}

// TestParseReviewFromHTML_NoReview verifies that a minimal HTML page without
// a proposal-review div returns nil, nil.
func TestParseReviewFromHTML_NoReview(t *testing.T) {
	const page = `<!DOCTYPE html><html><body><p>No review here.</p></body></html>`
	doc, err := xhtml.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	review, _, err := ParseReviewFromHTML(doc, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if review != nil {
		t.Errorf("expected nil review, got %+v", review)
	}
}
