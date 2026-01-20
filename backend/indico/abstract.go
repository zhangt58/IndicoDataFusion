package indico

import "strings"

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
}

// AggRatings aggregates ratings from a single review by question ID.
// For numeric values, it treats s numbers;
// For boolean values, it treats true/yes as 1 and false/no as 0.
// Returns a map of question ID to aggregated value.
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

// AbstractsResponse represents the top-level structure of abstracts.json
type AbstractsResponse struct {
	Abstracts []AbstractData `json:"abstracts"`
	Questions []QuestionData `json:"questions"`
	Version   int            `json:"version,omitempty"`
}
