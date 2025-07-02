package constraints

import (
	"testing"
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// TestByeConstraint tests the bye constraint implementation
func TestByeConstraint(t *testing.T) {
	constraint := NewByeConstraint()
	
	// Test constraint properties
	if constraint.Name() != "ByeConstraint" {
		t.Error("Wrong constraint name")
	}
	if !constraint.IsHard() {
		t.Error("Bye constraint should be hard")
	}
	
	// Test with valid bye distribution (3 teams, each gets 1 bye)
	draw := createTestDrawWithByes()
	score := constraint.Score(draw)
	if score != 1.0 {
		t.Errorf("Expected perfect score for valid bye distribution, got %f", score)
	}
	
	// Test bye analysis
	teamByes := constraint.GetTeamByes(draw)
	expectedByes := map[int]int{1: 1, 2: 1, 3: 1} // Each team should have 1 bye
	
	for teamID, expectedByeCount := range expectedByes {
		actualByeCount := len(teamByes[teamID])
		if actualByeCount != expectedByeCount {
			t.Errorf("Team %d should have %d byes, got %d", teamID, expectedByeCount, actualByeCount)
		}
	}
	
	// Test validation
	err := constraint.ValidateDrawByes(draw)
	if err != nil {
		t.Errorf("Valid bye distribution should pass validation: %v", err)
	}
}

// TestDoubleUpConstraint tests the double-up constraint implementation
func TestDoubleUpConstraint(t *testing.T) {
	constraint := NewDoubleUpConstraint(5)
	
	// Test constraint properties
	if constraint.Name() != "DoubleUpConstraint" {
		t.Error("Wrong constraint name")
	}
	if !constraint.IsHard() {
		t.Error("Double-up constraint should be hard")
	}
	if constraint.GetMinRoundsSeparation() != 5 {
		t.Error("Wrong minimum rounds separation")
	}
	
	// Create a draw with teams playing too close together
	draw := createTestDrawWithViolations()
	
	// This should score poorly due to violations
	score := constraint.Score(draw)
	if score == 1.0 {
		t.Error("Should have violations in test draw")
	}
	
	// Test getting violating matchups
	violatingMatchups := constraint.GetViolatingMatchups(draw)
	if len(violatingMatchups) == 0 {
		t.Error("Should detect violating matchups")
	}
	
	// Test validation errors
	errors := constraint.ValidateEntireDraw(draw)
	if len(errors) == 0 {
		t.Error("Should return validation errors for violating draw")
	}
}

// TestVenueAvailabilityConstraint tests venue availability constraint
func TestVenueAvailabilityConstraint(t *testing.T) {
	unavailableDates := []time.Time{
		time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 7, 4, 0, 0, 0, 0, time.UTC),
	}
	
	constraint := NewVenueAvailabilityConstraint(1, unavailableDates)
	
	// Test constraint properties
	if constraint.Name() != "VenueAvailability" {
		t.Error("Wrong constraint name")
	}
	if !constraint.IsHard() {
		t.Error("Venue availability constraint should be hard")
	}
	if constraint.GetVenueID() != 1 {
		t.Error("Wrong venue ID")
	}
	
	// Create match on unavailable date
	match := &models.Match{
		ID:         1,
		DrawID:     1,
		Round:      1,
		HomeTeamID: &[]int{1}[0],
		AwayTeamID: &[]int{2}[0],
		VenueID:    &[]int{1}[0], // Venue 1
		MatchDate:  &unavailableDates[0],
	}
	
	draw := &models.Draw{
		Matches: []*models.Match{match},
	}
	
	// Should violate constraint
	err := constraint.Validate(match, draw)
	if err == nil {
		t.Error("Should violate constraint for unavailable date")
	}
	
	// Should score poorly
	score := constraint.Score(draw)
	if score != 0.0 {
		t.Errorf("Expected score 0.0 for violation, got %f", score)
	}
	
	// Test with available date
	availableDate := time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC)
	match.MatchDate = &availableDate
	
	err = constraint.Validate(match, draw)
	if err != nil {
		t.Errorf("Should not violate constraint for available date: %v", err)
	}
}

// TestTeamAvailabilityConstraint tests team availability constraint
func TestTeamAvailabilityConstraint(t *testing.T) {
	unavailableDates := []time.Time{
		time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
	}
	
	constraint := NewTeamAvailabilityConstraint(1, unavailableDates)
	
	// Test constraint properties
	if constraint.GetTeamID() != 1 {
		t.Error("Wrong team ID")
	}
	
	// Create match with team on unavailable date
	match := &models.Match{
		ID:         1,
		DrawID:     1,
		Round:      1,
		HomeTeamID: &[]int{1}[0], // Team 1
		AwayTeamID: &[]int{2}[0],
		MatchDate:  &unavailableDates[0],
	}
	
	draw := &models.Draw{
		Matches: []*models.Match{match},
	}
	
	// Should violate constraint
	err := constraint.Validate(match, draw)
	if err == nil {
		t.Error("Should violate constraint for team unavailable date")
	}
	
	// Test conflicting matches
	conflictingMatches := constraint.GetConflictingMatches(draw)
	if len(conflictingMatches) != 1 {
		t.Error("Should detect one conflicting match")
	}
}

