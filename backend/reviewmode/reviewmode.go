package reviewmode

import "IndicoDataFusion/backend/indico"

// RedactionConfig defines which fields of AbstractData should be redacted (zeroed/nil'd)
// before exporting abstracts to a file. By default all sensitive/scoring fields are redacted.
type RedactionConfig struct {
	RedactScore           bool `yaml:"redact_score" json:"redactScore"`
	RedactJudge           bool `yaml:"redact_judge" json:"redactJudge"`
	RedactJudgmentComment bool `yaml:"redact_judgment_comment" json:"redactJudgmentComment"`
	RedactJudgmentDT      bool `yaml:"redact_judgment_dt" json:"redactJudgmentDT"`
	RedactSubmitter       bool `yaml:"redact_submitter" json:"redactSubmitter"`
	RedactReviews         bool `yaml:"redact_reviews" json:"redactReviews"`
	RedactComments        bool `yaml:"redact_comments" json:"redactComments"`
	RedactCustomFields    bool `yaml:"redact_custom_fields" json:"redactCustomFields"`
	RedactModifiedBy      bool `yaml:"redact_modified_by" json:"redactModifiedBy"`
	RedactFiles           bool `yaml:"redact_files" json:"redactFiles"`
}

// DefaultRedactionConfig returns a RedactionConfig with all fields redacted —
// the safe default for sharing abstracts outside the organising committee.
func DefaultRedactionConfig() *RedactionConfig {
	return &RedactionConfig{
		RedactScore:           true,
		RedactJudge:           true,
		RedactJudgmentComment: true,
		RedactJudgmentDT:      true,
		RedactSubmitter:       true,
		RedactReviews:         true,
		RedactComments:        true,
		RedactCustomFields:    true,
		RedactModifiedBy:      true,
		RedactFiles:           true,
	}
}

// NoRedactionConfig returns a RedactionConfig with nothing redacted (all fields exposed).
func NoRedactionConfig() *RedactionConfig {
	return &RedactionConfig{}
}

// rawFieldsToRedact maps each RedactionConfig flag to the JSON key(s) it removes
// from each abstract entry in the raw map[string]any payload.
// These names must match the json tags on AbstractData.
var rawFieldsToRedact = []struct {
	enabled func(*RedactionConfig) bool
	keys    []string
}{
	{func(rc *RedactionConfig) bool { return rc.RedactScore }, []string{"score"}},
	{func(rc *RedactionConfig) bool { return rc.RedactJudge }, []string{"judge"}},
	{func(rc *RedactionConfig) bool { return rc.RedactJudgmentComment }, []string{"judgment_comment"}},
	{func(rc *RedactionConfig) bool { return rc.RedactJudgmentDT }, []string{"judgment_dt"}},
	{func(rc *RedactionConfig) bool { return rc.RedactSubmitter }, []string{"submitter"}},
	{func(rc *RedactionConfig) bool { return rc.RedactReviews }, []string{"reviews"}},
	{func(rc *RedactionConfig) bool { return rc.RedactComments }, []string{"comments"}},
	{func(rc *RedactionConfig) bool { return rc.RedactCustomFields }, []string{"custom_fields"}},
	{func(rc *RedactionConfig) bool { return rc.RedactModifiedBy }, []string{"modified_by"}},
	{func(rc *RedactionConfig) bool { return rc.RedactFiles }, []string{"files"}},
}

