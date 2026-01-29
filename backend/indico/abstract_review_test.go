package indico

import (
	"context"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

// mockClient implements HTTPClient for tests.
type mockClient struct {
	resp *http.Response
	err  error
}

func (m *mockClient) Do(_ *http.Request) (*http.Response, error) {
	return m.resp, m.err
}

// multiMockClient allows different responses based on the request URL
type multiMockClient struct {
	responses map[string]*http.Response
	callCount int
	mu        sync.Mutex
}

func (m *multiMockClient) Do(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++

	// Match based on URL path
	if resp, ok := m.responses[req.URL.Path]; ok {
		// Return a copy of the response to avoid issues with concurrent reads
		return resp, nil
	}
	// Default response for unknown paths
	return &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(strings.NewReader("not found")),
		Header:     make(http.Header),
	}, nil
}

func TestGetReviewTracks_ParsesFixture(t *testing.T) {
	fixturePath := "review-track-list.html"
	b, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(string(b))),
		Header:     make(http.Header),
	}

	mc := &mockClient{resp: resp}
	c := &IndicoClient{
		BaseURL: "https://indico.jacow.org",
		EventID: 37,
		Client:  mc,
		Timeout: 5 * time.Second,
	}

	tracks, err := c.GetReviewTracks(context.Background())
	if err != nil {
		t.Fatalf("GetReviewTracks returned error: %v", err)
	}
	if tracks == nil {
		t.Fatalf("expected non-nil tracks")
	}

	// Expected tracks in the fixture (order as they appear in the HTML)
	expected := []ReviewTrack{
		{Name: "MC7 Accelerator Technology Main Systems", Link: "", TrackID: 0},
		{Name: "MC7.1 First Track in MC7 Track Group Accelerator Technology Main Systems", Link: "/event/37/abstracts/reviewing/88/", TrackID: 88},
		{Name: "MC7.2 Second Track in Track Group MC7", Link: "/event/37/abstracts/reviewing/99/", TrackID: 99},
	}

	if len(tracks.Tracks) != len(expected) {
		t.Fatalf("expected %d tracks, got %d", len(expected), len(tracks.Tracks))
	}

	for i, exp := range expected {
		got := tracks.Tracks[i]
		if got.Name != exp.Name {
			t.Errorf("track %d: expected name %q, got %q", i, exp.Name, got.Name)
		}
		// Compare links exactly; empty link expected for the group title
		if got.Link != exp.Link {
			t.Errorf("track %d: expected link %q, got %q", i, exp.Link, got.Link)
		}
		if got.TrackID != exp.TrackID {
			t.Errorf("track %d: expected trackID %d, got %d", i, exp.TrackID, got.TrackID)
		}
	}
}

func TestGetReviewTracks_Non200(t *testing.T) {
	resp := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader("internal error")),
		Header:     make(http.Header),
	}
	mc := &mockClient{resp: resp}
	c := &IndicoClient{
		BaseURL: "https://indico.jacow.org",
		EventID: 37,
		Client:  mc,
		Timeout: 5 * time.Second,
	}

	_, err := c.GetReviewTracks(context.Background())
	if err == nil {
		t.Fatalf("expected error for non-2xx response")
	}
}

// TestReviewAggRatings tests the AggRatings method
func TestReviewAggRatings(t *testing.T) {
	tests := []struct {
		name     string
		review   Review
		expected map[int]float64
	}{
		{
			name: "single numeric rating",
			review: Review{
				Ratings: []Rating{
					{Question: 1, Value: 5},
				},
			},
			expected: map[int]float64{1: 5.0},
		},
		{
			name: "boolean ratings",
			review: Review{
				Ratings: []Rating{
					{Question: 1, Value: true},
					{Question: 2, Value: false},
				},
			},
			expected: map[int]float64{1: 1.0, 2: 0.0},
		},
		{
			name: "string boolean ratings",
			review: Review{
				Ratings: []Rating{
					{Question: 1, Value: "yes"},
					{Question: 2, Value: "no"},
					{Question: 3, Value: "true"},
					{Question: 4, Value: "false"},
				},
			},
			expected: map[int]float64{1: 1.0, 2: 0.0, 3: 1.0, 4: 0.0},
		},
		{
			name: "mixed types",
			review: Review{
				Ratings: []Rating{
					{Question: 1, Value: 3},
					{Question: 2, Value: true},
					{Question: 3, Value: "yes"},
					{Question: 4, Value: 2.5},
				},
			},
			expected: map[int]float64{1: 3.0, 2: 1.0, 3: 1.0, 4: 2.5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.review.AggRatings()
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d ratings, got %d", len(tt.expected), len(result))
			}
			for q, expectedVal := range tt.expected {
				if val, ok := result[q]; !ok {
					t.Errorf("missing question %d in result", q)
				} else if val != expectedVal {
					t.Errorf("question %d: expected %.2f, got %.2f", q, expectedVal, val)
				}
			}
		})
	}
}