// TestTravelMinimizationConstraint tests travel minimization constraint
func TestTravelMinimizationConstraint(t *testing.T) {
	constraint := NewTravelMinimizationConstraint(2)
	
	// Test constraint properties
	if constraint.Name() != "TravelMinimization" {
		t.Error("Wrong constraint name")
	}
	if constraint.IsHard() {
		t.Error("Travel minimization should be soft constraint")
	}
	if constraint.GetMaxConsecutiveAway() != 2 {
		t.Error("Wrong max consecutive away")
	}
	
	// Create draw with excessive consecutive away games
	draw := createDrawWithConsecutiveAwayGames()
	
	// Should score less than perfect
	score := constraint.Score(draw)
	if score == 1.0 {
		t.Error("Should penalize excessive consecutive away games")
	}
	
	// Test team analysis
	analysis := constraint.AnalyzeTeamTravel(draw, 1)
	if analysis.TeamID != 1 {
		t.Error("Wrong team ID in analysis")
	}
	if analysis.LongestAwayStreak < 3 {
		t.Error("Should detect long away streak")
	}
	
	// Test getting worst travel teams
	worstTeams := constraint.GetWorstTravelTeams(draw, 1)
	if len(worstTeams) == 0 {
		t.Error("Should identify teams with poor travel")
	}
}

// TestRestPeriodConstraint tests rest period constraint
func TestRestPeriodConstraint(t *testing.T) {
	constraint := NewRestPeriodConstraint(3)
	
	// Test constraint properties
	if constraint.GetMinRestDays() != 3 {
		t.Error("Wrong minimum rest days")
	}
	
	// Create draw with insufficient rest periods
	draw := createDrawWithShortRestPeriods()
	
	// Should score less than perfect
	score := constraint.Score(draw)
	if score == 1.0 {
		t.Error("Should penalize insufficient rest periods")
	}
	
	// Test team analysis
	analysis := constraint.AnalyzeTeamRestPeriods(draw, 1)
	if analysis.ShortRestPeriods == 0 {
		t.Error("Should detect short rest periods")
	}
	
	// Test getting teams with short rest
	teamsWithShortRest := constraint.GetTeamsWithShortRest(draw)
	if len(teamsWithShortRest) == 0 {
		t.Error("Should identify teams with short rest")
	}
}

// TestPrimeTimeSpreadConstraint tests prime time spread constraint
func TestPrimeTimeSpreadConstraint(t *testing.T) {
	constraint := NewPrimeTimeSpreadConstraint(0.3, 0.1)
	
	// Test constraint properties
	if constraint.GetTargetPrimeTimeRatio() != 0.3 {
		t.Error("Wrong target ratio")
	}
	if constraint.GetMaxDeviation() != 0.1 {
		t.Error("Wrong max deviation")
	}
	
	// Create draw with uneven prime time distribution
	draw := createDrawWithUnevenPrimeTime()
	
	// Test team analysis
	analysis := constraint.AnalyzeTeamPrimeTimeDistribution(draw, 1)
	if analysis.TeamID != 1 {
		t.Error("Wrong team ID in analysis")
	}
	
	// Test getting teams with poor distribution
	poorDistribution := constraint.GetTeamsWithPoorDistribution(draw)
	if len(poorDistribution) == 0 {
		t.Error("Should identify teams with poor prime time distribution")
	}
	
	// Test adjustment suggestions
	adjustments := constraint.SuggestPrimeTimeAdjustments(draw)
	if len(adjustments) == 0 {
		t.Error("Should suggest adjustments for poor distribution")
	}
}

// TestHomeAwayBalanceConstraint tests home/away balance constraint
func TestHomeAwayBalanceConstraint(t *testing.T) {
	constraint := NewHomeAwayBalanceConstraint(0.1)
	
	// Test constraint properties
	if constraint.GetMaxDeviation() != 0.1 {
		t.Error("Wrong max deviation")
	}
	
	// Create draw with unbalanced home/away
	draw := createDrawWithUnbalancedHomeAway()
	
	// Test team analysis
	analysis := constraint.AnalyzeTeamHomeAwayBalance(draw, 1)
	if analysis.TeamID != 1 {
		t.Error("Wrong team ID in analysis")
	}
	
	// Test getting teams with poor balance
	poorBalance := constraint.GetTeamsWithPoorBalance(draw)
	if len(poorBalance) == 0 {
		t.Error("Should identify teams with poor balance")
	}
	
	// Test sequence analysis
	sequenceAnalysis := constraint.AnalyzeHomeAwaySequences(draw, 1)
	if sequenceAnalysis.TeamID != 1 {
		t.Error("Wrong team ID in sequence analysis")
	}
	
	// Test balance adjustments
	adjustments := constraint.SuggestBalanceAdjustments(draw)
	if len(adjustments) == 0 {
		t.Error("Should suggest balance adjustments")
	}
}

