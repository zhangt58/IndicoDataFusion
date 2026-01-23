package indico

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// mockClient implements HTTPClient for tests.
type mockClient struct {
	resp *http.Response
	err  error
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	return m.resp, m.err
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
