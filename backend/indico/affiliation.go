package indico

import (
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Affiliation represents a structured affiliation object for people and judges
type Affiliation struct {
	City        string `json:"city"`
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Postcode    string `json:"postcode"`
	Street      string `json:"street"`
	Raw         string `json:"raw,omitempty"` // original/raw affiliation string when present
}

// affiliationRegistry holds a map of affiliation name -> *Affiliation so we can
// deduplicate affiliation objects read from different JSON sources.
var affiliationRegistry = struct {
	mu sync.RWMutex
	m  map[string]*Affiliation
}{
	m: make(map[string]*Affiliation),
}

// normalizeAffiliationKey returns a normalized key for affiliation name.
// It performs Unicode NFKD decomposition, strips diacritics, removes punctuation/symbols,
// collapses whitespace, and lower-cases runes. This produces a robust key for deduping.
func normalizeAffiliationKey(name string) string {
	s := strings.TrimSpace(name)
	if s == "" {
		return ""
	}
	// Decompose (NFKD) so diacritics become separate combining marks we can drop
	t, _, err := transform.String(norm.NFKD, s)
	if err != nil {
		// if transform fails, fall back to simple cleanup
		t = s
	}
	var b strings.Builder
	lastSpace := false
	for _, r := range t {
		// drop combining marks (diacritics)
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		// drop punctuation and symbol characters
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			continue
		}
		// normalize spaces
		if unicode.IsSpace(r) {
			if !lastSpace {
				b.WriteByte(' ')
				lastSpace = true
			}
			continue
		}
		// write lower-cased rune (simple casefold)
		b.WriteRune(unicode.ToLower(r))
		lastSpace = false
	}
	return strings.TrimSpace(b.String())
}

// mergeAffiliation copies non-empty fields from src into dst when dst lacks them.
func mergeAffiliation(dst, src *Affiliation) {
	if dst == nil || src == nil {
		return
	}
	if dst.ID == 0 && src.ID != 0 {
		dst.ID = src.ID
	}
	if dst.Name == "" && src.Name != "" {
		dst.Name = src.Name
	}
	if dst.City == "" && src.City != "" {
		dst.City = src.City
	}
	if dst.CountryCode == "" && src.CountryCode != "" {
		dst.CountryCode = src.CountryCode
	}
	if dst.CountryName == "" && src.CountryName != "" {
		dst.CountryName = src.CountryName
	}
	if dst.Postcode == "" && src.Postcode != "" {
		dst.Postcode = src.Postcode
	}
	if dst.Street == "" && src.Street != "" {
		dst.Street = src.Street
	}
	// preserve raw if available
	if dst.Raw == "" && src.Raw != "" {
		dst.Raw = src.Raw
	}
}

// registerAffiliation registers an affiliation by name and returns a canonical pointer.
// If the affiliation has an empty name it is returned as-is (no registration).
func registerAffiliation(a *Affiliation) *Affiliation {
	if a == nil {
		return nil
	}
	name := strings.TrimSpace(a.Name)
	if name == "" {
		// if no structured name but Raw exists, use Raw as name for keying
		if a.Raw != "" {
			name = a.Raw
		} else {
			return a
		}
	}
	// ensure Raw is set when coming from plain string
	if a.Raw == "" {
		a.Raw = name
	}
	key := normalizeAffiliationKey(name)
	if key == "" {
		return a
	}

	// Fast path: read lock
	affiliationRegistry.mu.RLock()
	existing := affiliationRegistry.m[key]
	affiliationRegistry.mu.RUnlock()
	if existing != nil {
		// existing may be incomplete; merge under write lock to avoid races
		affiliationRegistry.mu.Lock()
		defer affiliationRegistry.mu.Unlock()
		// Re-check after acquiring write lock
		if e := affiliationRegistry.m[key]; e != nil {
			mergeAffiliation(e, a)
			return e
		}
		// if it disappeared (unlikely), fallthrough to insert
		affiliationRegistry.m[key] = a
		return a
	}

	// Slow path: acquire write lock and insert if still absent
	affiliationRegistry.mu.Lock()
	defer affiliationRegistry.mu.Unlock()
	if e := affiliationRegistry.m[key]; e != nil {
		mergeAffiliation(e, a)
		return e
	}
	// Insert provided affiliation as canonical
	affiliationRegistry.m[key] = a
	return a
}

// ClearAffiliationRegistry clears the global affiliation registry (useful for tests).
func ClearAffiliationRegistry() {
	affiliationRegistry.mu.Lock()
	defer affiliationRegistry.mu.Unlock()
	affiliationRegistry.m = make(map[string]*Affiliation)
}

// GetRegisteredAffiliation returns the canonical Affiliation pointer for the given name key
// or nil if none is registered. The input name will be normalized.
func GetRegisteredAffiliation(name string) *Affiliation {
	key := normalizeAffiliationKey(name)
	affiliationRegistry.mu.RLock()
	defer affiliationRegistry.mu.RUnlock()
	return affiliationRegistry.m[key]
}
