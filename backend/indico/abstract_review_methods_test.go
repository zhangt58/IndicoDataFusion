package indico

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestReview_SubmitNewReview tests the SubmitNewReview method
func TestReview_SubmitNewReview(t *testing.T) {
	// Mock client
	mc := &mockClient{
		resp: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"success": true}`)),
			Header:     make(http.Header),
		},
	}

	client := &IndicoClient{
		BaseURL:   "https://indico.example.com",
		EventID:   123,
		Client:    mc,
		Timeout:   5 * time.Second,
		csrfToken: "test-csrf-token",
	}

	// Create abstract with questions map
	abstract := &AbstractData{
		ID: 456,
		Questions: map[int]*QuestionData{
			101: {ID: 101, Title: "First priority"},
			102: {ID: 102, Title: "Second priority"},
		},
	}

	// Create a new review
	contribType := &ContribType{ID: 42, Name: "Oral"}
	review := &Review{
		Track: Track{
			ID:    88,
			Code:  "MC1",
			Title: "Test Track",
		},
		ProposedAction:      "accept",
		ProposedContribType: contribType,
		Comment:             "Test comment",
	}

	// Submit new review using AbstractData.SubmitNewReview
	err := abstract.SubmitNewReview(
		context.Background(),
		client,
		review.Track.ID,
		1, // firstPriority
		0, // secondPriority
		review.ProposedAction,
		&contribType.ID,
		nil, // proposedTrackIDs
		nil, // proposedRelatedAbstractID
		"Great abstract!",
	)
	if err != nil {
		t.Fatalf("SubmitNewReview failed: %v", err)
	}
}

// TestReview_SubmitNewReview_MissingTrack tests error when track is missing
func TestReview_SubmitNewReview_MissingTrack(t *testing.T) {
	client := &IndicoClient{
		BaseURL:   "https://indico.example.com",
		EventID:   123,
		csrfToken: "test-csrf-token",
	}

	abstract := &AbstractData{
		ID: 456,
		Questions: map[int]*QuestionData{
			101: {ID: 101, Title: "First priority"},
			102: {ID: 102, Title: "Second priority"},
		},
	}

	review := &Review{
		ProposedAction: "accept",
	}

	// Call with trackID 0 to simulate missing track
	err := abstract.SubmitNewReview(
		context.Background(),
		client,
		0, // missing track
		1,
		0,
		review.ProposedAction,
		nil,
		nil,
		nil,
		"Test",
	)
	if err == nil {
		t.Fatal("Expected error for missing track, got nil")
	}
	if !strings.Contains(err.Error(), "track_id is required") {
		t.Errorf("Expected 'track_id is required' error, got: %v", err)
	}
}

// TestReview_UpdateReview tests the UpdateReview method
func TestReview_UpdateReview(t *testing.T) {
	// Mock client
	mc := &mockClient{
		resp: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"success": true}`)),
			Header:     make(http.Header),
		},
	}

	client := &IndicoClient{
		BaseURL:   "https://indico.example.com",
		EventID:   123,
		Client:    mc,
		Timeout:   5 * time.Second,
		csrfToken: "test-csrf-token",
	}

	abstract := &AbstractData{
		ID: 456,
		Questions: map[int]*QuestionData{
			101: {ID: 101, Title: "First priority"},
			102: {ID: 102, Title: "Second priority"},
		},
	}

	// Create an existing review
	contribType := &ContribType{ID: 50, Name: "Poster"}
	review := &Review{
		ID: 1234, // Existing review ID
		Track: Track{
			ID:    88,
			Code:  "MC1",
			Title: "Test Track",
		},
		ProposedAction:      "accept",
		ProposedContribType: contribType,
		Comment:             "Original comment",
	}

	// Update the review using AbstractData.UpdateReview
	err := abstract.UpdateReview(
		context.Background(),
		client,
		review.ID,
		review.Track.ID,
		1, // firstPriority
		1, // secondPriority
		review.ProposedAction,
		&contribType.ID,
		nil, // proposedTrackIDs
		nil, // proposedRelatedAbstractID
		"Updated comment!",
	)
	if err != nil {
		t.Fatalf("UpdateReview failed: %v", err)
	}
}

