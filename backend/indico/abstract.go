package indico

import (
	"context"
	"fmt"
	"strings"
)

// Track represents a conference track
type Track struct {
	Code  string `json:"code"`
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// ContribType represents the contribution type
type ContribType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// QuestionData represents a review question from the abstracts response
type QuestionData struct {
	ID       int    `json:"id"`
	NoScore  bool   `json:"no_score"`
	Position int    `json:"position"`
	Title    string `json:"title"`
}

// Rating represents a rating given in a review
type Rating struct {
	Question        int           `json:"question"`
	Value           interface{}   `json:"value"`                      // can be int or bool depending on question type
	QuestionDetails *QuestionData `json:"question_details,omitempty"` // Expanded question info
}

// RelatedAbstract represents a reference to another abstract
type RelatedAbstract struct {
	FriendlyID int    `json:"friendly_id"`
	ID         int    `json:"id"`
	Title      string `json:"title"`
}

// Review represents a review of the abstract
type Review struct {
	Comment                 string           `json:"comment"`
	CreatedDT               string           `json:"created_dt"`
	ID                      int              `json:"id"`
	ModifiedDT              *string          `json:"modified_dt"` // pointer to handle null values
	ProposedAction          string           `json:"proposed_action"`
	ProposedContribType     *ContribType     `json:"proposed_contrib_type"`     // pointer to handle null values
	ProposedRelatedAbstract *RelatedAbstract `json:"proposed_related_abstract"` // pointer to handle null values
	ProposedTracks          []Track          `json:"proposed_tracks"`
	Ratings                 []Rating         `json:"ratings"`
	Track                   Track            `json:"track"`
	User                    Reviewer         `json:"user"`
}

// CustomField represents custom fields in the abstract
type CustomField struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// AbstractData represents abstract information from Indico
type AbstractData struct {
	// Basic Information
	ID         int    `json:"id"`
	FriendlyID int    `json:"friendly_id"`
	State      string `json:"state"`
	Title      string `json:"title"`
	Content    string `json:"content"`

	// the original URL of the abstract in Indico
	IndicoURL string `json:"indico_url"`

	// Scoring and Judgment
	Score           *float64 `json:"score"` // pointer to handle null values
	Judge           *Judge   `json:"judge"` // pointer to handle null values
	JudgmentComment string   `json:"judgment_comment"`
	JudgmentDT      string   `json:"judgment_dt"`

	// People
	Persons   []Person   `json:"persons"`
	Submitter *Submitter `json:"submitter"`

	// Tracks and Types
	AcceptedTrack        *Track       `json:"accepted_track"`
	AcceptedContribType  *ContribType `json:"accepted_contrib_type"`
	SubmittedContribType *ContribType `json:"submitted_contrib_type"`
	ReviewedForTracks    []Track      `json:"reviewed_for_tracks"`
	SubmittedForTracks   []Track      `json:"submitted_for_tracks"`

	// Reviews and Comments
	Reviews      []Review      `json:"reviews"`
	Comments     []interface{} `json:"comments"` // generic for now
	CustomFields []CustomField `json:"custom_fields"`

	// Metadata
	SubmittedDT       string        `json:"submitted_dt"`
	ModifiedDT        string        `json:"modified_dt"`
	ModifiedBy        *Submitter    `json:"modified_by"`
	SubmissionComment string        `json:"submission_comment"`
	DuplicateOf       *int          `json:"duplicate_of"`
	MergedInto        *int          `json:"merged_into"`
	Files             []interface{} `json:"files"` // generic for now

	// if this abstract to be reviewed by the current user
	IsMyReview bool `json:"is_my_review"`
	// review URL is the original Indico URL to review this abstract
	// review is available only if the current user is a reviewer for this abstract
	ReviewURL string `json:"review_url"`

	// MyReview is the current user's review for this abstract (nil if not reviewed yet)
	MyReview *Review `json:"my_review,omitempty"`

	// Questions is a pointer to the shared question map for this event
	// Key: question ID, Value: question details
	Questions map[int]*QuestionData `json:"questions,omitempty"`

	// Aggregated Ratings (computed fields for frontend convenience)
	FirstPriority  float64 `json:"first_priority"`
	SecondPriority float64 `json:"second_priority"`
}

// AbstractsResponse represents the top-level structure of abstracts.json
type AbstractsResponse struct {
	Abstracts []AbstractData `json:"abstracts"`
	Questions []QuestionData `json:"questions"`
	Version   int            `json:"version,omitempty"`
}

// FindQuestionIDByTitle searches for a question ID by its title (case-insensitive).
// Returns the question ID and true if found, 0 and false otherwise.
func (a *AbstractData) FindQuestionIDByTitle(title string) (int, bool) {
	if a.Questions == nil {
		return 0, false
	}

	lowerTitle := strings.ToLower(title)
	for id, q := range a.Questions {
		if strings.ToLower(q.Title) == lowerTitle {
			return id, true
		}
	}
	return 0, false
}

// GetFirstPriorityQuestionID returns the question ID for "First priority".
// Returns 0 if not found.
func (a *AbstractData) GetFirstPriorityQuestionID() int {
	if id, ok := a.FindQuestionIDByTitle("First priority"); ok {
		return id
	}
	return 0
}

// GetSecondPriorityQuestionID returns the question ID for "Second priority".
// Returns 0 if not found.
func (a *AbstractData) GetSecondPriorityQuestionID() int {
	if id, ok := a.FindQuestionIDByTitle("Second priority"); ok {
		return id
	}
	return 0
}

// buildReviewRequest is a helper method that builds a ReviewSubmissionRequest from the review data.
// It consolidates the shared logic between SubmitNewReview and UpdateReview.
func (r *Review) buildReviewRequest(firstPriorityQuestionID, firstPriorityValue,
	secondPriorityQuestionID, secondPriorityValue int,
	comment string, reviewID *int) *ReviewSubmissionRequest {
	// Build question ratings map for first_priority and second_priority
	questionRatings := map[int]int{}

	if firstPriorityQuestionID > 0 {
		questionRatings[firstPriorityQuestionID] = firstPriorityValue
	}
	if secondPriorityQuestionID > 0 {
		questionRatings[secondPriorityQuestionID] = secondPriorityValue
	}

	// Determine proposed contribution type
	var contribTypeID *int
	if r.ProposedContribType != nil {
		contribTypeID = &r.ProposedContribType.ID
	}

	// Build proposed tracks
	var proposedTracks []int
	for _, track := range r.ProposedTracks {
		proposedTracks = append(proposedTracks, track.ID)
	}

	// Build proposed related abstract
	var relatedAbstractID *int
	if r.ProposedRelatedAbstract != nil {
		relatedAbstractID = &r.ProposedRelatedAbstract.ID
	}

	return &ReviewSubmissionRequest{
		ReviewID:                 reviewID,
		TrackID:                  r.Track.ID,
		QuestionRatings:          questionRatings,
		ProposedAction:           r.ProposedAction,
		ProposedContributionType: contribTypeID,
		ProposedTracks:           proposedTracks,
		ProposedRelatedAbstract:  relatedAbstractID,
		Comment:                  comment,
	}
}

// SubmitNewReview creates a new review for an abstract using the SubmitAbstractReview API.
// Parameters:
//   - ctx: context for the request
//   - client: IndicoClient to use for submission
//   - abstract: the AbstractData containing question information
//   - firstPriorityValue: rating value (0 or 1) for first priority
//   - secondPriorityValue: rating value (0 or 1) for second priority
//   - comment: review comment
//
// Returns error if submission fails.
func (r *Review) SubmitNewReview(ctx context.Context, client *IndicoClient, abstract *AbstractData,
	firstPriorityValue, secondPriorityValue int, comment string) error {

	if r.Track.ID == 0 {
		return fmt.Errorf("track ID is required in review")
	}

	// Get question IDs from abstract
	firstPriorityQuestionID := abstract.GetFirstPriorityQuestionID()
	secondPriorityQuestionID := abstract.GetSecondPriorityQuestionID()

	req := r.buildReviewRequest(firstPriorityQuestionID, firstPriorityValue,
		secondPriorityQuestionID, secondPriorityValue, comment, nil)
	return client.SubmitAbstractReview(ctx, abstract.ID, req)
}

// UpdateReview updates an existing review using the SubmitAbstractReview API.
// Parameters:
//   - ctx: context for the request
//   - client: IndicoClient to use for submission
//   - abstract: the AbstractData containing question information
//   - firstPriorityValue: rating value (0 or 1) for first priority
//   - secondPriorityValue: rating value (0 or 1) for second priority
//   - comment: review comment
//
// Returns error if submission fails.
func (r *Review) UpdateReview(ctx context.Context, client *IndicoClient, abstract *AbstractData,
	firstPriorityValue, secondPriorityValue int, comment string) error {

	if r.ID == 0 {
		return fmt.Errorf("review ID is required for update - use SubmitNewReview for new reviews")
	}

	// Get question IDs from abstract
	firstPriorityQuestionID := abstract.GetFirstPriorityQuestionID()
	secondPriorityQuestionID := abstract.GetSecondPriorityQuestionID()

	reviewID := r.ID
	req := r.buildReviewRequest(firstPriorityQuestionID, firstPriorityValue,
		secondPriorityQuestionID, secondPriorityValue, comment, &reviewID)
	return client.SubmitAbstractReview(ctx, abstract.ID, req)
}