// Helper functions for creating test draws with specific patterns

func createTestDrawWithViolations() *models.Draw {
	// Create a draw where teams play each other too close together
	draw := &models.Draw{
		ID:         1,
		Name:       "Test Draw with Violations",
		SeasonYear: 2025,
		Rounds:     4,
		Status:     models.DrawStatusDraft,
		Matches: []*models.Match{
			{ID: 1, DrawID: 1, Round: 1, HomeTeamID: &[]int{1}[0], AwayTeamID: &[]int{2}[0]},
			{ID: 2, DrawID: 1, Round: 2, HomeTeamID: &[]int{1}[0], AwayTeamID: &[]int{2}[0]}, // Same teams too soon
		},
	}
	return draw
}

func createDrawWithConsecutiveAwayGames() *models.Draw {
	// Team 1 plays 4 consecutive away games
	draw := &models.Draw{
		ID:         1,
		Name:       "Draw with Consecutive Away",
		SeasonYear: 2025,
		Rounds:     4,
		Status:     models.DrawStatusDraft,
		Matches: []*models.Match{
			{ID: 1, DrawID: 1, Round: 1, HomeTeamID: &[]int{2}[0], AwayTeamID: &[]int{1}[0]}, // Away
			{ID: 2, DrawID: 1, Round: 2, HomeTeamID: &[]int{3}[0], AwayTeamID: &[]int{1}[0]}, // Away
			{ID: 3, DrawID: 1, Round: 3, HomeTeamID: &[]int{4}[0], AwayTeamID: &[]int{1}[0]}, // Away
			{ID: 4, DrawID: 1, Round: 4, HomeTeamID: &[]int{5}[0], AwayTeamID: &[]int{1}[0]}, // Away
		},
	}
	return draw
}

func createDrawWithShortRestPeriods() *models.Draw {
	// Matches with very short rest periods
	date1 := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC) // Only 1 day rest
	
	draw := &models.Draw{
		ID:         1,
		Name:       "Draw with Short Rest",
		SeasonYear: 2025,
		Rounds:     2,
		Status:     models.DrawStatusDraft,
		Matches: []*models.Match{
			{ID: 1, DrawID: 1, Round: 1, HomeTeamID: &[]int{1}[0], AwayTeamID: &[]int{2}[0], MatchDate: &date1},
			{ID: 2, DrawID: 1, Round: 2, HomeTeamID: &[]int{1}[0], AwayTeamID: &[]int{3}[0], MatchDate: &date2},
		},
	}
	return draw
}

func createDrawWithUnevenPrimeTime() *models.Draw {
	// Team 1 gets all prime time games, team 2 gets none
	draw := &models.Draw{
		ID:         1,
		Name:       "Draw with Uneven Prime Time",
		SeasonYear: 2025,
		Rounds:     4,
		Status:     models.DrawStatusDraft,
		Matches: []*models.Match{
			{ID: 1, DrawID: 1, Round: 1, HomeTeamID: &[]int{1}[0], AwayTeamID: &[]int{3}[0], IsPrimeTime: true},
			{ID: 2, DrawID: 1, Round: 2, HomeTeamID: &[]int{1}[0], AwayTeamID: &[]int{4}[0], IsPrimeTime: true},
			{ID: 3, DrawID: 1, Round: 3, HomeTeamID: &[]int{2}[0], AwayTeamID: &[]int{3}[0], IsPrimeTime: false},
			{ID: 4, DrawID: 1, Round: 4, HomeTeamID: &[]int{2}[0], AwayTeamID: &[]int{4}[0], IsPrimeTime: false},
		},
	}
	return draw
}

func createDrawWithUnbalancedHomeAway() *models.Draw {
	// Team 1 plays all home games, team 2 plays all away games
	draw := &models.Draw{
		ID:         1,
		Name:       "Draw with Unbalanced Home/Away",
		SeasonYear: 2025,
		Rounds:     4,
		Status:     models.DrawStatusDraft,
		Matches: []*models.Match{
			{ID: 1, DrawID: 1, Round: 1, HomeTeamID: &[]int{1}[0], AwayTeamID: &[]int{2}[0]},
			{ID: 2, DrawID: 1, Round: 2, HomeTeamID: &[]int{1}[0], AwayTeamID: &[]int{3}[0]},
			{ID: 3, DrawID: 1, Round: 3, HomeTeamID: &[]int{1}[0], AwayTeamID: &[]int{4}[0]},
			{ID: 4, DrawID: 1, Round: 4, HomeTeamID: &[]int{3}[0], AwayTeamID: &[]int{2}[0]},
		},
	}
	return draw
}