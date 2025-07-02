package constraints

import (
	"math"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// TravelMinimizationConstraint minimizes consecutive away games for teams
type TravelMinimizationConstraint struct {
	BaseConstraint
	maxConsecutiveAway int
	penaltyWeight      float64
}

// NewTravelMinimizationConstraint creates a new travel minimization constraint
func NewTravelMinimizationConstraint(maxConsecutiveAway int) *TravelMinimizationConstraint {
	return &TravelMinimizationConstraint{
		BaseConstraint: NewBaseConstraint(
			"TravelMinimization",
			"Minimize consecutive away games to reduce travel burden",
			false, // This is a soft constraint
		),
		maxConsecutiveAway: maxConsecutiveAway,
		penaltyWeight:      1.0,
	}
}

// Validate always returns nil for soft constraints (no hard violations)
func (tmc *TravelMinimizationConstraint) Validate(match *models.Match, draw *models.Draw) error {
	// Soft constraints don't have hard validation failures
	return nil
}

// Score calculates how well the draw minimizes travel
func (tmc *TravelMinimizationConstraint) Score(draw *models.Draw) float64 {
	teams := tmc.getUniqueTeams(draw)
	if len(teams) == 0 {
		return 1.0
	}

	totalScore := 0.0

	for _, team := range teams {
		teamScore := tmc.scoreTeamTravel(draw, team)
		totalScore += teamScore
	}

	return totalScore / float64(len(teams))
}

// scoreTeamTravel calculates the travel score for a specific team
func (tmc *TravelMinimizationConstraint) scoreTeamTravel(draw *models.Draw, teamID int) float64 {
	teamMatches := tmc.getTeamMatchesByRound(draw, teamID)
	if len(teamMatches) == 0 {
		return 1.0
	}

	consecutiveAwayStreak := 0
	maxStreak := 0
	totalPenalty := 0.0

	// Analyze consecutive away games
	for round := 1; round <= draw.Rounds; round++ {
		match, exists := teamMatches[round]
		if !exists {
			// Bye round - reset streak
			consecutiveAwayStreak = 0
			continue
		}

		// Check if this is an away game
		if isAway, _ := match.IsHomeGame(teamID); !isAway {
			consecutiveAwayStreak++
			if consecutiveAwayStreak > maxStreak {
				maxStreak = consecutiveAwayStreak
			}

			// Apply penalty for excessive consecutive away games
			if consecutiveAwayStreak > tmc.maxConsecutiveAway {
				excess := consecutiveAwayStreak - tmc.maxConsecutiveAway
				totalPenalty += float64(excess) * tmc.penaltyWeight
			}
		} else {
			// Home game - reset streak
			consecutiveAwayStreak = 0
		}
	}

	// Calculate score based on penalties
	// Perfect score (1.0) when no excessive streaks
	// Score decreases with penalty accumulation
	if totalPenalty == 0 {
		return 1.0
	}

	// Normalize penalty to a score between 0 and 1
	// Maximum penalty would be if all games were away and exceeded limit
	maxPossiblePenalty := float64(len(teamMatches)) * tmc.penaltyWeight
	score := 1.0 - (totalPenalty / maxPossiblePenalty)

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	return score
}

// getUniqueTeams extracts all unique team IDs from the draw
func (tmc *TravelMinimizationConstraint) getUniqueTeams(draw *models.Draw) []int {
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

// getTeamMatchesByRound returns team matches organized by round
func (tmc *TravelMinimizationConstraint) getTeamMatchesByRound(draw *models.Draw, teamID int) map[int]*models.Match {
	matches := make(map[int]*models.Match)

	for _, match := range draw.Matches {
		if match.HasTeam(teamID) {
			matches[match.Round] = match
		}
	}

	return matches
}

// GetMaxConsecutiveAway returns the maximum allowed consecutive away games
func (tmc *TravelMinimizationConstraint) GetMaxConsecutiveAway() int {
	return tmc.maxConsecutiveAway
}

// SetPenaltyWeight sets the penalty weight for excessive consecutive away games
func (tmc *TravelMinimizationConstraint) SetPenaltyWeight(weight float64) {
	tmc.penaltyWeight = weight
}

// AnalyzeTeamTravel provides detailed travel analysis for a team
func (tmc *TravelMinimizationConstraint) AnalyzeTeamTravel(draw *models.Draw, teamID int) TravelAnalysis {
	analysis := TravelAnalysis{
		TeamID:     teamID,
		TotalGames: 0,
		HomeGames:  0,
		AwayGames:  0,
		Streaks:    []ConsecutiveAwayStreak{},
	}

	teamMatches := tmc.getTeamMatchesByRound(draw, teamID)
	analysis.TotalGames = len(teamMatches)

	consecutiveAwayCount := 0
	streakStart := 0

	for round := 1; round <= draw.Rounds; round++ {
		match, exists := teamMatches[round]
		if !exists {
			// Bye round - end current streak if any
			if consecutiveAwayCount > 0 {
				analysis.Streaks = append(analysis.Streaks, ConsecutiveAwayStreak{
					StartRound:   streakStart,
					EndRound:     round - 1,
					Length:       consecutiveAwayCount,
					ExceedsLimit: consecutiveAwayCount > tmc.maxConsecutiveAway,
				})
				consecutiveAwayCount = 0
			}
			continue
		}

		// Check if this is a home or away game
		if isHome, _ := match.IsHomeGame(teamID); isHome {
			analysis.HomeGames++

			// End current away streak if any
			if consecutiveAwayCount > 0 {
				analysis.Streaks = append(analysis.Streaks, ConsecutiveAwayStreak{
					StartRound:   streakStart,
					EndRound:     round - 1,
					Length:       consecutiveAwayCount,
					ExceedsLimit: consecutiveAwayCount > tmc.maxConsecutiveAway,
				})
				consecutiveAwayCount = 0
			}
		} else {
			analysis.AwayGames++

			if consecutiveAwayCount == 0 {
				streakStart = round
			}
			consecutiveAwayCount++
		}
	}

	// Handle final streak if it ends with the season
	if consecutiveAwayCount > 0 {
		analysis.Streaks = append(analysis.Streaks, ConsecutiveAwayStreak{
			StartRound:   streakStart,
			EndRound:     draw.Rounds,
			Length:       consecutiveAwayCount,
			ExceedsLimit: consecutiveAwayCount > tmc.maxConsecutiveAway,
		})
	}

	// Calculate longest streak
	analysis.LongestAwayStreak = 0
	analysis.ViolatingStreaks = 0
	for _, streak := range analysis.Streaks {
		if streak.Length > analysis.LongestAwayStreak {
			analysis.LongestAwayStreak = streak.Length
		}
		if streak.ExceedsLimit {
			analysis.ViolatingStreaks++
		}
	}

	return analysis
}

// TravelAnalysis contains detailed travel analysis for a team
type TravelAnalysis struct {
	TeamID            int                     `json:"team_id"`
	TotalGames        int                     `json:"total_games"`
	HomeGames         int                     `json:"home_games"`
	AwayGames         int                     `json:"away_games"`
	LongestAwayStreak int                     `json:"longest_away_streak"`
	ViolatingStreaks  int                     `json:"violating_streaks"`
	Streaks           []ConsecutiveAwayStreak `json:"streaks"`
}

// ConsecutiveAwayStreak represents a streak of consecutive away games
type ConsecutiveAwayStreak struct {
	StartRound   int  `json:"start_round"`
	EndRound     int  `json:"end_round"`
	Length       int  `json:"length"`
	ExceedsLimit bool `json:"exceeds_limit"`
}

// GetAllTeamTravelAnalysis returns travel analysis for all teams
func (tmc *TravelMinimizationConstraint) GetAllTeamTravelAnalysis(draw *models.Draw) []TravelAnalysis {
	teams := tmc.getUniqueTeams(draw)
	analyses := make([]TravelAnalysis, len(teams))

	for i, teamID := range teams {
		analyses[i] = tmc.AnalyzeTeamTravel(draw, teamID)
	}

	return analyses
}

// GetWorstTravelTeams returns teams with the worst travel burden
func (tmc *TravelMinimizationConstraint) GetWorstTravelTeams(draw *models.Draw, limit int) []TravelAnalysis {
	analyses := tmc.GetAllTeamTravelAnalysis(draw)

	// Sort by longest streak (descending) and then by violating streaks
	for i := 0; i < len(analyses)-1; i++ {
		for j := i + 1; j < len(analyses); j++ {
			if analyses[i].LongestAwayStreak < analyses[j].LongestAwayStreak ||
				(analyses[i].LongestAwayStreak == analyses[j].LongestAwayStreak &&
					analyses[i].ViolatingStreaks < analyses[j].ViolatingStreaks) {
				analyses[i], analyses[j] = analyses[j], analyses[i]
			}
		}
	}

	if limit > len(analyses) {
		limit = len(analyses)
	}

	return analyses[:limit]
}

// CalculateTotalTravelDistance calculates total travel distance (requires venue coordinates)
func (tmc *TravelMinimizationConstraint) CalculateTravelDistance(draw *models.Draw, teamID int) float64 {
	teamMatches := tmc.getTeamMatchesByRound(draw, teamID)
	totalDistance := 0.0

	var previousVenueID *int

	for round := 1; round <= draw.Rounds; round++ {
		match, exists := teamMatches[round]
		if !exists {
			continue // Bye round
		}

		// For away games, calculate travel distance
		if isHome, _ := match.IsHomeGame(teamID); !isHome {
			if previousVenueID != nil && match.VenueID != nil {
				// Calculate distance between venues (placeholder - would need actual coordinates)
				distance := tmc.calculateVenueDistance(*previousVenueID, *match.VenueID)
				totalDistance += distance
			}
			if match.VenueID != nil {
				previousVenueID = match.VenueID
			}
		} else {
			// Home game - reset to home venue
			// This would need the team's home venue ID
			previousVenueID = nil
		}
	}

	return totalDistance
}

// calculateVenueDistance is a placeholder for actual distance calculation
func (tmc *TravelMinimizationConstraint) calculateVenueDistance(venue1ID, venue2ID int) float64 {
	// This would use actual venue coordinates from the database
	// For now, return a simple placeholder based on venue ID difference
	// TODO: implement this properly
	return math.Abs(float64(venue1ID-venue2ID)) * 100 // Placeholder
}
