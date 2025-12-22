package indico

// Person represents a person in an abstract (author, speaker, etc.)
type Person struct {
	Affiliation string `json:"affiliation"`
	Email       string `json:"email"`
	AuthorType  string `json:"author_type"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	IsSpeaker   bool   `json:"is_speaker"`
	PersonID    int    `json:"person_id"`
}

// Judge represents the judge information for an abstract
type Judge struct {
	Affiliation string `json:"affiliation"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	FullName    string `json:"full_name"`
	AvatarURL   string `json:"avatar_url"`
	ID          int    `json:"id"`
	Identifier  string `json:"identifier"`
}

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

// Submitter represents the person who submitted the abstract
type Submitter struct {
	Affiliation string `json:"affiliation"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	FullName    string `json:"full_name"`
	AvatarURL   string `json:"avatar_url"`
	ID          int    `json:"id"`
	Identifier  string `json:"identifier"`
}

// Review represents a review of the abstract
type Review struct {
	// Add fields as needed when review structure is known
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

// AbstractsResponse represents the top-level structure of abstracts.json
type AbstractsResponse struct {
	Abstracts []AbstractData `json:"abstracts"`
}
