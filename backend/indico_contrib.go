package backend

// DateInfo represents date and time information in Indico
type DateInfo struct {
	Date string `json:"date"`
	Time string `json:"time"`
	TZ   string `json:"tz"`
}

// ContributionParticipation represents a person's participation in a contribution
type ContributionParticipation struct {
	Type        string `json:"_type"`
	Fossil      string `json:"_fossil"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	FullName    string `json:"fullName"`
	ID          string `json:"id"`
	Affiliation string `json:"affiliation"`
	EmailHash   string `json:"emailHash"`
	DBID        int    `json:"db_id"`
	PersonID    int    `json:"person_id"`
	Email       string `json:"email"`
}

// Folder represents a folder containing attachments
type Folder struct {
	Type          string        `json:"_type"`
	ID            int           `json:"id"`
	Title         *string       `json:"title"`
	Description   string        `json:"description"`
	Attachments   []interface{} `json:"attachments"` // Can be expanded if needed
	DefaultFolder bool          `json:"default_folder"`
	IsProtected   bool          `json:"is_protected"`
}

// ContributionData represents contribution information from Indico
type ContributionData struct {
	// Type information
	Type   string `json:"_type"`
	Fossil string `json:"_fossil"`

	// Basic Information
	ID          string `json:"id"`
	DBID        int    `json:"db_id"`
	FriendlyID  int    `json:"friendly_id"`
	Title       string `json:"title"`
	Description string `json:"description"`

	// Date and Time
	StartDate DateInfo `json:"startDate"`
	EndDate   DateInfo `json:"endDate"`
	Duration  int      `json:"duration"` // in minutes

	// Location
	Location     string `json:"location"`
	Room         string `json:"room"`
	RoomFullname string `json:"roomFullname"`

	// Type and Session
	ContribType string `json:"type"`
	Session     string `json:"session"`
	Track       string `json:"track"`

	// People
	Speakers       []ContributionParticipation `json:"speakers"`
	PrimaryAuthors []ContributionParticipation `json:"primaryauthors"`
	CoAuthors      []ContributionParticipation `json:"coauthors"`

	// Additional Information
	Keywords    []interface{} `json:"keywords"`
	References  []interface{} `json:"references"`
	BoardNumber string        `json:"board_number"`
	Code        string        `json:"code"`
	URL         string        `json:"url"`
	Note        interface{}   `json:"note"`

	// Materials and Folders
	Material []interface{} `json:"material"`
	Folders  []Folder      `json:"folders"`

	// Permissions
	Allowed map[string]interface{} `json:"allowed"`
}

// Conference represents the conference container for contributions
type Conference struct {
	Type          string             `json:"_type"`
	ID            string             `json:"id"`
	Title         string             `json:"title"`
	Description   string             `json:"description"`
	Contributions []ContributionData `json:"contributions"`
}

// ContributionsAPIResponse represents the top-level structure of contribs.json
type ContributionsAPIResponse struct {
	Count          int                    `json:"count"`
	AdditionalInfo map[string]interface{} `json:"additionalInfo"`
	Timestamp      int64                  `json:"ts"`
	URL            string                 `json:"url"`
	Results        []Conference           `json:"results"`
}
