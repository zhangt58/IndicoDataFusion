package indico

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
	IsMyReview bool   `json:"is_my_review"`
	ReviewURL  string `json:"review_url"`

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