// ApplyRedactionToRawMap redacts sensitive fields from the raw map[string]any
// returned by FetchAbstractsData (and written by the fetch-abstracts-data cmd).
// It mutates each abstract entry map in-place inside a deep copy of the top-level map,
// so the caller's original map is never modified.
func (rc *RedactionConfig) ApplyRedactionToRawMap(rawData map[string]any) map[string]any {
	if rc == nil || rawData == nil {
		return rawData
	}

	// Build the list of keys to delete for this config.
	var keysToDelete []string
	for _, f := range rawFieldsToRedact {
		if f.enabled(rc) {
			keysToDelete = append(keysToDelete, f.keys...)
		}
	}

	// If nothing to redact, return the original map unchanged.
	if len(keysToDelete) == 0 {
		return rawData
	}

	// Shallow-copy the top-level map so we don't mutate the caller's map.
	result := make(map[string]any, len(rawData))
	for k, v := range rawData {
		result[k] = v
	}

	// Walk the "abstracts" array and redact each entry.
	abstractsRaw, ok := rawData["abstracts"]
	if !ok {
		return result
	}
	abstractsList, ok := abstractsRaw.([]any)
	if !ok {
		return result
	}

	redactedList := make([]any, len(abstractsList))
	for i, entry := range abstractsList {
		entryMap, ok := entry.(map[string]any)
		if !ok {
			redactedList[i] = entry
			continue
		}
		// Shallow-copy the abstract entry map.
		cp := make(map[string]any, len(entryMap))
		for k, v := range entryMap {
			cp[k] = v
		}
		// Delete the redacted keys.
		for _, key := range keysToDelete {
			delete(cp, key)
		}
		redactedList[i] = cp
	}
	result["abstracts"] = redactedList

	return result
}

// ApplyRedaction is a convenience wrapper that applies redaction to a flat
// []AbstractData slice. Prefer ApplyRedactionToRawMap for file exports.
func (rc *RedactionConfig) ApplyRedaction(abstracts []indico.AbstractData) []indico.AbstractData {
	if rc == nil {
		return abstracts
	}
	result := make([]indico.AbstractData, len(abstracts))
	for i, a := range abstracts {
		cp := a
		if rc.RedactScore {
			cp.Score = nil
		}
		if rc.RedactJudge {
			cp.Judge = nil
		}
		if rc.RedactJudgmentComment {
			cp.JudgmentComment = ""
		}
		if rc.RedactJudgmentDT {
			cp.JudgmentDT = ""
		}
		if rc.RedactSubmitter {
			cp.Submitter = nil
		}
		if rc.RedactReviews {
			cp.Reviews = nil
		}
		if rc.RedactComments {
			cp.Comments = nil
		}
		if rc.RedactCustomFields {
			cp.CustomFields = nil
		}
		if rc.RedactModifiedBy {
			cp.ModifiedBy = nil
		}
		if rc.RedactFiles {
			cp.Files = nil
		}
		result[i] = cp
	}
	return result
}

// VisibilityConfig defines what data should be visible in review mode
type VisibilityConfig struct {
	// Abstract card/table view settings
	ShowFirstPriority  bool
	ShowSecondPriority bool

	// Chart view settings
	ShowSubmissionTab bool
	ShowByReviewerTab bool
	ShowByTrackTab    bool
	ShowByActionTab   bool
	ShowRatingsTab    bool
	ShowTimelineTab   bool
	ShowMatrixTab     bool
}

// DefaultVisibilityConfig returns the default visibility configuration
// when NOT in review mode (everything visible)
func DefaultVisibilityConfig() *VisibilityConfig {
	return &VisibilityConfig{
		ShowFirstPriority:  true,
		ShowSecondPriority: true,
		ShowSubmissionTab:  true,
		ShowByReviewerTab:  true,
		ShowByTrackTab:     true,
		ShowByActionTab:    true,
		ShowRatingsTab:     true,
		ShowTimelineTab:    true,
		ShowMatrixTab:      true,
	}
}

// ReviewModeVisibilityConfig returns the visibility configuration
// for review mode (restricted visibility)
func ReviewModeVisibilityConfig() *VisibilityConfig {
	return &VisibilityConfig{
		ShowFirstPriority:  false,
		ShowSecondPriority: false,
		ShowSubmissionTab:  false,
		ShowByReviewerTab:  false,
		ShowByTrackTab:     false,
		ShowByActionTab:    false,
		ShowRatingsTab:     false,
		ShowTimelineTab:    false,
		ShowMatrixTab:      false,
	}
}