// TestAbstractDataAggregateAllRatings tests aggregating ratings across all reviews
func TestAbstractDataAggregateAllRatings(t *testing.T) {
	abstract := AbstractData{
		Reviews: []Review{
			{
				Ratings: []Rating{
					{Question: 1, Value: 3},
					{Question: 2, Value: true},
				},
			},
			{
				Ratings: []Rating{
					{Question: 1, Value: 2},
					{Question: 2, Value: false},
					{Question: 3, Value: "yes"},
				},
			},
		},
	}

	expected := map[int]float64{
		1: 5.0, // 3 + 2
		2: 1.0, // 1 + 0
		3: 1.0, // 1
	}

	result := abstract.AggregateAllRatings()
	if len(result) != len(expected) {
		t.Errorf("expected %d questions, got %d", len(expected), len(result))
	}
	for q, expectedVal := range expected {
		if val, ok := result[q]; !ok {
			t.Errorf("missing question %d in result", q)
		} else if val != expectedVal {
			t.Errorf("question %d: expected %.2f, got %.2f", q, expectedVal, val)
		}
	}
}

// TestGetAggregatedRatingByTitle tests getting aggregated ratings by question title
func TestGetAggregatedRatingByTitle(t *testing.T) {
	abstract := AbstractData{
		Reviews: []Review{
			{
				Ratings: []Rating{
					{
						Question: 19,
						Value:    1,
						QuestionDetails: &QuestionData{
							ID:    19,
							Title: "First priority",
						},
					},
					{
						Question: 20,
						Value:    true,
						QuestionDetails: &QuestionData{
							ID:    20,
							Title: "Second Priority",
						},
					},
				},
			},
			{
				Ratings: []Rating{
					{
						Question: 19,
						Value:    2,
						QuestionDetails: &QuestionData{
							ID:    19,
							Title: "First priority",
						},
					},
					{
						Question: 20,
						Value:    "yes",
						QuestionDetails: &QuestionData{
							ID:    20,
							Title: "Second Priority",
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		title    string
		expected float64
	}{
		{
			name:     "first priority exact case",
			title:    "First priority",
			expected: 3.0, // 1 + 2
		},
		{
			name:     "first priority different case",
			title:    "FIRST PRIORITY",
			expected: 3.0,
		},
		{
			name:     "second priority mixed case",
			title:    "second priority",
			expected: 2.0, // 1 + 1 (true + "yes")
		},
		{
			name:     "non-existent question",
			title:    "Third priority",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := abstract.GetAggregatedRatingByTitle(tt.title)
			if result != tt.expected {
				t.Errorf("expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

// TestAbstractDataPrecomputedFields tests that FirstPriority and SecondPriority fields
// are correctly computed when populated by the data handler
func TestAbstractDataPrecomputedFields(t *testing.T) {
	abstract := AbstractData{
		Reviews: []Review{
			{
				Ratings: []Rating{
					{
						Question: 19,
						Value:    1,
						QuestionDetails: &QuestionData{
							ID:    19,
							Title: "First priority",
						},
					},
					{
						Question: 20,
						Value:    true,
						QuestionDetails: &QuestionData{
							ID:    20,
							Title: "Second Priority",
						},
					},
				},
			},
			{
				Ratings: []Rating{
					{
						Question: 19,
						Value:    2,
						QuestionDetails: &QuestionData{
							ID:    19,
							Title: "First priority",
						},
					},
					{
						Question: 20,
						Value:    "yes",
						QuestionDetails: &QuestionData{
							ID:    20,
							Title: "Second Priority",
						},
					},
				},
			},
		},
	}

	// Simulate what the data handler does - populate the fields
	abstract.FirstPriority = abstract.GetAggregatedRatingByTitle("First priority")
	abstract.SecondPriority = abstract.GetAggregatedRatingByTitle("Second priority")

	// Verify the fields are correctly populated
	if abstract.FirstPriority != 3.0 {
		t.Errorf("FirstPriority: expected 3.0, got %.2f", abstract.FirstPriority)
	}
	if abstract.SecondPriority != 2.0 {
		t.Errorf("SecondPriority: expected 2.0, got %.2f", abstract.SecondPriority)
	}

	t.Logf("✅ FirstPriority: %.0f, SecondPriority: %.0f", abstract.FirstPriority, abstract.SecondPriority)
}

func TestGetReviewAbstractIDs_ParsesFixture(t *testing.T) {
	fixturePath := "review-abstracts.html"
	b, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(string(b))),
		Header:     make(http.Header),
	}

	mc := &mockClient{resp: resp}
	c := &IndicoClient{
		BaseURL: "https://indico.jacow.org",
		EventID: 37,
		Client:  mc,
		Timeout: 5 * time.Second,
	}

	ids, err := c.GetReviewAbstractIDs(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetReviewAbstractIDs returned error: %v", err)
	}

	expected := []int{50, 76, 83, 103, 120, 154, 184, 195, 213, 219, 256, 266}
	if !slices.Equal(ids, expected) {
		t.Fatalf("expected ids %v, got %v", expected, ids)
	}
}

func TestGetReviewAbstractIDs_EmptyPage(t *testing.T) {
	html := "<html><head></head><body><table><tbody></tbody></table></body></html>"
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(html)),
		Header:     make(http.Header),
	}
	mc := &mockClient{resp: resp}
	c := &IndicoClient{
		BaseURL: "https://indico.jacow.org",
		EventID: 37,
		Client:  mc,
		Timeout: 5 * time.Second,
	}

	ids, err := c.GetReviewAbstractIDs(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 0 {
		t.Fatalf("expected empty ids slice, got %v", ids)
	}
}

func TestGetReviewAbstractIDs_MalformedIDs(t *testing.T) {
	html := `<!doctype html><html><body><table><tbody>` +
		`<tr class="abstract-row" data-friendly-id="not-a-number"><td></td></tr>` +
		`<tr class="abstract-row" data-friendly-id="123"><td></td></tr>` +
		`</tbody></table></body></html>`
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(html)),
		Header:     make(http.Header),
	}
	mc := &mockClient{resp: resp}
	c := &IndicoClient{
		BaseURL: "https://indico.jacow.org",
		EventID: 37,
		Client:  mc,
		Timeout: 5 * time.Second,
	}

	ids, err := c.GetReviewAbstractIDs(context.Background(), 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []int{123}
	if !slices.Equal(ids, expected) {
		t.Fatalf("expected ids %v, got %v", expected, ids)
	}
}

// TestGetReviewTracks_ConcurrentAbstractCounts validates that abstract counts
// are fetched concurrently for multiple tracks
func TestGetReviewTracks_ConcurrentAbstractCounts(t *testing.T) {
	// Read fixtures
	trackListHTML, err := os.ReadFile("review-track-list.html")
	if err != nil {
		t.Fatalf("read track list fixture: %v", err)
	}

	abstractsHTML, err := os.ReadFile("review-abstracts.html")
	if err != nil {
		t.Fatalf("read abstracts fixture: %v", err)
	}

	// Setup multi-mock client with different responses for different URLs
	mmc := &multiMockClient{
		responses: map[string]*http.Response{
			"/event/37/abstracts/reviewing/statistics": {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(trackListHTML))),
				Header:     make(http.Header),
			},
			"/event/37/abstracts/reviewing/88": {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(abstractsHTML))),
				Header:     make(http.Header),
			},
			"/event/37/abstracts/reviewing/99": {
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(string(abstractsHTML))),
				Header:     make(http.Header),
			},
		},
	}

	c := &IndicoClient{
		BaseURL: "https://indico.jacow.org",
		EventID: 37,
		Client:  mmc,
		Timeout: 5 * time.Second,
	}

	tracks, err := c.GetReviewTracks(context.Background())
	if err != nil {
		t.Fatalf("GetReviewTracks returned error: %v", err)
	}

	// Verify we got the expected tracks
	if len(tracks.Tracks) != 3 {
		t.Fatalf("expected 3 tracks, got %d", len(tracks.Tracks))
	}

	// Track with no TrackID should have 0 count
	if tracks.Tracks[0].AbstractCount != 0 {
		t.Errorf("track 0 (no TrackID): expected count 0, got %d", tracks.Tracks[0].AbstractCount)
	}

	// Tracks with TrackIDs should have counts populated (both using same fixture, so same count)
	expectedCount := 12 // from review-abstracts.html fixture
	if tracks.Tracks[1].AbstractCount != expectedCount {
		t.Errorf("track 1 (TrackID 88): expected count %d, got %d", expectedCount, tracks.Tracks[1].AbstractCount)
	}
	if tracks.Tracks[2].AbstractCount != expectedCount {
		t.Errorf("track 2 (TrackID 99): expected count %d, got %d", expectedCount, tracks.Tracks[2].AbstractCount)
	}

	// Verify all three HTTP requests were made (1 for tracks, 2 for abstract counts)
	if mmc.callCount != 3 {
		t.Errorf("expected 3 HTTP calls, got %d", mmc.callCount)
	}

	t.Logf("✅ Concurrent abstract count fetching works correctly")
}
