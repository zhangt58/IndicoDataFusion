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
