package constraints

import (
	"fmt"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// ByeConstraint ensures each team gets exactly one bye per full round-robin
type ByeConstraint struct {
	BaseConstraint
}

// NewByeConstraint creates a new bye constraint
func NewByeConstraint() *ByeConstraint {
	return &ByeConstraint{
		BaseConstraint: NewBaseConstraint(
			"ByeConstraint",
			"Each team must get exactly one bye per full round-robin cycle",
			true, // This is a hard constraint
		),
	}
}

// Validate checks if the bye distribution violates the constraint
func (bc *ByeConstraint) Validate(match *models.Match, draw *models.Draw) error {
	// This constraint is evaluated at the draw level, not per match
	// Individual match validation always passes
	return nil
}

// Score calculates how well the draw satisfies the bye constraint
func (bc *ByeConstraint) Score(draw *models.Draw) float64 {
	// Get unique teams in the draw
	teamIDs := bc.getUniqueTeams(draw)
	if len(teamIDs) == 0 {
		return 1.0
	}
	
	// Calculate expected byes per team based on total rounds and team count
	totalTeams := len(teamIDs)
	
	// If even number of teams, no byes needed
	if totalTeams%2 == 0 {
		// Check that no team has any byes
		for _, teamID := range teamIDs {
			if bc.countByesForTeam(draw, teamID) > 0 {
				return 0.0
			}
		}
		return 1.0
	}
	
	// For odd number of teams, each team should have equal byes
	// In a single round-robin, each team should have exactly 1 bye
	expectedByesPerTeam := 1
	if draw.Rounds > totalTeams-1 {
		// For multiple round-robins, calculate expected byes
		fullRoundRobins := draw.Rounds / (totalTeams - 1)
		expectedByesPerTeam = fullRoundRobins
	}
	
	correctByeCount := 0
	for _, teamID := range teamIDs {
		actualByes := bc.countByesForTeam(draw, teamID)
		if actualByes == expectedByesPerTeam {
			correctByeCount++
		}
	}
	
	return float64(correctByeCount) / float64(totalTeams)
}

// ValidateDrawByes performs comprehensive bye validation for the entire draw
func (bc *ByeConstraint) ValidateDrawByes(draw *models.Draw) error {
	teamIDs := bc.getUniqueTeams(draw)
	if len(teamIDs) == 0 {
		return fmt.Errorf("no teams found in draw")
	}
	
	totalTeams := len(teamIDs)
	
	// If even number of teams, no byes should exist
	if totalTeams%2 == 0 {
		for _, teamID := range teamIDs {
			byeCount := bc.countByesForTeam(draw, teamID)
			if byeCount > 0 {
				return fmt.Errorf("team %d has %d byes but none expected with %d teams", 
					teamID, byeCount, totalTeams)
			}
		}
		return nil
	}
	
	// For odd number of teams, validate bye distribution
	expectedByesPerTeam := 1
	if draw.Rounds > totalTeams-1 {
		fullRoundRobins := draw.Rounds / (totalTeams - 1)
		expectedByesPerTeam = fullRoundRobins
	}
	
	for _, teamID := range teamIDs {
		actualByes := bc.countByesForTeam(draw, teamID)
		if actualByes != expectedByesPerTeam {
			return fmt.Errorf("team %d has %d byes but expected %d", 
				teamID, actualByes, expectedByesPerTeam)
		}
	}
	
	// Validate bye distribution across rounds
	return bc.validateByeDistribution(draw, teamIDs)
}

// getUniqueTeams extracts all unique team IDs from the draw
func (bc *ByeConstraint) getUniqueTeams(draw *models.Draw) []int {
	teamSet := make(map[int]bool)
	
	for _, match := range draw.Matches {
		if match.HomeTeamID != nil {
			teamSet[*match.HomeTeamID] = true
		}
		if match.AwayTeamID != nil {
			teamSet[*match.AwayTeamID] = true
		}
	}
	
	var teams []int
	for teamID := range teamSet {
		teams = append(teams, teamID)
	}
	
	return teams
}

// countByesForTeam counts how many byes a specific team has
func (bc *ByeConstraint) countByesForTeam(draw *models.Draw, teamID int) int {
	byeCount := 0
	
	// Count rounds where team has no matches
	roundsWithMatches := make(map[int]bool)
	
	for _, match := range draw.Matches {
		if (match.HomeTeamID != nil && *match.HomeTeamID == teamID) ||
			(match.AwayTeamID != nil && *match.AwayTeamID == teamID) {
			roundsWithMatches[match.Round] = true
		}
	}
	
	// Count total rounds vs rounds with matches
	for round := 1; round <= draw.Rounds; round++ {
		if !roundsWithMatches[round] {
			byeCount++
		}
	}
	
	return byeCount
}

// validateByeDistribution ensures byes are properly distributed across rounds
func (bc *ByeConstraint) validateByeDistribution(draw *models.Draw, teamIDs []int) error {
	// Count byes per round
	byesPerRound := make(map[int]int)
	
	for round := 1; round <= draw.Rounds; round++ {
		teamsInRound := make(map[int]bool)
		
		roundMatches := draw.GetMatchesByRound(round)
		for _, match := range roundMatches {
			if match.HomeTeamID != nil {
				teamsInRound[*match.HomeTeamID] = true
			}
			if match.AwayTeamID != nil {
				teamsInRound[*match.AwayTeamID] = true
			}
		}
		
		byesPerRound[round] = len(teamIDs) - len(teamsInRound)
	}
	
	// For odd number of teams, each round should have exactly 1 bye
	if len(teamIDs)%2 == 1 {
		expectedByesPerRound := 1
		for round, byeCount := range byesPerRound {
			if byeCount != expectedByesPerRound {
				return fmt.Errorf("round %d has %d byes but expected %d", 
					round, byeCount, expectedByesPerRound)
			}
		}
	}
	
	return nil
}

// GetTeamByes returns bye information for all teams
func (bc *ByeConstraint) GetTeamByes(draw *models.Draw) map[int][]int {
	teamByes := make(map[int][]int)
	teamIDs := bc.getUniqueTeams(draw)
	
	for _, teamID := range teamIDs {
		var byeRounds []int
		
		for round := 1; round <= draw.Rounds; round++ {
			hasMatchInRound := false
			
			for _, match := range draw.Matches {
				if match.Round == round && 
					((match.HomeTeamID != nil && *match.HomeTeamID == teamID) ||
					 (match.AwayTeamID != nil && *match.AwayTeamID == teamID)) {
					hasMatchInRound = true
					break
				}
			}
			
			if !hasMatchInRound {
				byeRounds = append(byeRounds, round)
			}
		}
		
		teamByes[teamID] = byeRounds
	}
	
	return teamByes
}