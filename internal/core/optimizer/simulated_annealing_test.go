package optimizer

import (
	"testing"
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/constraints"
	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

func TestNewSimulatedAnnealing(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 1000, engine)

	if sa.Temperature != 100.0 {
		t.Errorf("Expected temperature 100.0, got %f", sa.Temperature)
	}
	if sa.CoolingRate != 0.99 {
		t.Errorf("Expected cooling rate 0.99, got %f", sa.CoolingRate)
	}
	if sa.MaxIterations != 1000 {
		t.Errorf("Expected max iterations 1000, got %d", sa.MaxIterations)
	}
	if sa.ConstraintEngine != engine {
		t.Error("Expected constraint engine to be set")
	}
}

func TestOptimize_NilDraw(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	result, err := sa.Optimize(nil, nil)

	if err == nil {
		t.Error("Expected error for nil draw")
	}
	if result != nil {
		t.Error("Expected nil result for nil draw")
	}
}

func TestOptimize_EmptyDraw(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	draw := &models.Draw{
		ID:         1,
		Name:       "Test Draw",
		SeasonYear: 2025,
		Rounds:     26,
		Matches:    []*models.Match{},
	}

	result, err := sa.Optimize(draw, nil)

	if err == nil {
		t.Error("Expected error for empty draw")
	}
	if result != nil {
		t.Error("Expected nil result for empty draw")
	}
}

func TestOptimize_ValidDraw(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	// Create a simple draw with some matches
	draw := createTestDraw()

	result, err := sa.Optimize(draw, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Error("Expected optimization result")
	}

	// Verify result structure
	if result.Iterations != 100 {
		t.Errorf("Expected 100 iterations, got %d", result.Iterations)
	}
	if result.Duration <= 0 {
		t.Error("Expected positive duration")
	}
	if result.BestDraw == nil {
		t.Error("Expected best draw in result")
	}
}

func TestOptimize_WithCallback(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 500, engine)

	draw := createTestDraw()
	callbackCount := 0

	callback := func(progress OptimizationProgress) {
		callbackCount++
		if progress.Iteration < 0 {
			t.Error("Expected non-negative iteration")
		}
		if progress.Temperature < 0 {
			t.Error("Expected non-negative temperature")
		}
	}

	result, err := sa.Optimize(draw, callback)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if callbackCount == 0 {
		t.Error("Expected callback to be called")
	}
	if result == nil {
		t.Error("Expected optimization result")
	}
}

func TestCopyDraw(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	original := createTestDraw()
	copy := sa.copyDraw(original)

	// Verify basic properties are copied
	if copy.ID != original.ID {
		t.Error("Draw ID not copied correctly")
	}
	if copy.Name != original.Name {
		t.Error("Draw name not copied correctly")
	}
	if len(copy.Matches) != len(original.Matches) {
		t.Error("Matches not copied correctly")
	}

	// Verify it's a deep copy by modifying original
	original.Name = "Modified"
	if copy.Name == "Modified" {
		t.Error("Copy is not independent of original")
	}

	// Verify matches are deep copied
	if len(original.Matches) > 0 && len(copy.Matches) > 0 {
		original.Matches[0].Round = 999
		if copy.Matches[0].Round == 999 {
			t.Error("Match copy is not independent of original")
		}
	}
}

func TestGenerateNeighbor(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)

	draw := createTestDraw()
	
	neighbor, err := sa.generateNeighbor(draw)
	
	if err != nil {
		t.Errorf("Unexpected error generating neighbor: %v", err)
	}
	if neighbor == nil {
		t.Error("Expected neighbor draw")
	}
	
	// Verify it's a different object
	if neighbor == draw {
		t.Error("Neighbor should be a different object")
	}
	
	// Verify basic structure is maintained
	if len(neighbor.Matches) != len(draw.Matches) {
		t.Error("Neighbor should have same number of matches")
	}
}

func createTestDraw() *models.Draw {
	homeTeam1 := 1
	awayTeam1 := 2
	homeTeam2 := 3
	awayTeam2 := 4
	venue1 := 1
	venue2 := 2

	return &models.Draw{
		ID:         1,
		Name:       "Test Draw",
		SeasonYear: 2025,
		Rounds:     4,
		Status:     models.DrawStatusDraft,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Matches: []*models.Match{
			{
				ID:         1,
				DrawID:     1,
				Round:      1,
				HomeTeamID: &homeTeam1,
				AwayTeamID: &awayTeam1,
				VenueID:    &venue1,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			{
				ID:         2,
				DrawID:     1,
				Round:      1,
				HomeTeamID: &homeTeam2,
				AwayTeamID: &awayTeam2,
				VenueID:    &venue2,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			{
				ID:         3,
				DrawID:     1,
				Round:      2,
				HomeTeamID: &awayTeam1,
				AwayTeamID: &homeTeam1,
				VenueID:    &venue1,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			{
				ID:         4,
				DrawID:     1,
				Round:      2,
				HomeTeamID: &awayTeam2,
				AwayTeamID: &homeTeam2,
				VenueID:    &venue2,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
	}
}

