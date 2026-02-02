package indico

import (
	"encoding/json"
	"testing"
)

func TestEventJSONDeserialization(t *testing.T) {
	// Sample JSON from py/info.json
	jsonData := `{
		"_type": "Conference",
		"id": "82",
		"title": "HIAT2025 - the 16th International Conference on Heavy Ion Accelerator Technology",
		"description": "<p>Test conference</p>",
		"startDate": {"date": "2025-06-22", "time": "15:00:00", "tz": "Europe/Zurich"},
		"endDate": {"date": "2025-06-27", "time": "23:30:00", "tz": "Europe/Zurich"},
		"timezone": "US/Eastern",
		"location": "Kellogg Hotel and Conference Center",
		"address": "Michigan State University, East Lansing, MI 48824, USA",
		"category": "HIAT",
		"categoryId": 12,
		"type": "conference",
		"url": "https://indico.jacow.org/event/82/",
		"hasAnyProtection": false,
		"folders": [{
			"_type": "folder",
			"id": 392,
			"title": null,
			"description": "",
			"attachments": [{
				"_type": "attachment",
				"id": 5059,
				"download_url": "https://indico.jacow.org/event/82/attachments/392/5059/HIAT2025_abstracts.pdf",
				"title": "HIAT2025_abstracts.pdf",
				"description": "",
				"modified_dt": "2025-06-22T01:00:58.850828+00:00",
				"type": "file",
				"is_protected": false,
				"filename": "HIAT2025_abstracts.pdf",
				"content_type": "application/pdf",
				"size": 822910,
				"checksum": "54bedbaffd7fad8949d067d3ed03db61"
			}],
			"default_folder": true,
			"is_protected": false
		}],
		"chairs": [{
			"_type": "ConferenceChair",
			"first_name": "Peter",
			"last_name": "Ostroumov",
			"fullName": "Ostroumov, Peter",
			"id": "34",
			"affiliation": "Facility for Rare Isotope Beams, Michigan State University",
			"emailHash": "877b19d8ffba5e0694ab351384ed1183",
			"db_id": 34,
			"person_id": 14033,
			"email": "ostroumov@frib.msu.edu"
		}],
		"creator": {
			"_type": "Avatar",
			"first_name": "Ivan",
			"last_name": "Andrian",
			"fullName": "Andrian, Ivan",
			"id": "9",
			"affiliation": "Elettra-Sincrotrone Trieste S.C.p.A.",
			"emailHash": "2fa5099820cd6b82339193b5e49c008b",
			"email": "ivan.andrian@elettra.eu"
		}
	}`

	var event Event
	err := json.Unmarshal([]byte(jsonData), &event)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify basic fields
	if event.ID != "82" {
		t.Errorf("Expected ID '82', got '%s'", event.ID)
	}
	if event.Title != "HIAT2025 - the 16th International Conference on Heavy Ion Accelerator Technology" {
		t.Errorf("Unexpected title: %s", event.Title)
	}
	if event.Timezone != "US/Eastern" {
		t.Errorf("Expected timezone 'US/Eastern', got '%s'", event.Timezone)
	}

	// Verify folders
	if len(event.Folders) != 1 {
		t.Fatalf("Expected 1 folder, got %d", len(event.Folders))
	}
	folder := event.Folders[0]
	if folder.ID != 392 {
		t.Errorf("Expected folder ID 392, got %d", folder.ID)
	}
	if !folder.DefaultFolder {
		t.Error("Expected default_folder to be true")
	}

	// Verify attachments
	if len(folder.Attachments) != 1 {
		t.Fatalf("Expected 1 attachment, got %d", len(folder.Attachments))
	}
	attachment := folder.Attachments[0]
	if attachment.ID != 5059 {
		t.Errorf("Expected attachment ID 5059, got %d", attachment.ID)
	}
	if attachment.DownloadURL != "https://indico.jacow.org/event/82/attachments/392/5059/HIAT2025_abstracts.pdf" {
		t.Errorf("Unexpected download URL: %s", attachment.DownloadURL)
	}
	if attachment.Title != "HIAT2025_abstracts.pdf" {
		t.Errorf("Expected title 'HIAT2025_abstracts.pdf', got '%s'", attachment.Title)
	}
	if attachment.Size != 822910 {
		t.Errorf("Expected size 822910, got %d", attachment.Size)
	}
	if attachment.ContentType != "application/pdf" {
		t.Errorf("Expected content_type 'application/pdf', got '%s'", attachment.ContentType)
	}

	// Verify creator
	if event.Creator == nil {
		t.Fatal("Expected creator to be present")
	}
	if event.Creator.FirstName != "Ivan" {
		t.Errorf("Expected creator first name 'Ivan', got '%s'", event.Creator.FirstName)
	}

	// Verify chairs
	if len(event.Chairs) != 1 {
		t.Fatalf("Expected 1 chair, got %d", len(event.Chairs))
	}
	chair := event.Chairs[0]
	if chair.FirstName != "Peter" {
		t.Errorf("Expected chair first name 'Peter', got '%s'", chair.FirstName)
	}
	if chair.LastName != "Ostroumov" {
		t.Errorf("Expected chair last name 'Ostroumov', got '%s'", chair.LastName)
	}
	if chair.Affiliation != "Facility for Rare Isotope Beams, Michigan State University" {
		t.Errorf("Unexpected chair affiliation: %s", chair.Affiliation)
	}
}

func TestAttachmentTypeDeserialization(t *testing.T) {
	jsonData := `{
		"_type": "attachment",
		"id": 5059,
		"download_url": "https://example.com/file.pdf",
		"title": "Test File",
		"description": "Test description",
		"modified_dt": "2025-06-22T01:00:58.850828+00:00",
		"type": "file",
		"is_protected": true,
		"filename": "test.pdf",
		"content_type": "application/pdf",
		"size": 1024000,
		"checksum": "abc123"
	}`

	var attachment Attachment
	err := json.Unmarshal([]byte(jsonData), &attachment)
	if err != nil {
		t.Fatalf("Failed to unmarshal attachment JSON: %v", err)
	}

	if attachment.ID != 5059 {
		t.Errorf("Expected ID 5059, got %d", attachment.ID)
	}
	if !attachment.IsProtected {
		t.Error("Expected is_protected to be true")
	}
	if attachment.Size != 1024000 {
		t.Errorf("Expected size 1024000, got %d", attachment.Size)
	}
}

func TestEmptyFoldersAndAttachments(t *testing.T) {
	// Test event with no folders
	jsonData := `{
		"id": "1",
		"title": "Test Event",
		"description": "Test"
	}`

	var event Event
	err := json.Unmarshal([]byte(jsonData), &event)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if event.Folders != nil && len(event.Folders) > 0 {
		t.Error("Expected no folders for event without folders field")
	}
}
