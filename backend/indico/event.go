package indico

// Event holds the essential information about an Indico event.
type Event struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	StartDate   DateInfo `json:"startDate,omitempty"`
	EndDate     DateInfo `json:"endDate,omitempty"`
	Location    string   `json:"location,omitempty"`
	Address     string   `json:"address,omitempty"`
	Category    string   `json:"category,omitempty"`
}

type EventAPIResponse struct {
	Results []Event `json:"results"`
	Count   int     `json:"count"`
	Url     string  `json:"url"`
}
