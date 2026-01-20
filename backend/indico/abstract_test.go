package indico

import (
	"encoding/json"
	"os"
	"testing"
)

// TestAbstractDataUnmarshal tests that AbstractData can unmarshal the real JSON data
func TestAbstractDataUnmarshal(t *testing.T) {
	// Read the test JSON file
	data, err := os.ReadFile("../../py/abstracts-37.json")
	if err != nil {
		t.Skipf("Test data file not found: %v", err)
		return
	}

	var response AbstractsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("Failed to unmarshal abstracts: %v", err)
	}

	if len(response.Abstracts) == 0 {
		t.Fatal("No abstracts found in response")
	}

	// Test the first abstract which should have reviews
	abstract := response.Abstracts[0]

	t.Logf("Abstract ID: %d, FriendlyID: %d", abstract.ID, abstract.FriendlyID)
	t.Logf("Title: %s", abstract.Title)
	t.Logf("State: %s", abstract.State)
	t.Logf("Number of reviews: %d", len(abstract.Reviews))
	t.Logf("Number of comments: %d", len(abstract.Comments))
	t.Logf("Number of persons: %d", len(abstract.Persons))

	// Verify reviews are parsed
	if len(abstract.Reviews) == 0 {
		t.Error("Expected reviews but got none")
	}

	// Check the first review
	if len(abstract.Reviews) > 0 {
		review := abstract.Reviews[0]
		t.Logf("Review ID: %d", review.ID)
		t.Logf("Review Comment: %s", review.Comment)
		t.Logf("Review ProposedAction: %s", review.ProposedAction)
		t.Logf("Review User: %s (%s)", review.User.FullName, review.User.Email)
		t.Logf("Review Track: %s (ID: %d)", review.Track.Title, review.Track.ID)
		t.Logf("Number of ratings: %d", len(review.Ratings))
		t.Logf("Number of proposed tracks: %d", len(review.ProposedTracks))

		if review.User.Affiliation != nil {
			t.Logf("Reviewer affiliation: %s (%s, %s)",
				review.User.Affiliation.Name,
				review.User.Affiliation.City,
				review.User.Affiliation.CountryName)
		}

		if len(review.Ratings) > 0 {
			t.Logf("First rating - Question: %d, Value: %d",
				review.Ratings[0].Question,
				review.Ratings[0].Value)
		}

		if review.ProposedContribType != nil {
			t.Logf("Proposed contrib type: %s (ID: %d)",
				review.ProposedContribType.Name,
				review.ProposedContribType.ID)
		}

		if review.ProposedRelatedAbstract != nil {
			t.Logf("Proposed related abstract: %s (ID: %d, FriendlyID: %d)",
				review.ProposedRelatedAbstract.Title,
				review.ProposedRelatedAbstract.ID,
				review.ProposedRelatedAbstract.FriendlyID)
		}
	}
}

// TestReviewFieldsParsing specifically tests review fields
func TestReviewFieldsParsing(t *testing.T) {
	// Read the test JSON file
	data, err := os.ReadFile("../../py/abstracts-37.json")
	if err != nil {
		t.Skipf("Test data file not found: %v", err)
		return
	}

	var response AbstractsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("Failed to unmarshal abstracts: %v", err)
	}

	foundAcceptReview := false
	foundRejectReview := false
	foundDuplicateReview := false
	foundChangeTracksReview := false

	for _, abstract := range response.Abstracts {
		for _, review := range abstract.Reviews {
			switch review.ProposedAction {
			case "accept":
				foundAcceptReview = true
			case "reject":
				foundRejectReview = true
			case "mark_as_duplicate":
				foundDuplicateReview = true
			case "change_tracks":
				foundChangeTracksReview = true
			}
		}
	}

	if !foundAcceptReview {
		t.Error("Did not find any 'accept' reviews")
	}
	if !foundRejectReview {
		t.Error("Did not find any 'reject' reviews")
	}
	if !foundDuplicateReview {
		t.Error("Did not find any 'mark_as_duplicate' reviews")
	}
	if !foundChangeTracksReview {
		t.Error("Did not find any 'change_tracks' reviews")
	}

	t.Logf("Successfully found all review action types")
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
