package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"IndicoDataFusion/backend/data"
	"IndicoDataFusion/backend/indico"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test-review-api <config-file>")
		fmt.Println("Example: test-review-api config/config.yaml")
		os.Exit(1)
	}

	configPath := os.Args[1]

	// Initialize data handler from config
	handler, err := data.NewDataSourceHandlerFromConfigFile(configPath)
	if err != nil {
		log.Fatalf("Failed to initialize data handler: %v", err)
	}

	ctx := context.Background()

	// Ensure we have a client (not test mode)
	client := handler.GetClient()
	if client == nil {
		log.Fatal("No Indico client available - test mode not supported for this test")
	}

	// Step 1: Get review tracks to populate user ID
	fmt.Println("=== Step 1: Fetching Review Tracks ===")
	tracks, err := client.GetReviewTracks(ctx)
	if err != nil {
		log.Fatalf("Failed to get review tracks: %v", err)
	}

	fmt.Printf("Found %d review tracks\n", len(tracks.Tracks))
	fmt.Printf("Current User ID: %d\n\n", client.UserID)

	if client.UserID == 0 {
		log.Fatal("User ID not found - cannot proceed with review testing")
	}

	// Step 2: Get abstracts to find ones to review
	fmt.Println("=== Step 2: Fetching Abstracts ===")
	abstracts, err := handler.GetAbstracts(ctx)
	if err != nil {
		log.Fatalf("Failed to get abstracts: %v", err)
	}

	fmt.Printf("Total abstracts: %d\n\n", len(abstracts))

	// Find abstracts assigned for review
	var myReviewAbstracts []indico.AbstractData
	for _, abstract := range abstracts {
		if abstract.IsMyReview {
			myReviewAbstracts = append(myReviewAbstracts, abstract)
		}
	}

	fmt.Printf("Abstracts assigned for my review: %d\n\n", len(myReviewAbstracts))

	if len(myReviewAbstracts) == 0 {
		fmt.Println("No abstracts assigned for review")
		return
	}

	// Step 3: Display my reviews
	fmt.Println("=== Step 3: My Review Status ===")
	reviewedCount := 0
	notReviewedCount := 0

	for i, abstract := range myReviewAbstracts {
		fmt.Printf("\n[%d] Abstract #%d: %s\n", i+1, abstract.FriendlyID, abstract.Title)
		fmt.Printf("    State: %s\n", abstract.State)
		fmt.Printf("    Database ID: %d\n", abstract.ID)

		if abstract.MyReview != nil {
			reviewedCount++
			fmt.Printf("    ✅ REVIEWED (Review ID: %d)\n", abstract.MyReview.ID)
			fmt.Printf("       Track: %s (ID: %d)\n", abstract.MyReview.Track.Title, abstract.MyReview.Track.ID)
			fmt.Printf("       Proposed Action: %s\n", abstract.MyReview.ProposedAction)
			fmt.Printf("       Comment: %s\n", abstract.MyReview.Comment)
			fmt.Printf("       Created: %s\n", abstract.MyReview.CreatedDT)
			if abstract.MyReview.ModifiedDT != nil {
				fmt.Printf("       Modified: %s\n", *abstract.MyReview.ModifiedDT)
			}

			// Display ratings
			if len(abstract.MyReview.Ratings) > 0 {
				fmt.Printf("       Ratings:\n")
				for _, rating := range abstract.MyReview.Ratings {
					questionTitle := "Unknown"
					if rating.QuestionDetails != nil {
						questionTitle = rating.QuestionDetails.Title
					}
					fmt.Printf("         - %s (Q%d): %v\n", questionTitle, rating.Question, rating.Value)
				}
			}
		} else {
			notReviewedCount++
			fmt.Printf("    ⏳ NOT YET REVIEWED\n")
			if len(abstract.ReviewedForTracks) > 0 {
				fmt.Printf("       Assigned Tracks:\n")
				for _, track := range abstract.ReviewedForTracks {
					fmt.Printf("         - %s (ID: %d)\n", track.Title, track.ID)
				}
			}
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Reviewed: %d/%d\n", reviewedCount, len(myReviewAbstracts))
	fmt.Printf("Not Reviewed: %d/%d\n\n", notReviewedCount, len(myReviewAbstracts))

	// Step 4: Get question IDs
	fmt.Println("=== Step 4: Review Questions ===")

	// Get questions from the first abstract's Questions map
	var firstPriorityQID, secondPriorityQID int

	if len(myReviewAbstracts) > 0 && myReviewAbstracts[0].Questions != nil {
		// Use the Questions map from the first abstract
		questionMap := myReviewAbstracts[0].Questions

		for qID, q := range questionMap {
			title := strings.ToLower(q.Title)
			if title == "first priority" {
				firstPriorityQID = qID
			}
			if title == "second priority" {
				secondPriorityQID = qID
			}
		}

		fmt.Println("Available Questions:")
		for qID, q := range questionMap {
			marker := ""
			if qID == firstPriorityQID {
				marker = " ← FIRST PRIORITY"
			} else if qID == secondPriorityQID {
				marker = " ← SECOND PRIORITY"
			}
			fmt.Printf("  Q%d: %s%s\n", qID, q.Title, marker)
		}
		fmt.Println()
	} else {
		fmt.Println("No questions available in abstracts")
	}

	if firstPriorityQID == 0 || secondPriorityQID == 0 {
		fmt.Println("Warning: Could not identify first/second priority question IDs")
		fmt.Println("Please check the questions manually")
		return
	}

	// Step 5: Review API URLs
	fmt.Println("=== Step 5: Review Submission URLs ===")

	// Find one reviewed and one not reviewed abstract for demo
	var reviewedAbstract, notReviewedAbstract *indico.AbstractData
	for i := range myReviewAbstracts {
		if myReviewAbstracts[i].MyReview != nil && reviewedAbstract == nil {
			reviewedAbstract = &myReviewAbstracts[i]
		}
		if myReviewAbstracts[i].MyReview == nil && notReviewedAbstract == nil {
			notReviewedAbstract = &myReviewAbstracts[i]
		}
		if reviewedAbstract != nil && notReviewedAbstract != nil {
			break
		}
	}

	// Demonstrate UPDATE URL (for already reviewed abstract)
	if reviewedAbstract != nil {
		fmt.Printf("📝 UPDATE EXISTING REVIEW:\n")
		fmt.Printf("   Abstract: #%d - %s\n", reviewedAbstract.FriendlyID, reviewedAbstract.Title)
		fmt.Printf("   URL: POST %s/event/%d/abstracts/%d/reviews/%d/edit\n",
			client.BaseURL, client.EventID, reviewedAbstract.ID, reviewedAbstract.MyReview.ID)
		fmt.Printf("   Payload Example:\n")
		fmt.Printf("     track-%d-csrf_token={token}\n", reviewedAbstract.MyReview.Track.ID)
		fmt.Printf("     track-%d-question_%d=1  # first_priority\n", reviewedAbstract.MyReview.Track.ID, firstPriorityQID)
		fmt.Printf("     track-%d-question_%d=0  # second_priority\n", reviewedAbstract.MyReview.Track.ID, secondPriorityQID)
		fmt.Printf("     track-%d-proposed_action=accept\n", reviewedAbstract.MyReview.Track.ID)
		fmt.Printf("     track-%d-proposed_contribution_type=42\n", reviewedAbstract.MyReview.Track.ID)
		fmt.Printf("     track-%d-comment=Updated review\n\n", reviewedAbstract.MyReview.Track.ID)
	}

	// Demonstrate CREATE URL (for not yet reviewed abstract)
	if notReviewedAbstract != nil && len(notReviewedAbstract.ReviewedForTracks) > 0 {
		trackID := notReviewedAbstract.ReviewedForTracks[0].ID
		fmt.Printf("✨ CREATE NEW REVIEW:\n")
		fmt.Printf("   Abstract: #%d - %s\n", notReviewedAbstract.FriendlyID, notReviewedAbstract.Title)
		fmt.Printf("   URL: POST %s/event/%d/abstracts/%d/review/track/%d\n",
			client.BaseURL, client.EventID, notReviewedAbstract.ID, trackID)
		fmt.Printf("   Payload Example:\n")
		fmt.Printf("     track-%d-csrf_token={token}\n", trackID)
		fmt.Printf("     track-%d-question_%d=1  # first_priority\n", trackID, firstPriorityQID)
		fmt.Printf("     track-%d-question_%d=1  # second_priority\n", trackID, secondPriorityQID)
		fmt.Printf("     track-%d-proposed_action=accept\n", trackID)
		fmt.Printf("     track-%d-proposed_contribution_type=42\n", trackID)
		fmt.Printf("     track-%d-comment=Excellent work!\n\n", trackID)
	}

	fmt.Println("To actually submit/update reviews, use the abstract.SubmitNewReview()")
	fmt.Println("or abstract.UpdateReview() methods from your application code.")

	if reviewedAbstract != nil {
		fmt.Printf("\nContribTypes available: %+v\n", *reviewedAbstract.ContribTypesMap)
	}

	// test the review submission methods with the first abstract if available
	// update review:
	fmt.Printf("\n=== Review Submission Demo ===\n")
	if reviewedAbstract != nil && reviewedAbstract.MyReview != nil {
		fmt.Printf("\n📝 UPDATING EXISTING REVIEW\n")
		fmt.Printf("   Abstract: #%d - %s\n", reviewedAbstract.FriendlyID, reviewedAbstract.Title)

		// accept
		err = reviewedAbstract.UpdateReview(
			ctx, client,
			reviewedAbstract.MyReview.ID,
			reviewedAbstract.MyReview.Track.ID,
			1, 0,
			"accept",
			nil,
			nil,
			nil,
			"Updated proposed action to accept",
		)
		if err != nil {
			fmt.Printf("   ❌ Failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Successfully updated review\n")
		}

		// reject
		err = reviewedAbstract.UpdateReview(
			ctx, client,
			reviewedAbstract.MyReview.ID,
			reviewedAbstract.MyReview.Track.ID,
			1, 0,
			"reject",
			nil,
			nil,
			nil,
			"Updated proposed action to reject",
		)
		if err != nil {
			fmt.Printf("   ❌ Failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Successfully updated review\n")
		}

		// change_tracks
		proposedTrackIDs := []int{99}
		err = reviewedAbstract.UpdateReview(
			ctx, client,
			reviewedAbstract.MyReview.ID,
			reviewedAbstract.MyReview.Track.ID,
			1, 0,
			"change_tracks",
			nil,
			proposedTrackIDs,
			nil,
			"Updated proposed change to track 99",
		)
		if err != nil {
			fmt.Printf("   ❌ Failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Successfully updated review\n")
		}

		// mark_as_duplicate
		proposedRelatedAbstractFriendlyID := 123
		err = reviewedAbstract.UpdateReview(
			ctx, client,
			reviewedAbstract.MyReview.ID,
			reviewedAbstract.MyReview.Track.ID,
			1, 0,
			"mark_as_duplicate",
			nil,
			nil,
			&proposedRelatedAbstractFriendlyID,
			"Updated proposed duplicate of abstract 123 (friendly ID)",
		)
		if err != nil {
			fmt.Printf("   ❌ Failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Successfully updated review\n")
		}

		// merge
		proposedRelatedAbstractID := 12345
		err = reviewedAbstract.UpdateReview(
			ctx, client,
			reviewedAbstract.MyReview.ID,
			reviewedAbstract.MyReview.Track.ID,
			1, 0,
			"merge",
			nil,
			nil,
			&proposedRelatedAbstractID,
			"Updated proposed duplicate of abstract 12345 (database ID)",
		)
		if err != nil {
			fmt.Printf("   ❌ Failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Successfully updated review\n")
		}

	}

	// submit new review:
	if notReviewedAbstract != nil && len(notReviewedAbstract.ReviewedForTracks) > 0 {
		fmt.Printf("\n✨ CREATING NEW REVIEW\n")
		fmt.Printf("   Abstract: #%d - %s\n", notReviewedAbstract.FriendlyID, notReviewedAbstract.Title)

		// accept
		trackID := notReviewedAbstract.ReviewedForTracks[0].ID
		err = notReviewedAbstract.SubmitNewReview(
			ctx, client,
			trackID,
			1, 1,
			"accept",
			nil,
			nil,
			nil,
			"Excellent work!",
		)
		if err != nil {
			fmt.Printf("   ❌ Failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Successfully created new review\n")
		}

		// reject
		err = notReviewedAbstract.SubmitNewReview(
			ctx, client,
			trackID,
			1, 1,
			"reject",
			nil,
			nil,
			nil,
			"Excellent work, but I have to reject it.",
		)
		if err != nil {
			fmt.Printf("   ❌ Failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Successfully created new review\n")
		}

		// change_tracks
		err = notReviewedAbstract.SubmitNewReview(
			ctx, client,
			trackID,
			1, 1,
			"change_tracks",
			nil,
			[]int{91},
			nil,
			"Excellent work, but I have to change to track 91.",
		)
		if err != nil {
			fmt.Printf("   ❌ Failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Successfully created new review\n")
		}

		// mark_as_duplicate
		proposedRelatedAbstractFriendlyID := 123
		proposedRelatedAbstractID := 12345
		err = notReviewedAbstract.SubmitNewReview(
			ctx, client,
			trackID,
			1, 1,
			"mark_as_duplicate",
			nil,
			nil,
			&proposedRelatedAbstractID,
			fmt.Sprintf("Excellent work, but I have to mark it as dup of #%d (database ID %d).",
				proposedRelatedAbstractFriendlyID, proposedRelatedAbstractID),
		)
		if err != nil {
			fmt.Printf("   ❌ Failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Successfully created new review\n")
		}

		// merge
		err = notReviewedAbstract.SubmitNewReview(
			ctx, client,
			trackID,
			1, 1,
			"merge",
			nil,
			nil,
			&proposedRelatedAbstractID,
			fmt.Sprintf("Excellent work, but I have to merge with #%d (database ID %d).",
				proposedRelatedAbstractFriendlyID, proposedRelatedAbstractID),
		)
		if err != nil {
			fmt.Printf("   ❌ Failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Successfully created new review\n")
		}

	}

	fmt.Println("\n=== Test Complete ===")

}
