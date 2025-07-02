package constraints

import (
	"testing"
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// TestConstraintEngine tests the basic constraint engine functionality
func TestConstraintEngine(t *testing.T) {
	engine := NewConstraintEngine()
	
	// Test empty engine
	if len(engine.GetHardConstraints()) != 0 {
		t.Error("New engine should have no hard constraints")
	}
	if len(engine.GetSoftConstraints()) != 0 {
		t.Error("New engine should have no soft constraints")
	}
	
	// Create test constraints
	byeConstraint := NewByeConstraint()
	travelConstraint := NewTravelMinimizationConstraint(3)
	
	// Add constraints
	engine.AddHardConstraint(byeConstraint)
	engine.AddSoftConstraint(travelConstraint, 0.8)
	
	// Verify constraints were added
	if len(engine.GetHardConstraints()) != 1 {
		t.Error("Engine should have 1 hard constraint")
	}
	if len(engine.GetSoftConstraints()) != 1 {
		t.Error("Engine should have 1 soft constraint")
	}
	
	// Test constraint retrieval
	hardConstraints := engine.GetHardConstraints()
	if hardConstraints[0].Name() != "ByeConstraint" {
		t.Error("Wrong hard constraint name")
	}
	
	softConstraints := engine.GetSoftConstraints()
	if softConstraints[0].Constraint.Name() != "TravelMinimization" {
		t.Error("Wrong soft constraint name")
	}
	if softConstraints[0].Weight != 0.8 {
		t.Error("Wrong soft constraint weight")
	}
}

// TestConstraintEngineValidation tests draw validation
func TestConstraintEngineValidation(t *testing.T) {
	engine := NewConstraintEngine()
	
	// Create test draw with known violations
	draw := createTestDraw()
	
	// Add double-up constraint with tight restriction
	doubleUpConstraint := NewDoubleUpConstraint(10) // Teams can't play twice within 10 rounds
	engine.AddHardConstraint(doubleUpConstraint)
	
	// Validate draw
	violations := engine.ValidateDraw(draw)
	
	// Since our test draw is small (6 rounds), double-up should be satisfied
	if len(violations) > 0 {
		t.Errorf("Expected no violations for simple draw, got %d", len(violations))
	}
}

// TestConstraintEngineScoring tests draw scoring
func TestConstraintEngineScoring(t *testing.T) {
	engine := NewConstraintEngine()
	draw := createTestDraw()
	
	// Test with no constraints - should return perfect score
	score := engine.ScoreDraw(draw)
	if score != 1.0 {
		t.Errorf("Expected perfect score (1.0) with no constraints, got %f", score)
	}
	
	// Add soft constraints
	engine.AddSoftConstraint(NewTravelMinimizationConstraint(2), 0.5)
	engine.AddSoftConstraint(NewHomeAwayBalanceConstraint(0.1), 0.5)
	
	// Score should still be > 0
	score = engine.ScoreDraw(draw)
	if score < 0 || score > 1 {
		t.Errorf("Score should be between 0 and 1, got %f", score)
	}
}

// TestConstraintEngineAnalysis tests comprehensive draw analysis
func TestConstraintEngineAnalysis(t *testing.T) {
	engine := NewConstraintEngine()
	draw := createTestDraw()
	
	// Add various constraints
	engine.AddHardConstraint(NewByeConstraint())
	engine.AddSoftConstraint(NewTravelMinimizationConstraint(2), 0.8)
	
	// Analyze draw
	violations := engine.AnalyzeDraw(draw)
	
	// Should have some analysis results
	if violations == nil {
		t.Error("Analysis should return results, not nil")
	}
	
	// Verify violation structure
	for _, violation := range violations {
		if violation.ConstraintName == "" {
			t.Error("Violation should have constraint name")
		}
		if violation.Severity == "" {
			t.Error("Violation should have severity")
		}
	}
}

// TestBaseConstraint tests the base constraint functionality
func TestBaseConstraint(t *testing.T) {
	base := NewBaseConstraint("TestConstraint", "Test description", true)
	
	if base.Name() != "TestConstraint" {
		t.Error("Wrong constraint name")
	}
	if base.Description() != "Test description" {
		t.Error("Wrong constraint description")
	}
	if !base.IsHard() {
		t.Error("Constraint should be hard")
	}
	
	// Test soft constraint
	softBase := NewBaseConstraint("SoftTest", "Soft description", false)
	if softBase.IsHard() {
		t.Error("Constraint should be soft")
	}
}

// TestDateConstraint tests date-based constraint functionality
func TestDateConstraint(t *testing.T) {
	unavailableDates := []time.Time{
		time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 7, 4, 0, 0, 0, 0, time.UTC),
	}
	
	dateConstraint := NewDateConstraint("TestDate", "Test date constraint", true, unavailableDates)
	
	// Test date availability checking
	testDate1 := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	testDate2 := time.Date(2025, 6, 16, 0, 0, 0, 0, time.UTC)
	
	if !dateConstraint.IsDateUnavailable(testDate1) {
		t.Error("Date should be unavailable")
	}
	if dateConstraint.IsDateUnavailable(testDate2) {
		t.Error("Date should be available")
	}
	
	// Test getting unavailable dates
	retrievedDates := dateConstraint.GetUnavailableDates()
	if len(retrievedDates) != 2 {
		t.Error("Should retrieve 2 unavailable dates")
	}
}

