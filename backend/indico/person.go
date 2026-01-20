package indico

import "encoding/json"

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
	Affiliation *Affiliation `json:"affiliation,omitempty"` // populated from affiliation_link
	Email       string       `json:"email"`
	AuthorType  string       `json:"author_type"`
	FirstName   string       `json:"first_name"`
	LastName    string       `json:"last_name"`
	IsSpeaker   bool         `json:"is_speaker"`
	PersonID    int          `json:"person_id"`
}

// UnmarshalJSON custom unmarshaler for Person to extract affiliation from affiliation_link
func (p *Person) UnmarshalJSON(data []byte) error {
	type Alias Person
	aux := &struct {
		AffiliationString string       `json:"affiliation"` // ignore the string field
		AffiliationLink   *Affiliation `json:"affiliation_link"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	p.Affiliation = aux.AffiliationLink
	return nil
}

// Judge represents the judge information for an abstract
type Judge struct {
	Affiliation *Affiliation `json:"affiliation,omitempty"` // populated from affiliation_meta
	Email       string       `json:"email"`
	FirstName   string       `json:"first_name"`
	LastName    string       `json:"last_name"`
	FullName    string       `json:"full_name"`
	AvatarURL   string       `json:"avatar_url"`
	ID          int          `json:"id"`
	Identifier  string       `json:"identifier"`
}

// UnmarshalJSON custom unmarshaler for Judge to extract affiliation from affiliation_meta
func (j *Judge) UnmarshalJSON(data []byte) error {
	type Alias Judge
	aux := &struct {
		AffiliationString string       `json:"affiliation"` // ignore the string field
		AffiliationMeta   *Affiliation `json:"affiliation_meta"`
		*Alias
	}{
		Alias: (*Alias)(j),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	j.Affiliation = aux.AffiliationMeta
	return nil
}

// Submitter represents the person who submitted the abstract
// Affiliation is a structured Affiliation pointer.
type Submitter struct {
	Affiliation *Affiliation `json:"affiliation,omitempty"` // populated from affiliation_meta
	Email       string       `json:"email"`
	FirstName   string       `json:"first_name"`
	LastName    string       `json:"last_name"`
	FullName    string       `json:"full_name"`
	AvatarURL   string       `json:"avatar_url"`
	ID          int          `json:"id"`
	Identifier  string       `json:"identifier"`
}

// UnmarshalJSON custom unmarshaler for Submitter to extract affiliation from affiliation_meta
func (s *Submitter) UnmarshalJSON(data []byte) error {
	type Alias Submitter
	aux := &struct {
		AffiliationString string       `json:"affiliation"` // ignore the string field
		AffiliationMeta   *Affiliation `json:"affiliation_meta"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	s.Affiliation = aux.AffiliationMeta
	return nil
}

// Reviewer represents the user who submitted a review
// Similar structure to Judge/Submitter
type Reviewer struct {
	Affiliation *Affiliation `json:"affiliation,omitempty"` // populated from affiliation_meta
	Email       string       `json:"email"`
	FirstName   string       `json:"first_name"`
	LastName    string       `json:"last_name"`
	FullName    string       `json:"full_name"`
	AvatarURL   string       `json:"avatar_url"`
	ID          int          `json:"id"`
	Identifier  string       `json:"identifier"`
	Title       *string      `json:"title"` // pointer to handle null values
}

// UnmarshalJSON custom unmarshaler for Reviewer to extract affiliation from affiliation_meta
func (r *Reviewer) UnmarshalJSON(data []byte) error {
	type Alias Reviewer
	aux := &struct {
		AffiliationString string       `json:"affiliation"` // ignore the string field
		AffiliationMeta   *Affiliation `json:"affiliation_meta"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.Affiliation = aux.AffiliationMeta
	return nil
}
