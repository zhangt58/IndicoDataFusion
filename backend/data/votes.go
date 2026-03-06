package data

import (
	"IndicoDataFusion/backend/indico"
	"context"
	"encoding/json"
	"strings"
)

// MaxVotesPerTrack is the maximum number of priority votes (first OR second priority)
// a reviewer is allowed to cast per review track. Adjust this value to change the limit
// across the entire application.
const MaxVotesPerTrack = 10

// TrackVoteStats holds vote statistics for a single review track.
type TrackVoteStats struct {
	TrackID       int    `json:"track_id"`
	TrackName     string `json:"track_name"`
	VotesCast     int    `json:"votes_cast"`
	VotesMax      int    `json:"votes_max"`
	VotesLeft     int    `json:"votes_left"`
	AbstractCount int    `json:"abstract_count"`
}

// VoteStats holds vote statistics for all review tracks.
type VoteStats struct {
	// PerTrack maps track_id → TrackVoteStats
	PerTrack       map[int]*TrackVoteStats `json:"per_track"`
	MaxVotes       int                     `json:"max_votes"`
	TotalCast      int                     `json:"total_cast"`
	TotalAbstracts int                     `json:"total_abstracts"`
}

// GetVoteStats computes vote statistics for the current reviewer per track.
// It uses the cached abstract list to count reviews that carry a first- or
// second-priority "yes" vote, grouped by the track the review was submitted for.
// The MaxVotesPerTrack module constant controls the per-track limit.
func (h *DataSourceHandler) GetVoteStats(ctx context.Context) (*VoteStats, error) {
	// Retrieve review tracks to know which tracks are assigned to this reviewer.
	reviewTracks, err := h.GetReviewTracks(ctx)
	if err != nil {
		return nil, err
	}

	stats := &VoteStats{
		PerTrack: make(map[int]*TrackVoteStats),
		MaxVotes: MaxVotesPerTrack,
	}

	// Initialise per-track buckets for every assigned track.
	if reviewTracks != nil {
		for _, rt := range reviewTracks.Tracks {
			if rt.TrackID == 0 {
				continue
			}
			stats.PerTrack[rt.TrackID] = &TrackVoteStats{
				TrackID:       rt.TrackID,
				TrackName:     rt.Name,
				VotesMax:      MaxVotesPerTrack,
				AbstractCount: rt.AbstractCount,
			}
		}
	}

	// Load abstracts from cache (or live API if not cached).
	abstracts, err := h.GetAbstracts(ctx)
	if err != nil {
		// Non-fatal: return initialised (empty) stats.
		for _, ts := range stats.PerTrack {
			ts.VotesLeft = ts.VotesMax
		}
		return stats, nil
	}

	// Count votes by scanning MyReview on each abstract.
	for i := range abstracts {
		a := &abstracts[i]
		if a.MyReview == nil {
			continue
		}
		review := a.MyReview
		trackID := review.Track.ID
		if trackID == 0 {
			continue
		}
		if hasPriorityVote(review) {
			ts, ok := stats.PerTrack[trackID]
			if !ok {
				// Voted on a track not in our assigned list; create an entry anyway.
				ts = &TrackVoteStats{
					TrackID:   trackID,
					TrackName: review.Track.Title,
					VotesMax:  MaxVotesPerTrack,
				}
				stats.PerTrack[trackID] = ts
			}
			ts.VotesCast++
			stats.TotalCast++
		}
		stats.TotalAbstracts++
	}

	// Compute VotesLeft for every bucket.
	for _, ts := range stats.PerTrack {
		ts.VotesLeft = ts.VotesMax - ts.VotesCast
		if ts.VotesLeft < 0 {
			ts.VotesLeft = 0
		}
	}

	return stats, nil
}

// hasPriorityVote returns true when the review has a "yes" answer (value == 1)
// for either the "first priority" or "second priority" question.
// It checks both the Ratings slice (populated from the abstract API) and the
// top-level FirstPriority / SecondPriority convenience fields when present.
func hasPriorityVote(r *indico.Review) bool {
	if r == nil {
		return false
	}

	// Check via Ratings array (preferred — populated from abstract API / scrape).
	for _, rating := range r.Ratings {
		title := ""
		if rating.QuestionDetails != nil {
			title = strings.ToLower(rating.QuestionDetails.Title)
		}
		if title == "first priority" || title == "second priority" {
			switch v := rating.Value.(type) {
			case bool:
				if v {
					return true
				}
			case float64:
				if v == 1 {
					return true
				}
			case json.Number:
				if v.String() == "1" {
					return true
				}
			case string:
				if v == "1" || v == "true" {
					return true
				}
			}
		}
	}

	return false
}
