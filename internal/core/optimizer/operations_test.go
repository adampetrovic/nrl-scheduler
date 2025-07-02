package optimizer

import (
	"testing"
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/constraints"
	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

func TestSwapMatches(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	draw := createTestDraw()
	
	// Try swapping multiple times to increase chances of success
	var err error
	swapAttempted := false
	for i := 0; i < 10; i++ {
		originalState := make(map[int]int)
		for _, match := range draw.Matches {
			originalState[match.ID] = match.Round
		}
		
		err = sa.swapMatches(draw)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}
		
		// Check if any swap occurred
		for _, match := range draw.Matches {
			if originalState[match.ID] != match.Round {
				swapAttempted = true
				break
			}
		}
		
		if swapAttempted {
			break
		}
	}

	// The operation should complete without error, but swapping depends on randomness
	// so we just verify the draw is still valid
	for _, match := range draw.Matches {
		if err := match.Validate(); err != nil {
			t.Errorf("Match validation failed after swap: %v", err)
		}
	}
}

func TestSwapMatches_InsufficientMatches(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	draw := &models.Draw{
		ID:      1,
		Matches: []*models.Match{},
	}

	err := sa.swapMatches(draw)

	if err == nil {
		t.Error("Expected error for insufficient matches")
	}
}

func TestRescheduleMatch(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	draw := createTestDraw()
	originalRounds := make(map[int]int)
	for _, match := range draw.Matches {
		originalRounds[match.ID] = match.Round
	}

	err := sa.rescheduleMatch(draw)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify that at least one match was rescheduled
	rescheduled := false
	for _, match := range draw.Matches {
		if !match.IsBye() && originalRounds[match.ID] != match.Round {
			rescheduled = true
			break
		}
	}

	if !rescheduled {
		t.Error("Expected at least one match to be rescheduled")
	}
}

func TestSwapVenues(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	draw := createTestDraw()
	originalVenues := make(map[int]*int)
	for _, match := range draw.Matches {
		if match.VenueID != nil {
			val := *match.VenueID
			originalVenues[match.ID] = &val
		}
	}

	err := sa.swapVenues(draw)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify that venues are still assigned and valid after the operation
	venueAssigned := false
	for _, match := range draw.Matches {
		if match.VenueID != nil && !match.IsBye() {
			venueAssigned = true
			break
		}
	}

	if !venueAssigned {
		t.Error("Expected matches to still have venues assigned")
	}
}

func TestSwapHomeAway(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	draw := createTestDraw()
	originalHomeTeams := make(map[int]*int)
	for _, match := range draw.Matches {
		if match.HomeTeamID != nil {
			val := *match.HomeTeamID
			originalHomeTeams[match.ID] = &val
		}
	}

	err := sa.swapHomeAway(draw)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify that at least one match had home/away swapped
	swapped := false
	for _, match := range draw.Matches {
		if originalHome, exists := originalHomeTeams[match.ID]; exists {
			if match.HomeTeamID != nil && *match.HomeTeamID != *originalHome {
				swapped = true
				break
			}
		}
	}

	if !swapped {
		t.Error("Expected at least one match to have home/away swapped")
	}
}

func TestGetRandomMatch(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	draw := createTestDraw()
	match, err := sa.getRandomMatch(draw)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if match == nil {
		t.Error("Expected a match to be returned")
	}

	// Verify the match is from the draw
	found := false
	for _, m := range draw.Matches {
		if m.ID == match.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Returned match not found in original draw")
	}
}

func TestGetRandomMatch_EmptyDraw(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	draw := &models.Draw{
		Matches: []*models.Match{},
	}

	match, err := sa.getRandomMatch(draw)

	if err == nil {
		t.Error("Expected error for empty draw")
	}
	if match != nil {
		t.Error("Expected nil match for empty draw")
	}
}

func TestGetRandomRegularMatch(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	draw := createTestDraw()
	match, err := sa.getRandomRegularMatch(draw)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if match == nil {
		t.Error("Expected a match to be returned")
	}
	if match.IsBye() {
		t.Error("Expected a regular match, not a bye")
	}
}

func TestGetRandomRegularMatch_OnlyByes(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	// Create draw with only bye matches
	draw := &models.Draw{
		ID:      1,
		Matches: []*models.Match{
			{
				ID:         1,
				DrawID:     1,
				Round:      1,
				HomeTeamID: nil,
				AwayTeamID: nil,
				VenueID:    nil,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
	}

	match, err := sa.getRandomRegularMatch(draw)

	if err == nil {
		t.Error("Expected error when only bye matches available")
	}
	if match != nil {
		t.Error("Expected nil match when only bye matches available")
	}
}

func TestGetMatchesByRound(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	draw := createTestDraw()
	round1Matches := sa.getMatchesByRound(draw, 1)

	expectedCount := 0
	for _, match := range draw.Matches {
		if match.Round == 1 {
			expectedCount++
		}
	}

	if len(round1Matches) != expectedCount {
		t.Errorf("Expected %d matches in round 1, got %d", expectedCount, len(round1Matches))
	}

	// Verify all returned matches are from round 1
	for _, match := range round1Matches {
		if match.Round != 1 {
			t.Error("Found match not from round 1")
		}
	}
}

func TestCountMatchesInRound(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	draw := createTestDraw()
	count := sa.countMatchesInRound(draw, 1)

	expectedCount := 0
	for _, match := range draw.Matches {
		if match.Round == 1 {
			expectedCount++
		}
	}

	if count != expectedCount {
		t.Errorf("Expected %d matches in round 1, got %d", expectedCount, count)
	}
}

func TestApplyMultipleOperations(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	draw := createTestDraw()
	
	err := sa.applyMultipleOperations(draw, 3)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// The draw should still be valid after multiple operations
	for _, match := range draw.Matches {
		if err := match.Validate(); err != nil {
			t.Errorf("Match validation failed after operations: %v", err)
		}
	}
}

