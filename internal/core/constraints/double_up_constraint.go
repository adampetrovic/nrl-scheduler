package constraints

import (
	"fmt"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// DoubleUpConstraint ensures teams don't play each other twice within X rounds
type DoubleUpConstraint struct {
	BaseConstraint
	minRoundsSeparation int
}

// NewDoubleUpConstraint creates a new double-up constraint
func NewDoubleUpConstraint(minRoundsSeparation int) *DoubleUpConstraint {
	return &DoubleUpConstraint{
		BaseConstraint: NewBaseConstraint(
			"DoubleUpConstraint",
			fmt.Sprintf("Teams cannot play each other twice within %d rounds", minRoundsSeparation),
			true, // This is a hard constraint
		),
		minRoundsSeparation: minRoundsSeparation,
	}
}

// Validate checks if a match violates the double-up constraint
func (duc *DoubleUpConstraint) Validate(match *models.Match, draw *models.Draw) error {
	// Skip validation for bye matches
	if match.IsBye() {
		return nil
	}
	
	// Get the teams involved in this match
	homeTeam := *match.HomeTeamID
	awayTeam := *match.AwayTeamID
	
	// Find all other matches between these teams
	for _, otherMatch := range draw.Matches {
		// Skip the same match
		if otherMatch.ID == match.ID {
			continue
		}
		
		// Skip bye matches
		if otherMatch.IsBye() {
			continue
		}
		
		// Check if this is a match between the same teams
		if duc.areMatchesBetweenSameTeams(match, otherMatch) {
			// Check if they're too close together
			roundDiff := duc.calculateRoundDifference(match.Round, otherMatch.Round)
			if roundDiff < duc.minRoundsSeparation {
				return fmt.Errorf("teams %d and %d play each other in rounds %d and %d (only %d rounds apart, minimum %d required)",
					homeTeam, awayTeam, match.Round, otherMatch.Round, roundDiff, duc.minRoundsSeparation)
			}
		}
	}
	
	return nil
}

// Score calculates how well the draw satisfies this constraint
func (duc *DoubleUpConstraint) Score(draw *models.Draw) float64 {
	totalMatchups := 0
	violatingMatchups := 0
	
	// Get all unique team matchups
	matchups := duc.getAllMatchups(draw)
	
	for _, rounds := range matchups {
		if len(rounds) < 2 {
			continue // Single matchup, no violation possible
		}
		
		totalMatchups++
		
		// Check if any pair of rounds is too close
		for i := 0; i < len(rounds); i++ {
			for j := i + 1; j < len(rounds); j++ {
				roundDiff := duc.calculateRoundDifference(rounds[i], rounds[j])
				if roundDiff < duc.minRoundsSeparation {
					violatingMatchups++
					break
				}
			}
		}
	}
	
	// If no repeated matchups, constraint is perfectly satisfied
	if totalMatchups == 0 {
		return 1.0
	}
	
	// Return the percentage of non-violating matchups
	return float64(totalMatchups-violatingMatchups) / float64(totalMatchups)
}

// areMatchesBetweenSameTeams checks if two matches are between the same teams
func (duc *DoubleUpConstraint) areMatchesBetweenSameTeams(match1, match2 *models.Match) bool {
	if match1.IsBye() || match2.IsBye() {
		return false
	}
	
	home1, away1 := *match1.HomeTeamID, *match1.AwayTeamID
	home2, away2 := *match2.HomeTeamID, *match2.AwayTeamID
	
	// Check both directions (home/away can be swapped)
	return (home1 == home2 && away1 == away2) || (home1 == away2 && away1 == home2)
}

// calculateRoundDifference calculates the absolute difference between two rounds
func (duc *DoubleUpConstraint) calculateRoundDifference(round1, round2 int) int {
	diff := round1 - round2
	if diff < 0 {
		diff = -diff
	}
	return diff
}

// getAllMatchups returns a map of team matchups and the rounds they occur in
func (duc *DoubleUpConstraint) getAllMatchups(draw *models.Draw) map[string][]int {
	matchups := make(map[string][]int)
	
	for _, match := range draw.Matches {
		if match.IsBye() {
			continue
		}
		
		key := duc.getMatchupKey(*match.HomeTeamID, *match.AwayTeamID)
		matchups[key] = append(matchups[key], match.Round)
	}
	
	return matchups
}

// getMatchupKey creates a consistent key for a team matchup
func (duc *DoubleUpConstraint) getMatchupKey(team1, team2 int) string {
	// Always put the smaller team ID first for consistency
	if team1 > team2 {
		team1, team2 = team2, team1
	}
	return fmt.Sprintf("%d-%d", team1, team2)
}

// GetMinRoundsSeparation returns the minimum rounds separation
func (duc *DoubleUpConstraint) GetMinRoundsSeparation() int {
	return duc.minRoundsSeparation
}

// GetViolatingMatchups returns all matchups that violate the constraint
func (duc *DoubleUpConstraint) GetViolatingMatchups(draw *models.Draw) map[string][]int {
	violatingMatchups := make(map[string][]int)
	matchups := duc.getAllMatchups(draw)
	
	for matchupKey, rounds := range matchups {
		if len(rounds) < 2 {
			continue
		}
		
		// Check if any pair of rounds is too close
		hasViolation := false
		for i := 0; i < len(rounds) && !hasViolation; i++ {
			for j := i + 1; j < len(rounds); j++ {
				roundDiff := duc.calculateRoundDifference(rounds[i], rounds[j])
				if roundDiff < duc.minRoundsSeparation {
					hasViolation = true
					break
				}
			}
		}
		
		if hasViolation {
			violatingMatchups[matchupKey] = rounds
		}
	}
	
	return violatingMatchups
}

// ValidateEntireDraw performs comprehensive validation for the entire draw
func (duc *DoubleUpConstraint) ValidateEntireDraw(draw *models.Draw) []error {
	var errors []error
	
	violatingMatchups := duc.GetViolatingMatchups(draw)
	
	for matchupKey, rounds := range violatingMatchups {
		for i := 0; i < len(rounds); i++ {
			for j := i + 1; j < len(rounds); j++ {
				roundDiff := duc.calculateRoundDifference(rounds[i], rounds[j])
				if roundDiff < duc.minRoundsSeparation {
					errors = append(errors, fmt.Errorf(
						"matchup %s in rounds %d and %d (only %d rounds apart, minimum %d required)",
						matchupKey, rounds[i], rounds[j], roundDiff, duc.minRoundsSeparation))
				}
			}
		}
	}
	
	return errors
}

// GetMatchupFrequency returns how many times each team matchup occurs
func (duc *DoubleUpConstraint) GetMatchupFrequency(draw *models.Draw) map[string]int {
	frequency := make(map[string]int)
	matchups := duc.getAllMatchups(draw)
	
	for matchupKey, rounds := range matchups {
		frequency[matchupKey] = len(rounds)
	}
	
	return frequency
}

// GetRecommendedRescheduling suggests how to reschedule violating matches
func (duc *DoubleUpConstraint) GetRecommendedRescheduling(draw *models.Draw) map[string][]int {
	recommendations := make(map[string][]int)
	violatingMatchups := duc.GetViolatingMatchups(draw)
	
	for matchupKey, rounds := range violatingMatchups {
		var suggestedRounds []int
		
		// For each round, suggest alternatives that satisfy the constraint
		for _, round := range rounds {
			// Find nearby rounds that would satisfy the minimum separation
			for candidate := 1; candidate <= draw.Rounds; candidate++ {
				// Check if this candidate round would satisfy separation from all other rounds
				satisfiesSeparation := true
				for _, otherRound := range rounds {
					if otherRound == round {
						continue // Skip the round we're trying to reschedule
					}
					if duc.calculateRoundDifference(candidate, otherRound) < duc.minRoundsSeparation {
						satisfiesSeparation = false
						break
					}
				}
				
				if satisfiesSeparation {
					suggestedRounds = append(suggestedRounds, candidate)
				}
			}
		}
		
		if len(suggestedRounds) > 0 {
			recommendations[matchupKey] = suggestedRounds
		}
	}
	
	return recommendations
}