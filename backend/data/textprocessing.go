package data

import (
	"strings"

	pluralize "github.com/gertd/go-pluralize"
)

var pluralClient = pluralize.NewClient()

// WordFrequency represents a word and its frequency count
type WordFrequency struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

// normalizeWord converts simple plural and possessive forms to a canonical singular
// It uses github.com/gertd/go-pluralize to detect and convert plurals when possible,
// and falls back to a few heuristic rules for cases not covered by the library.
func normalizeWord(word string) string {
	// remove English possessive endings (e.g., "author's" -> "author")
	if strings.HasSuffix(word, "'s") || strings.HasSuffix(word, "’s") {
		if len(word) > 2 {
			word = word[:len(word)-2]
		}
	}

	if word == "" {
		return word
	}

	// If the pluralize library thinks this is plural, convert to singular
	if pluralClient.IsPlural(word) {
		sing := pluralClient.Singular(word)
		if sing != "" {
			return sing
		}
	}
	return word
}

// GetWordFrequencies computes word frequencies from input text with stopword filtering
// New optional parameter: enablePluralNormalization ...bool
// If provided and true, plural forms will be normalized to singular before counting.
// When omitted (default) plural normalization is disabled.
// customExcludedWords: additional words to exclude from the result (case-insensitive)
func GetWordFrequencies(text string, minLength int, topN int, enablePluralNorm bool, customExcludedWords []string) []WordFrequency {

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

	// Add custom excluded words to stopwords map (case-insensitive)
	for _, word := range customExcludedWords {
		if word != "" {
			stopwords[strings.ToLower(strings.TrimSpace(word))] = true
		}
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
		if word == "" {
			continue
		}
		// normalize plural/possessive forms to singular only when option enabled
		n := word
		if enablePluralNorm {
			n = normalizeWord(word)
		}
		// Filter by length and stopwords (use normalized form for both checks)
		if len(n) >= minLength && !stopwords[n] {
			wordCount[n]++
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
