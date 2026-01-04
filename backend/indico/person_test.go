package indico

import (
	"encoding/json"
	"testing"
)

func TestPersonUnmarshalJSON(t *testing.T) {
	jsonData := []byte(`{
		"affiliation": "Facility for Rare Isotope Beams, Michigan State University",
		"affiliation_link": {
			"city": "East Lansing",
			"country_code": "US",
			"country_name": "United States",
			"id": 852614,
			"name": "Facility for Rare Isotope Beams",
			"postcode": "",
			"street": ""
		},
		"author_type": "primary",
		"email": "test@example.com",
		"first_name": "John",
		"last_name": "Doe",
		"is_speaker": true,
		"person_id": 12345
	}`)

	var person Person
	err := json.Unmarshal(jsonData, &person)
	if err != nil {
		t.Fatalf("Failed to unmarshal Person: %v", err)
	}

	// Verify basic fields
	if person.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", person.Email)
	}
	if person.FirstName != "John" {
		t.Errorf("Expected first name 'John', got '%s'", person.FirstName)
	}
	if person.LastName != "Doe" {
		t.Errorf("Expected last name 'Doe', got '%s'", person.LastName)
	}

	// Verify affiliation was extracted from affiliation_link
	if person.Affiliation == nil {
		t.Fatal("Affiliation is nil, expected it to be populated from affiliation_link")
	}
	if person.Affiliation.ID != 852614 {
		t.Errorf("Expected affiliation ID 852614, got %d", person.Affiliation.ID)
	}
	if person.Affiliation.Name != "Facility for Rare Isotope Beams" {
		t.Errorf("Expected affiliation name 'Facility for Rare Isotope Beams', got '%s'", person.Affiliation.Name)
	}
	if person.Affiliation.City != "East Lansing" {
		t.Errorf("Expected city 'East Lansing', got '%s'", person.Affiliation.City)
	}
	if person.Affiliation.CountryCode != "US" {
		t.Errorf("Expected country code 'US', got '%s'", person.Affiliation.CountryCode)
	}
}

func TestJudgeUnmarshalJSON(t *testing.T) {
	jsonData := []byte(`{
		"affiliation": "Facility for Rare Isotope Beams, Michigan State University",
		"affiliation_meta": {
			"city": "East Lansing",
			"country_code": "US",
			"country_name": "United States",
			"id": 852614,
			"name": "Facility for Rare Isotope Beams",
			"postcode": "",
			"street": ""
		},
		"avatar_url": "/user/3804/picture-default/MzgwNA.aIZQtPyL",
		"email": "judge@example.com",
		"first_name": "Jane",
		"full_name": "Jane Smith",
		"id": 3804,
		"identifier": "User:3804:MzgwNA.ZqLrHTxE",
		"last_name": "Smith"
	}`)

	var judge Judge
	err := json.Unmarshal(jsonData, &judge)
	if err != nil {
		t.Fatalf("Failed to unmarshal Judge: %v", err)
	}

	// Verify basic fields
	if judge.Email != "judge@example.com" {
		t.Errorf("Expected email 'judge@example.com', got '%s'", judge.Email)
	}
	if judge.FirstName != "Jane" {
		t.Errorf("Expected first name 'Jane', got '%s'", judge.FirstName)
	}

	// Verify affiliation was extracted from affiliation_meta
	if judge.Affiliation == nil {
		t.Fatal("Affiliation is nil, expected it to be populated from affiliation_meta")
	}
	if judge.Affiliation.ID != 852614 {
		t.Errorf("Expected affiliation ID 852614, got %d", judge.Affiliation.ID)
	}
	if judge.Affiliation.Name != "Facility for Rare Isotope Beams" {
		t.Errorf("Expected affiliation name 'Facility for Rare Isotope Beams', got '%s'", judge.Affiliation.Name)
	}
}

func TestSubmitterUnmarshalJSON(t *testing.T) {
	jsonData := []byte(`{
		"affiliation": "Facility for Rare Isotope Beams, Michigan State University",
		"affiliation_meta": {
			"city": "East Lansing",
			"country_code": "US",
			"country_name": "United States",
			"id": 852614,
			"name": "Facility for Rare Isotope Beams",
			"postcode": "",
			"street": ""
		},
		"avatar_url": "/user/344/picture-default/MzQ0.5atl",
		"email": "submitter@example.com",
		"first_name": "Bob",
		"full_name": "Bob Johnson",
		"id": 344,
		"identifier": "User:344:MzQ0.Admz",
		"last_name": "Johnson"
	}`)

	var submitter Submitter
	err := json.Unmarshal(jsonData, &submitter)
	if err != nil {
		t.Fatalf("Failed to unmarshal Submitter: %v", err)
	}

	// Verify basic fields
	if submitter.Email != "submitter@example.com" {
		t.Errorf("Expected email 'submitter@example.com', got '%s'", submitter.Email)
	}
	if submitter.FirstName != "Bob" {
		t.Errorf("Expected first name 'Bob', got '%s'", submitter.FirstName)
	}

	// Verify affiliation was extracted from affiliation_meta
	if submitter.Affiliation == nil {
		t.Fatal("Affiliation is nil, expected it to be populated from affiliation_meta")
	}
	if submitter.Affiliation.ID != 852614 {
		t.Errorf("Expected affiliation ID 852614, got %d", submitter.Affiliation.ID)
	}
	if submitter.Affiliation.Name != "Facility for Rare Isotope Beams" {
		t.Errorf("Expected affiliation name 'Facility for Rare Isotope Beams', got '%s'", submitter.Affiliation.Name)
	}
}

func TestAffiliationUniqueID(t *testing.T) {
	// Test that multiple people from the same institution share the same affiliation ID
	jsonData1 := []byte(`{
		"affiliation_link": {"id": 852614, "name": "FRIB", "city": "", "country_code": "", "country_name": "", "postcode": "", "street": ""},
		"email": "person1@example.com",
		"author_type": "primary",
		"first_name": "Person",
		"last_name": "One",
		"is_speaker": true,
		"person_id": 1
	}`)

	jsonData2 := []byte(`{
		"affiliation_link": {"id": 852614, "name": "FRIB", "city": "", "country_code": "", "country_name": "", "postcode": "", "street": ""},
		"email": "person2@example.com",
		"author_type": "secondary",
		"first_name": "Person",
		"last_name": "Two",
		"is_speaker": false,
		"person_id": 2
	}`)

	var person1, person2 Person
	if err := json.Unmarshal(jsonData1, &person1); err != nil {
		t.Fatalf("Failed to unmarshal person1: %v", err)
	}
	if err := json.Unmarshal(jsonData2, &person2); err != nil {
		t.Fatalf("Failed to unmarshal person2: %v", err)
	}

	// Both should have the same affiliation ID
	if person1.Affiliation.ID != person2.Affiliation.ID {
		t.Errorf("Expected both persons to have the same affiliation ID, got %d and %d",
			person1.Affiliation.ID, person2.Affiliation.ID)
	}
	if person1.Affiliation.ID != 852614 {
		t.Errorf("Expected affiliation ID 852614, got %d", person1.Affiliation.ID)
	}
}
