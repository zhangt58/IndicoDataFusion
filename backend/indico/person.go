package indico

// Affiliation represents a structured affiliation object for people and judges
type Affiliation struct {
	City        string `json:"city"`
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Postcode    string `json:"postcode"`
	Street      string `json:"street"`
}

// Person represents a person in an abstract (author, speaker, etc.)
type Person struct {
	Affiliation *Affiliation `json:"affiliation"`
	Email       string       `json:"email"`
	AuthorType  string       `json:"author_type"`
	FirstName   string       `json:"first_name"`
	LastName    string       `json:"last_name"`
	IsSpeaker   bool         `json:"is_speaker"`
	PersonID    int          `json:"person_id"`
}

// Judge represents the judge information for an abstract
type Judge struct {
	Affiliation *Affiliation `json:"affiliation"`
	Email       string       `json:"email"`
	FirstName   string       `json:"first_name"`
	LastName    string       `json:"last_name"`
	FullName    string       `json:"full_name"`
	AvatarURL   string       `json:"avatar_url"`
	ID          int          `json:"id"`
	Identifier  string       `json:"identifier"`
}

// Submitter represents the person who submitted the abstract
// Affiliation is a structured Affiliation pointer.
type Submitter struct {
	Affiliation *Affiliation `json:"affiliation"`
	Email       string       `json:"email"`
	FirstName   string       `json:"first_name"`
	LastName    string       `json:"last_name"`
	FullName    string       `json:"full_name"`
	AvatarURL   string       `json:"avatar_url"`
	ID          int          `json:"id"`
	Identifier  string       `json:"identifier"`
}