// TestReview_UpdateReview_MissingReviewID tests error when review ID is missing
func TestReview_UpdateReview_MissingReviewID(t *testing.T) {
	client := &IndicoClient{
		BaseURL:   "https://indico.example.com",
		EventID:   123,
		csrfToken: "test-csrf-token",
	}

	abstract := &AbstractData{
		ID: 456,
		Questions: map[int]*QuestionData{
			101: {ID: 101, Title: "First priority"},
			102: {ID: 102, Title: "Second priority"},
		},
	}

	review := &Review{
		Track: Track{
			ID:    88,
			Code:  "MC1",
			Title: "Test Track",
		},
		ProposedAction: "accept",
	}

	// Call UpdateReview with reviewID 0 to simulate missing review ID
	err := abstract.UpdateReview(
		context.Background(),
		client,
		0, // missing review ID
		review.Track.ID,
		1,
		0,
		review.ProposedAction,
		nil,
		nil,
		nil,
		"Test",
	)
	if err == nil {
		t.Fatal("Expected error for missing review ID, got nil")
	}
	if !strings.Contains(err.Error(), "review ID is required") {
		t.Errorf("Expected 'review ID is required' error, got: %v", err)
	}
}

// TestReview_SubmitNewReview_WithProposedTracks tests submitting with proposed tracks
func TestReview_SubmitNewReview_WithProposedTracks(t *testing.T) {
	mc := &mockClient{
		resp: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"success": true}`)),
			Header:     make(http.Header),
		},
	}

	client := &IndicoClient{
		BaseURL:   "https://indico.example.com",
		EventID:   123,
		Client:    mc,
		Timeout:   5 * time.Second,
		csrfToken: "test-csrf-token",
	}

	abstract := &AbstractData{
		ID: 456,
		Questions: map[int]*QuestionData{
			101: {ID: 101, Title: "First priority"},
			102: {ID: 102, Title: "Second priority"},
		},
	}

	review := &Review{
		Track: Track{
			ID:    88,
			Code:  "MC1",
			Title: "Test Track",
		},
		// fix action string to match server-side expectation
		ProposedAction: "change_tracks",
		ProposedTracks: []Track{
			{ID: 99, Code: "MC2", Title: "Track 2"},
			{ID: 100, Code: "MC3", Title: "Track 3"},
		},
	}

	err := abstract.SubmitNewReview(
		context.Background(),
		client,
		review.Track.ID,
		1,
		0,
		review.ProposedAction,
		nil,
		[]int{99, 100},
		nil,
		"Better fit elsewhere",
	)
	if err != nil {
		t.Fatalf("SubmitNewReview with proposed tracks failed: %v", err)
	}
}

// TestReview_SubmitNewReview_WithRelatedAbstract tests submitting with related abstract
func TestReview_SubmitNewReview_WithRelatedAbstract(t *testing.T) {
	mc := &mockClient{
		resp: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"success": true}`)),
			Header:     make(http.Header),
		},
	}

	client := &IndicoClient{
		BaseURL:   "https://indico.example.com",
		EventID:   123,
		Client:    mc,
		Timeout:   5 * time.Second,
		csrfToken: "test-csrf-token",
	}

	abstract := &AbstractData{
		ID: 456,
		Questions: map[int]*QuestionData{
			101: {ID: 101, Title: "First priority"},
			102: {ID: 102, Title: "Second priority"},
		},
	}

	review := &Review{
		Track: Track{
			ID:    88,
			Code:  "MC1",
			Title: "Test Track",
		},
		ProposedAction: "mark_as_duplicate",
		ProposedRelatedAbstract: &RelatedAbstract{
			ID:         789,
			FriendlyID: 10,
			Title:      "Related Abstract",
		},
	}

	// supply a valid track ID; SubmitAbstractReview requires track_id for new reviews
	err := abstract.SubmitNewReview(
		context.Background(),
		client,
		review.Track.ID, // use non-zero track id
		0,
		0,
		review.ProposedAction,
		nil,
		nil,
		&review.ProposedRelatedAbstract.ID,
		"Duplicate of #10",
	)
	if err != nil {
		t.Fatalf("SubmitNewReview with related abstract failed: %v", err)
	}
}
