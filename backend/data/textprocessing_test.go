package data

import (
	"testing"
)

func TestGetWordFrequencies(t *testing.T) {
	text := "machine learning is a powerful learning technique. Machine learning algorithms can learn from data."

	frequencies := GetWordFrequencies(text, 3, 10, false)

	if len(frequencies) == 0 {
		t.Fatal("Expected word frequencies, got none")
	}

	// "learning" should appear 3 times
	// "machine" should appear 2 times
	foundLearning := false
	foundMachine := false

	for _, freq := range frequencies {
		if freq.Word == "learning" {
			foundLearning = true
			if freq.Count != 3 {
				t.Errorf("Expected 'learning' count to be 3, got %d", freq.Count)
			}
		}
		if freq.Word == "machine" {
			foundMachine = true
			if freq.Count != 2 {
				t.Errorf("Expected 'machine' count to be 2, got %d", freq.Count)
			}
		}
	}

	if !foundLearning {
		t.Error("Expected to find 'learning' in word frequencies")
	}
	if !foundMachine {
		t.Error("Expected to find 'machine' in word frequencies")
	}

	// Check that words are sorted by frequency
	if len(frequencies) >= 2 {
		if frequencies[0].Count < frequencies[1].Count {
			t.Error("Word frequencies should be sorted in descending order")
		}
	}

	t.Logf("Found %d unique words", len(frequencies))
	for i, freq := range frequencies {
		t.Logf("  %d. %s: %d", i+1, freq.Word, freq.Count)
	}
}

func TestGetWordFrequenciesStopwords(t *testing.T) {
	text := "the the the machine learning is the best"

	frequencies := GetWordFrequencies(text, 3, 10, false)

	// "the", "is" should be filtered out as stopwords
	for _, freq := range frequencies {
		if freq.Word == "the" || freq.Word == "is" {
			t.Errorf("Stopword '%s' should have been filtered out", freq.Word)
		}
	}

	// "machine", "learning", "best" should be present
	words := make(map[string]bool)
	for _, freq := range frequencies {
		words[freq.Word] = true
	}

	if !words["machine"] {
		t.Error("Expected 'machine' in results")
	}
	if !words["learning"] {
		t.Error("Expected 'learning' in results")
	}
	if !words["best"] {
		t.Error("Expected 'best' in results")
	}
}

func TestGetWordFrequenciesMinLength(t *testing.T) {
	text := "a ab abc abcd abcde"

	frequencies := GetWordFrequencies(text, 4, 10, false)

	// Only "abcd" and "abcde" should pass the min length of 4
	if len(frequencies) != 2 {
		t.Errorf("Expected 2 words with min length 4, got %d", len(frequencies))
	}

	for _, freq := range frequencies {
		if len(freq.Word) < 4 {
			t.Errorf("Word '%s' is shorter than min length 4", freq.Word)
		}
	}
}

func TestGetWordFrequenciesTopN(t *testing.T) {
	text := "one two three four five six seven eight nine ten"

	frequencies := GetWordFrequencies(text, 1, 5, false)

	if len(frequencies) > 5 {
		t.Errorf("Expected at most 5 words, got %d", len(frequencies))
	}
}

func TestGetWordFrequenciesEmpty(t *testing.T) {
	text := ""

	frequencies := GetWordFrequencies(text, 3, 10, false)

	if len(frequencies) != 0 {
		t.Errorf("Expected empty result for empty text, got %d words", len(frequencies))
	}
}

func TestGetWordFrequenciesOnlyStopwords(t *testing.T) {
	text := "the the and or but if then"

	frequencies := GetWordFrequencies(text, 2, 10, false)

	if len(frequencies) != 0 {
		t.Errorf("Expected empty result for text with only stopwords, got %d words", len(frequencies))
	}
}
