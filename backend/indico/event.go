package indico

import (
	"encoding/json"
	"fmt"
)

// FlexibleID can hold an ID that is either an int or a string. It implements
// custom JSON unmarshalling so fields that sometimes come back as numbers and
// sometimes as strings are handled transparently.
type FlexibleID struct {
	IsInt bool
	Int   int
	Str   string
}

// UnmarshalJSON implements json.Unmarshaler and accepts either a JSON number
// or a JSON string.
func (f *FlexibleID) UnmarshalJSON(b []byte) error {
	// try int first
	var i int
	if err := json.Unmarshal(b, &i); err == nil {
		f.IsInt = true
		f.Int = i
		f.Str = ""
		return nil
	}
	// try string
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		f.IsInt = false
		f.Int = 0
		f.Str = s
		return nil
	}
	return fmt.Errorf("FlexibleID: value is not int or string: %s", string(b))
}

// MarshalJSON implements json.Marshaler to emit the original type when
// possible (int if IsInt, otherwise string).
func (f FlexibleID) MarshalJSON() ([]byte, error) {
	if f.IsInt {
		return json.Marshal(f.Int)
	}
	return json.Marshal(f.Str)
}

// String returns a string representation of the ID regardless of underlying type.
func (f FlexibleID) String() string {
	if f.IsInt {
		return fmt.Sprintf("%d", f.Int)
	}
	return f.Str
}

// Event holds the essential information about an Indico event.
type Event struct {
	Type         string      `json:"_type,omitempty"`
	ID           string      `json:"id"`
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	StartDate    DateInfo    `json:"startDate,omitempty"`
	EndDate      DateInfo    `json:"endDate,omitempty"`
	Timezone     string      `json:"timezone,omitempty"`
	Location     string      `json:"location,omitempty"`
	Address      string      `json:"address,omitempty"`
	Room         string      `json:"room,omitempty"`
	RoomFullname string      `json:"roomFullname,omitempty"`
	RoomMapURL   string      `json:"roomMapURL,omitempty"`
	Category     string      `json:"category,omitempty"`
	CategoryID   int         `json:"categoryId,omitempty"`
	EventType    string      `json:"type,omitempty"`
	URL          string      `json:"url,omitempty"`
	Note         interface{} `json:"note,omitempty"`

	// Creation and creator info
	CreationDate DateInfo      `json:"creationDate,omitempty"`
	Creator      *EventCreator `json:"creator,omitempty"`

	// Protection and visibility
	HasAnyProtection bool                   `json:"hasAnyProtection,omitempty"`
	Visibility       *EventVisibility       `json:"visibility,omitempty"`
	Allowed          map[string]interface{} `json:"allowed,omitempty"`

	// People
	Chairs []EventChair `json:"chairs,omitempty"`

	// Attachments and materials
	Folders  []Folder   `json:"folders,omitempty"`
	Material []Material `json:"material,omitempty"`

	// Additional metadata
	Keywords   []interface{} `json:"keywords,omitempty"`
	References []interface{} `json:"references,omitempty"`
	Organizer  string        `json:"organizer,omitempty"`
	Language   *string       `json:"language,omitempty"`
	Label      *string       `json:"label,omitempty"`
}

// EventCreator represents the person who created the event
type EventCreator struct {
	Type        string `json:"_type,omitempty"`
	Fossil      string `json:"_fossil,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	FullName    string `json:"fullName,omitempty"`
	ID          string `json:"id,omitempty"`
	Affiliation string `json:"affiliation,omitempty"`
	EmailHash   string `json:"emailHash,omitempty"`
	Email       string `json:"email,omitempty"`
}

// EventChair represents a conference chair
type EventChair struct {
	Type        string `json:"_type,omitempty"`
	Fossil      string `json:"_fossil,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	FullName    string `json:"fullName,omitempty"`
	ID          string `json:"id,omitempty"`
	Affiliation string `json:"affiliation,omitempty"`
	EmailHash   string `json:"emailHash,omitempty"`
	Email       string `json:"email,omitempty"`
	DBID        int    `json:"db_id,omitempty"`
	PersonID    int    `json:"person_id,omitempty"`
}

// EventVisibility represents visibility settings
type EventVisibility struct {
	ID   FlexibleID `json:"id,omitempty"`
	Name string     `json:"name,omitempty"`
}

// Material represents deprecated material structure (for backward compatibility)
type Material struct {
	Type       string             `json:"_type,omitempty"`
	Fossil     string             `json:"_fossil,omitempty"`
	Deprecated bool               `json:"_deprecated,omitempty"`
	Title      string             `json:"title,omitempty"`
	ID         string             `json:"id,omitempty"`
	Resources  []MaterialResource `json:"resources,omitempty"`
}

// MaterialResource represents a resource within a material
type MaterialResource struct {
	Name       string `json:"name,omitempty"`
	Deprecated bool   `json:"_deprecated,omitempty"`
	Type       string `json:"_type,omitempty"`
	Fossil     string `json:"_fossil,omitempty"`
	ID         string `json:"id,omitempty"`
	FileName   string `json:"fileName,omitempty"`
	URL        string `json:"url,omitempty"`
}

type EventAPIResponse struct {
	Results []Event `json:"results"`
	Count   int     `json:"count"`
	Url     string  `json:"url"`
}
