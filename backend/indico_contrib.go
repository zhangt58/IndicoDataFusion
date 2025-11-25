package backend

import (
	"encoding/json"
	"fmt"
	"os"
)

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

// GetContributionData reads contribution data from py/contribs.json file
func (c *IndicoClient) GetContributionData() ([]ContributionData, error) {
	// Read the contribs.json file from py directory
	data, err := os.ReadFile("py/contribs.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read py/contribs.json: %w", err)
	}

	// Parse the JSON data
	var response ContributionsAPIResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse py/contribs.json: %w", err)
	}

	// Extract contributions from results
	if len(response.Results) == 0 {
		return nil, fmt.Errorf("no results found in contribs.json")
	}

	// The contributions are under results[0].contributions
	contributions := response.Results[0].Contributions

	return contributions, nil
}

// GetContributionByID retrieves a specific contribution by its ID
func (c *IndicoClient) GetContributionByID(id string) (*ContributionData, error) {
	contributions, err := c.GetContributionData()
	if err != nil {
		return nil, err
	}

	for _, contrib := range contributions {
		if contrib.ID == id {
			return &contrib, nil
		}
	}

	return nil, fmt.Errorf("contribution with ID %s not found", id)
}

// GetContributionsBySession retrieves all contributions for a specific session
func (c *IndicoClient) GetContributionsBySession(session string) ([]ContributionData, error) {
	allContribs, err := c.GetContributionData()
	if err != nil {
		return nil, err
	}

	var sessionContribs []ContributionData
	for _, contrib := range allContribs {
		if contrib.Session == session {
			sessionContribs = append(sessionContribs, contrib)
		}
	}

	return sessionContribs, nil
}

// GetContributionsByTrack retrieves all contributions for a specific track
func (c *IndicoClient) GetContributionsByTrack(track string) ([]ContributionData, error) {
	allContribs, err := c.GetContributionData()
	if err != nil {
		return nil, err
	}

	var trackContribs []ContributionData
	for _, contrib := range allContribs {
		if contrib.Track == track {
			trackContribs = append(trackContribs, contrib)
		}
	}

	return trackContribs, nil
}

// GetContributionsByType retrieves all contributions of a specific type
func (c *IndicoClient) GetContributionsByType(contribType string) ([]ContributionData, error) {
	allContribs, err := c.GetContributionData()
	if err != nil {
		return nil, err
	}

	var typeContribs []ContributionData
	for _, contrib := range allContribs {
		if contrib.ContribType == contribType {
			typeContribs = append(typeContribs, contrib)
		}
	}

	return typeContribs, nil
}