// createTestDraw creates a simple test draw for testing
func createTestDraw() *models.Draw {
	teams := []*models.Team{
		{ID: 1, Name: "Team A", VenueID: &[]int{1}[0]},
		{ID: 2, Name: "Team B", VenueID: &[]int{2}[0]},
		{ID: 3, Name: "Team C", VenueID: &[]int{3}[0]},
		{ID: 4, Name: "Team D", VenueID: &[]int{4}[0]},
	}
	
	draw := &models.Draw{
		ID:         1,
		Name:       "Test Draw",
		SeasonYear: 2025,
		Rounds:     6,
		Status:     models.DrawStatusDraft,
		Matches:    []*models.Match{},
	}
	
	// Create some test matches
	matches := []*models.Match{
		{ID: 1, DrawID: 1, Round: 1, HomeTeamID: &teams[0].ID, AwayTeamID: &teams[1].ID, VenueID: teams[0].VenueID},
		{ID: 2, DrawID: 1, Round: 1, HomeTeamID: &teams[2].ID, AwayTeamID: &teams[3].ID, VenueID: teams[2].VenueID},
		{ID: 3, DrawID: 1, Round: 2, HomeTeamID: &teams[0].ID, AwayTeamID: &teams[2].ID, VenueID: teams[0].VenueID},
		{ID: 4, DrawID: 1, Round: 2, HomeTeamID: &teams[1].ID, AwayTeamID: &teams[3].ID, VenueID: teams[1].VenueID},
		{ID: 5, DrawID: 1, Round: 3, HomeTeamID: &teams[0].ID, AwayTeamID: &teams[3].ID, VenueID: teams[0].VenueID},
		{ID: 6, DrawID: 1, Round: 3, HomeTeamID: &teams[1].ID, AwayTeamID: &teams[2].ID, VenueID: teams[1].VenueID},
	}
	
	draw.Matches = matches
	return draw
}

// createTestDrawWithByes creates a test draw including bye rounds for odd team count
func createTestDrawWithByes() *models.Draw {
	teams := []*models.Team{
		{ID: 1, Name: "Team A", VenueID: &[]int{1}[0]},
		{ID: 2, Name: "Team B", VenueID: &[]int{2}[0]},
		{ID: 3, Name: "Team C", VenueID: &[]int{3}[0]},
	}
	
	draw := &models.Draw{
		ID:         1,
		Name:       "Test Draw with Byes",
		SeasonYear: 2025,
		Rounds:     3,
		Status:     models.DrawStatusDraft,
		Matches:    []*models.Match{},
	}
	
	// Create matches for 3 teams (each team gets 1 bye)
	matches := []*models.Match{
		{ID: 1, DrawID: 1, Round: 1, HomeTeamID: &teams[0].ID, AwayTeamID: &teams[1].ID, VenueID: teams[0].VenueID},
		// Team 3 has bye in round 1
		{ID: 2, DrawID: 1, Round: 2, HomeTeamID: &teams[0].ID, AwayTeamID: &teams[2].ID, VenueID: teams[0].VenueID},
		// Team 2 has bye in round 2
		{ID: 3, DrawID: 1, Round: 3, HomeTeamID: &teams[1].ID, AwayTeamID: &teams[2].ID, VenueID: teams[1].VenueID},
		// Team 1 has bye in round 3
	}
	
	draw.Matches = matches
	return draw
}

// Benchmark tests for performance
func BenchmarkConstraintEngineValidation(b *testing.B) {
	engine := NewConstraintEngine()
	engine.AddHardConstraint(NewByeConstraint())
	engine.AddHardConstraint(NewDoubleUpConstraint(5))
	
	draw := createTestDraw()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.ValidateDraw(draw)
	}
}

func BenchmarkConstraintEngineScoring(b *testing.B) {
	engine := NewConstraintEngine()
	engine.AddSoftConstraint(NewTravelMinimizationConstraint(3), 0.8)
	engine.AddSoftConstraint(NewHomeAwayBalanceConstraint(0.1), 0.7)
	
	draw := createTestDraw()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.ScoreDraw(draw)
	}
}