package data

import (
	"strings"
)

// WordFrequency represents a word and its frequency count
type WordFrequency struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

// GetWordFrequencies computes word frequencies from input text with stopword filtering
func GetWordFrequencies(text string, minLength int, topN int) []WordFrequency {
	// Common English stopwords to filter out
	stopwords := map[string]bool{
		"the": true, "be": true, "to": true, "of": true, "and": true,
		"a": true, "in": true, "that": true, "have": true, "i": true,
		"it": true, "for": true, "not": true, "on": true, "with": true,
		"he": true, "as": true, "you": true, "do": true, "at": true,
		"this": true, "but": true, "his": true, "by": true, "from": true,
		"they": true, "we": true, "say": true, "her": true, "she": true,
		"or": true, "an": true, "will": true, "my": true, "one": true,
		"all": true, "would": true, "there": true, "their": true, "what": true,
		"so": true, "up": true, "out": true, "if": true, "about": true,
		"who": true, "get": true, "which": true, "go": true, "me": true,
		"when": true, "make": true, "can": true, "like": true, "time": true,
		"no": true, "just": true, "him": true, "know": true, "take": true,
		"people": true, "into": true, "year": true, "your": true, "good": true,
		"some": true, "could": true, "them": true, "see": true, "other": true,
		"than": true, "then": true, "now": true, "look": true, "only": true,
		"come": true, "its": true, "over": true, "think": true, "also": true,
		"back": true, "after": true, "use": true, "two": true, "how": true,
		"our": true, "work": true, "first": true, "well": true, "way": true,
		"even": true, "new": true, "want": true, "because": true, "any": true,
		"these": true, "give": true, "day": true, "most": true, "us": true,
		"is": true, "was": true, "are": true, "been": true, "has": true,
		"had": true, "were": true, "said": true, "did": true, "having": true,
	}

	// Count word frequencies
	wordCount := make(map[string]int)

	// Process text
	text = strings.ToLower(text)

	// Split into words (simple approach - split by non-alphanumeric)
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-')
	})

	for _, word := range words {
		word = strings.TrimSpace(word)
		// Filter by length and stopwords
		if len(word) >= minLength && !stopwords[word] {
			wordCount[word]++
		}
	}

	// Convert map to slice and sort by frequency
	frequencies := make([]WordFrequency, 0, len(wordCount))
	for word, count := range wordCount {
		frequencies = append(frequencies, WordFrequency{
			Word:  word,
			Count: count,
		})
	}

	// Sort by count (descending)
	for i := 0; i < len(frequencies); i++ {
		for j := i + 1; j < len(frequencies); j++ {
			if frequencies[j].Count > frequencies[i].Count {
				frequencies[i], frequencies[j] = frequencies[j], frequencies[i]
			}
		}
	}

	// Return top N
	if topN > 0 && topN < len(frequencies) {
		frequencies = frequencies[:topN]
	}

	return frequencies
}
