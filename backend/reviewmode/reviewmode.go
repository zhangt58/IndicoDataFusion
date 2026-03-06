package reviewmode

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
