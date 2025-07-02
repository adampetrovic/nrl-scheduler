package constraints

import (
	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// HomeAwayBalanceConstraint ensures fair distribution of home and away games
type HomeAwayBalanceConstraint struct {
	BaseConstraint
	maxDeviation float64 // Maximum allowed deviation from 50/50 split
}

// NewHomeAwayBalanceConstraint creates a new home/away balance constraint
func NewHomeAwayBalanceConstraint(maxDeviation float64) *HomeAwayBalanceConstraint {
	return &HomeAwayBalanceConstraint{
		BaseConstraint: NewBaseConstraint(
			"HomeAwayBalance",
			"Balance home and away games fairly for all teams",
			false, // This is a soft constraint
		),
		maxDeviation: maxDeviation,
	}
}

// Validate always returns nil for soft constraints
func (habc *HomeAwayBalanceConstraint) Validate(match *models.Match, draw *models.Draw) error {
	// Soft constraints don't have hard validation failures
	return nil
}

// Score calculates how well the draw balances home and away games
func (habc *HomeAwayBalanceConstraint) Score(draw *models.Draw) float64 {
	teams := habc.getUniqueTeams(draw)
	if len(teams) == 0 {
		return 1.0
	}
	
	totalScore := 0.0
	
	for _, team := range teams {
		teamScore := habc.scoreTeamBalance(draw, team)
		totalScore += teamScore
	}
	
	return totalScore / float64(len(teams))
}

// scoreTeamBalance calculates the home/away balance score for a specific team
func (habc *HomeAwayBalanceConstraint) scoreTeamBalance(draw *models.Draw, teamID int) float64 {
	teamMatches := draw.GetMatchesByTeam(teamID)
	if len(teamMatches) == 0 {
		return 1.0
	}
	
	homeGames := 0
	awayGames := 0
	
	for _, match := range teamMatches {
		if !match.IsBye() {
			if isHome, _ := match.IsHomeGame(teamID); isHome {
				homeGames++
			} else {
				awayGames++
			}
		}
	}
	
	totalGames := homeGames + awayGames
	if totalGames == 0 {
		return 1.0
	}
	
	// Calculate the deviation from perfect balance (50/50)
	homeRatio := float64(homeGames) / float64(totalGames)
	deviation := homeRatio - 0.5
	if deviation < 0 {
		deviation = -deviation
	}
	
	// Score based on how close to perfect balance
	if deviation <= habc.maxDeviation {
		// Within acceptable range - score based on proximity to perfect balance
		return 1.0 - (deviation / habc.maxDeviation)
	} else {
		// Outside acceptable range - heavily penalized
		return 0.0
	}
}

// getUniqueTeams extracts all unique team IDs from the draw
func (habc *HomeAwayBalanceConstraint) getUniqueTeams(draw *models.Draw) []int {
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

// GetMaxDeviation returns the maximum allowed deviation from 50/50 balance
func (habc *HomeAwayBalanceConstraint) GetMaxDeviation() float64 {
	return habc.maxDeviation
}

// SetMaxDeviation sets the maximum allowed deviation from 50/50 balance
func (habc *HomeAwayBalanceConstraint) SetMaxDeviation(deviation float64) {
	habc.maxDeviation = deviation
}

// AnalyzeTeamHomeAwayBalance provides detailed balance analysis for a team
func (habc *HomeAwayBalanceConstraint) AnalyzeTeamHomeAwayBalance(draw *models.Draw, teamID int) HomeAwayAnalysis {
	analysis := HomeAwayAnalysis{
		TeamID:             teamID,
		TotalGames:         0,
		HomeGames:          0,
		AwayGames:          0,
		HomeRatio:          0.0,
		AwayRatio:          0.0,
		DeviationFromBalance: 0.0,
		WithinAcceptableRange: false,
		HomeRounds:         []int{},
		AwayRounds:         []int{},
	}
	
	teamMatches := draw.GetMatchesByTeam(teamID)
	
	for _, match := range teamMatches {
		if !match.IsBye() {
			analysis.TotalGames++
			if isHome, _ := match.IsHomeGame(teamID); isHome {
				analysis.HomeGames++
				analysis.HomeRounds = append(analysis.HomeRounds, match.Round)
			} else {
				analysis.AwayGames++
				analysis.AwayRounds = append(analysis.AwayRounds, match.Round)
			}
		}
	}
	
	if analysis.TotalGames > 0 {
		analysis.HomeRatio = float64(analysis.HomeGames) / float64(analysis.TotalGames)
		analysis.AwayRatio = float64(analysis.AwayGames) / float64(analysis.TotalGames)
		analysis.DeviationFromBalance = analysis.HomeRatio - 0.5
		
		if analysis.DeviationFromBalance < 0 {
			analysis.DeviationFromBalance = -analysis.DeviationFromBalance
		}
		
		analysis.WithinAcceptableRange = analysis.DeviationFromBalance <= habc.maxDeviation
	}
	
	return analysis
}

// HomeAwayAnalysis contains detailed home/away balance analysis
type HomeAwayAnalysis struct {
	TeamID                int     `json:"team_id"`
	TotalGames            int     `json:"total_games"`
	HomeGames             int     `json:"home_games"`
	AwayGames             int     `json:"away_games"`
	HomeRatio             float64 `json:"home_ratio"`
	AwayRatio             float64 `json:"away_ratio"`
	DeviationFromBalance  float64 `json:"deviation_from_balance"`
	WithinAcceptableRange bool    `json:"within_acceptable_range"`
	HomeRounds            []int   `json:"home_rounds"`
	AwayRounds            []int   `json:"away_rounds"`
}

// GetAllTeamHomeAwayAnalysis returns balance analysis for all teams
func (habc *HomeAwayBalanceConstraint) GetAllTeamHomeAwayAnalysis(draw *models.Draw) []HomeAwayAnalysis {
	teams := habc.getUniqueTeams(draw)
	analyses := make([]HomeAwayAnalysis, len(teams))
	
	for i, teamID := range teams {
		analyses[i] = habc.AnalyzeTeamHomeAwayBalance(draw, teamID)
	}
	
	return analyses
}

// GetTeamsWithPoorBalance returns teams with poor home/away balance
func (habc *HomeAwayBalanceConstraint) GetTeamsWithPoorBalance(draw *models.Draw) []HomeAwayAnalysis {
	analyses := habc.GetAllTeamHomeAwayAnalysis(draw)
	var poorBalance []HomeAwayAnalysis
	
	for _, analysis := range analyses {
		if !analysis.WithinAcceptableRange {
			poorBalance = append(poorBalance, analysis)
		}
	}
	
	return poorBalance
}

// GetTeamsWithMostHomeGames returns teams with the highest home game ratio
func (habc *HomeAwayBalanceConstraint) GetTeamsWithMostHomeGames(draw *models.Draw, limit int) []HomeAwayAnalysis {
	analyses := habc.GetAllTeamHomeAwayAnalysis(draw)
	
	// Sort by home ratio (descending)
	for i := 0; i < len(analyses)-1; i++ {
		for j := i + 1; j < len(analyses); j++ {
			if analyses[i].HomeRatio < analyses[j].HomeRatio {
				analyses[i], analyses[j] = analyses[j], analyses[i]
			}
		}
	}
	
	if limit > len(analyses) {
		limit = len(analyses)
	}
	
	return analyses[:limit]
}

// GetTeamsWithMostAwayGames returns teams with the highest away game ratio
func (habc *HomeAwayBalanceConstraint) GetTeamsWithMostAwayGames(draw *models.Draw, limit int) []HomeAwayAnalysis {
	analyses := habc.GetAllTeamHomeAwayAnalysis(draw)
	
	// Sort by away ratio (descending)
	for i := 0; i < len(analyses)-1; i++ {
		for j := i + 1; j < len(analyses); j++ {
			if analyses[i].AwayRatio < analyses[j].AwayRatio {
				analyses[i], analyses[j] = analyses[j], analyses[i]
			}
		}
	}
	
	if limit > len(analyses) {
		limit = len(analyses)
	}
	
	return analyses[:limit]
}

// GetDrawBalanceStatistics returns overall balance statistics
func (habc *HomeAwayBalanceConstraint) GetDrawBalanceStatistics(draw *models.Draw) HomeAwayStatistics {
	analyses := habc.GetAllTeamHomeAwayAnalysis(draw)
	
	stats := HomeAwayStatistics{
		TotalTeams:           len(analyses),
		TotalMatches:         0,
		TotalHomeGames:       0,
		TotalAwayGames:       0,
		AverageHomeRatio:     0.0,
		MinHomeRatio:         1.0,
		MaxHomeRatio:         0.0,
		TeamsWithinRange:     0,
		TeamsOutsideRange:    0,
		AverageDeviation:     0.0,
	}
	
	totalHomeRatio := 0.0
	totalDeviation := 0.0
	
	for _, analysis := range analyses {
		stats.TotalMatches += analysis.TotalGames
		stats.TotalHomeGames += analysis.HomeGames
		stats.TotalAwayGames += analysis.AwayGames
		totalHomeRatio += analysis.HomeRatio
		totalDeviation += analysis.DeviationFromBalance
		
		if analysis.HomeRatio < stats.MinHomeRatio {
			stats.MinHomeRatio = analysis.HomeRatio
		}
		if analysis.HomeRatio > stats.MaxHomeRatio {
			stats.MaxHomeRatio = analysis.HomeRatio
		}
		
		if analysis.WithinAcceptableRange {
			stats.TeamsWithinRange++
		} else {
			stats.TeamsOutsideRange++
		}
	}
	
	if len(analyses) > 0 {
		stats.AverageHomeRatio = totalHomeRatio / float64(len(analyses))
		stats.AverageDeviation = totalDeviation / float64(len(analyses))
	}
	
	return stats
}

// HomeAwayStatistics contains overall home/away balance statistics
type HomeAwayStatistics struct {
	TotalTeams        int     `json:"total_teams"`
	TotalMatches      int     `json:"total_matches"`
	TotalHomeGames    int     `json:"total_home_games"`
	TotalAwayGames    int     `json:"total_away_games"`
	AverageHomeRatio  float64 `json:"average_home_ratio"`
	MinHomeRatio      float64 `json:"min_home_ratio"`
	MaxHomeRatio      float64 `json:"max_home_ratio"`
	TeamsWithinRange  int     `json:"teams_within_range"`
	TeamsOutsideRange int     `json:"teams_outside_range"`
	AverageDeviation  float64 `json:"average_deviation"`
}

// AnalyzeHomeAwaySequences analyzes consecutive home/away game patterns
func (habc *HomeAwayBalanceConstraint) AnalyzeHomeAwaySequences(draw *models.Draw, teamID int) SequenceAnalysis {
	analysis := SequenceAnalysis{
		TeamID:              teamID,
		LongestHomeSequence: 0,
		LongestAwaySequence: 0,
		HomeSequences:       []Sequence{},
		AwaySequences:       []Sequence{},
	}
	
	teamMatches := habc.getTeamMatchesByRound(draw, teamID)
	
	currentSequence := ""
	sequenceStart := 0
	sequenceLength := 0
	
	for round := 1; round <= draw.Rounds; round++ {
		match, exists := teamMatches[round]
		if !exists || match.IsBye() {
			// End current sequence if any
			if sequenceLength > 0 {
				habc.recordSequence(&analysis, currentSequence, sequenceStart, round-1, sequenceLength)
				sequenceLength = 0
			}
			continue
		}
		
		gameType := "away"
		if isHome, _ := match.IsHomeGame(teamID); isHome {
			gameType = "home"
		}
		
		if gameType == currentSequence {
			// Continue current sequence
			sequenceLength++
		} else {
			// End previous sequence and start new one
			if sequenceLength > 0 {
				habc.recordSequence(&analysis, currentSequence, sequenceStart, round-1, sequenceLength)
			}
			currentSequence = gameType
			sequenceStart = round
			sequenceLength = 1
		}
	}
	
	// Handle final sequence
	if sequenceLength > 0 {
		habc.recordSequence(&analysis, currentSequence, sequenceStart, draw.Rounds, sequenceLength)
	}
	
	return analysis
}

// recordSequence records a sequence in the analysis
func (habc *HomeAwayBalanceConstraint) recordSequence(analysis *SequenceAnalysis, 
	sequenceType string, start, end, length int) {
	
	seq := Sequence{
		Type:      sequenceType,
		Start:     start,
		End:       end,
		Length:    length,
	}
	
	if sequenceType == "home" {
		analysis.HomeSequences = append(analysis.HomeSequences, seq)
		if length > analysis.LongestHomeSequence {
			analysis.LongestHomeSequence = length
		}
	} else {
		analysis.AwaySequences = append(analysis.AwaySequences, seq)
		if length > analysis.LongestAwaySequence {
			analysis.LongestAwaySequence = length
		}
	}
}

// getTeamMatchesByRound returns team matches organized by round
func (habc *HomeAwayBalanceConstraint) getTeamMatchesByRound(draw *models.Draw, teamID int) map[int]*models.Match {
	matches := make(map[int]*models.Match)
	
	for _, match := range draw.Matches {
		if match.HasTeam(teamID) {
			matches[match.Round] = match
		}
	}
	
	return matches
}

// SequenceAnalysis contains analysis of consecutive home/away sequences
type SequenceAnalysis struct {
	TeamID              int        `json:"team_id"`
	LongestHomeSequence int        `json:"longest_home_sequence"`
	LongestAwaySequence int        `json:"longest_away_sequence"`
	HomeSequences       []Sequence `json:"home_sequences"`
	AwaySequences       []Sequence `json:"away_sequences"`
}

// Sequence represents a consecutive sequence of home or away games
type Sequence struct {
	Type   string `json:"type"`   // "home" or "away"
	Start  int    `json:"start"`  // Starting round
	End    int    `json:"end"`    // Ending round
	Length int    `json:"length"` // Number of games in sequence
}

// SuggestBalanceAdjustments suggests adjustments to improve home/away balance
func (habc *HomeAwayBalanceConstraint) SuggestBalanceAdjustments(draw *models.Draw) []BalanceAdjustment {
	var adjustments []BalanceAdjustment
	
	// Get teams with poor balance
	poorBalance := habc.GetTeamsWithPoorBalance(draw)
	
	for _, analysis := range poorBalance {
		if analysis.HomeRatio > 0.5+habc.maxDeviation {
			// Team has too many home games
			adjustments = append(adjustments, BalanceAdjustment{
				TeamID:      analysis.TeamID,
				Action:      "REDUCE_HOME",
				CurrentHomeRatio: analysis.HomeRatio,
				TargetHomeRatio:  0.5,
				Suggestion:  "Convert some home games to away games or swap venues",
			})
		} else if analysis.HomeRatio < 0.5-habc.maxDeviation {
			// Team has too few home games
			adjustments = append(adjustments, BalanceAdjustment{
				TeamID:      analysis.TeamID,
				Action:      "INCREASE_HOME",
				CurrentHomeRatio: analysis.HomeRatio,
				TargetHomeRatio:  0.5,
				Suggestion:  "Convert some away games to home games or swap venues",
			})
		}
	}
	
	return adjustments
}

// BalanceAdjustment represents a suggested adjustment to home/away balance
type BalanceAdjustment struct {
	TeamID           int     `json:"team_id"`
	Action           string  `json:"action"` // "INCREASE_HOME" or "REDUCE_HOME"
	CurrentHomeRatio float64 `json:"current_home_ratio"`
	TargetHomeRatio  float64 `json:"target_home_ratio"`
	Suggestion       string  `json:"suggestion"`
}