package constraints

import (
	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// PrimeTimeSpreadConstraint ensures fair distribution of prime-time games
type PrimeTimeSpreadConstraint struct {
	BaseConstraint
	targetPrimeTimeRatio float64 // Target ratio of prime time games per team
	maxDeviation         float64 // Maximum allowed deviation from target
}

// NewPrimeTimeSpreadConstraint creates a new prime time spread constraint
func NewPrimeTimeSpreadConstraint(targetRatio float64, maxDeviation float64) *PrimeTimeSpreadConstraint {
	return &PrimeTimeSpreadConstraint{
		BaseConstraint: NewBaseConstraint(
			"PrimeTimeSpread",
			"Distribute prime-time games fairly across all teams",
			false, // This is a soft constraint
		),
		targetPrimeTimeRatio: targetRatio,
		maxDeviation:         maxDeviation,
	}
}

// Validate always returns nil for soft constraints
func (ptsc *PrimeTimeSpreadConstraint) Validate(match *models.Match, draw *models.Draw) error {
	// Soft constraints don't have hard validation failures
	return nil
}

// Score calculates how well the draw distributes prime-time games
func (ptsc *PrimeTimeSpreadConstraint) Score(draw *models.Draw) float64 {
	teams := ptsc.getUniqueTeams(draw)
	if len(teams) == 0 {
		return 1.0
	}
	
	totalScore := 0.0
	
	for _, team := range teams {
		teamScore := ptsc.scoreTeamPrimeTimeDistribution(draw, team)
		totalScore += teamScore
	}
	
	return totalScore / float64(len(teams))
}

// scoreTeamPrimeTimeDistribution calculates prime time distribution score for a team
func (ptsc *PrimeTimeSpreadConstraint) scoreTeamPrimeTimeDistribution(draw *models.Draw, teamID int) float64 {
	teamMatches := draw.GetMatchesByTeam(teamID)
	if len(teamMatches) == 0 {
		return 1.0
	}
	
	primeTimeMatches := 0
	totalMatches := 0
	
	for _, match := range teamMatches {
		if !match.IsBye() {
			totalMatches++
			if match.IsPrimeTime {
				primeTimeMatches++
			}
		}
	}
	
	if totalMatches == 0 {
		return 1.0
	}
	
	// Calculate actual ratio
	actualRatio := float64(primeTimeMatches) / float64(totalMatches)
	
	// Calculate deviation from target
	deviation := actualRatio - ptsc.targetPrimeTimeRatio
	if deviation < 0 {
		deviation = -deviation
	}
	
	// Score based on how close to target ratio
	if deviation <= ptsc.maxDeviation {
		// Within acceptable range - score based on proximity to target
		return 1.0 - (deviation / ptsc.maxDeviation)
	} else {
		// Outside acceptable range - heavily penalized
		return 0.0
	}
}

// getUniqueTeams extracts all unique team IDs from the draw
func (ptsc *PrimeTimeSpreadConstraint) getUniqueTeams(draw *models.Draw) []int {
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

// GetTargetPrimeTimeRatio returns the target prime time ratio
func (ptsc *PrimeTimeSpreadConstraint) GetTargetPrimeTimeRatio() float64 {
	return ptsc.targetPrimeTimeRatio
}

// GetMaxDeviation returns the maximum allowed deviation
func (ptsc *PrimeTimeSpreadConstraint) GetMaxDeviation() float64 {
	return ptsc.maxDeviation
}

// SetTargetPrimeTimeRatio sets the target prime time ratio
func (ptsc *PrimeTimeSpreadConstraint) SetTargetPrimeTimeRatio(ratio float64) {
	ptsc.targetPrimeTimeRatio = ratio
}

// SetMaxDeviation sets the maximum allowed deviation
func (ptsc *PrimeTimeSpreadConstraint) SetMaxDeviation(deviation float64) {
	ptsc.maxDeviation = deviation
}

// AnalyzeTeamPrimeTimeDistribution provides detailed analysis for a team
func (ptsc *PrimeTimeSpreadConstraint) AnalyzeTeamPrimeTimeDistribution(draw *models.Draw, teamID int) PrimeTimeAnalysis {
	analysis := PrimeTimeAnalysis{
		TeamID:              teamID,
		TotalMatches:        0,
		PrimeTimeMatches:    0,
		RegularMatches:      0,
		PrimeTimeRatio:      0.0,
		DeviationFromTarget: 0.0,
		WithinAcceptableRange: false,
		PrimeTimeRounds:     []int{},
	}
	
	teamMatches := draw.GetMatchesByTeam(teamID)
	
	for _, match := range teamMatches {
		if !match.IsBye() {
			analysis.TotalMatches++
			if match.IsPrimeTime {
				analysis.PrimeTimeMatches++
				analysis.PrimeTimeRounds = append(analysis.PrimeTimeRounds, match.Round)
			} else {
				analysis.RegularMatches++
			}
		}
	}
	
	if analysis.TotalMatches > 0 {
		analysis.PrimeTimeRatio = float64(analysis.PrimeTimeMatches) / float64(analysis.TotalMatches)
		analysis.DeviationFromTarget = analysis.PrimeTimeRatio - ptsc.targetPrimeTimeRatio
		
		if analysis.DeviationFromTarget < 0 {
			analysis.DeviationFromTarget = -analysis.DeviationFromTarget
		}
		
		analysis.WithinAcceptableRange = analysis.DeviationFromTarget <= ptsc.maxDeviation
	}
	
	return analysis
}

// PrimeTimeAnalysis contains detailed prime time distribution analysis
type PrimeTimeAnalysis struct {
	TeamID                int     `json:"team_id"`
	TotalMatches          int     `json:"total_matches"`
	PrimeTimeMatches      int     `json:"prime_time_matches"`
	RegularMatches        int     `json:"regular_matches"`
	PrimeTimeRatio        float64 `json:"prime_time_ratio"`
	DeviationFromTarget   float64 `json:"deviation_from_target"`
	WithinAcceptableRange bool    `json:"within_acceptable_range"`
	PrimeTimeRounds       []int   `json:"prime_time_rounds"`
}

// GetAllTeamPrimeTimeAnalysis returns prime time analysis for all teams
func (ptsc *PrimeTimeSpreadConstraint) GetAllTeamPrimeTimeAnalysis(draw *models.Draw) []PrimeTimeAnalysis {
	teams := ptsc.getUniqueTeams(draw)
	analyses := make([]PrimeTimeAnalysis, len(teams))
	
	for i, teamID := range teams {
		analyses[i] = ptsc.AnalyzeTeamPrimeTimeDistribution(draw, teamID)
	}
	
	return analyses
}

// GetTeamsWithPoorDistribution returns teams with poor prime time distribution
func (ptsc *PrimeTimeSpreadConstraint) GetTeamsWithPoorDistribution(draw *models.Draw) []PrimeTimeAnalysis {
	analyses := ptsc.GetAllTeamPrimeTimeAnalysis(draw)
	var poorDistribution []PrimeTimeAnalysis
	
	for _, analysis := range analyses {
		if !analysis.WithinAcceptableRange {
			poorDistribution = append(poorDistribution, analysis)
		}
	}
	
	return poorDistribution
}

// GetTeamsWithMostPrimeTime returns teams with the most prime time games
func (ptsc *PrimeTimeSpreadConstraint) GetTeamsWithMostPrimeTime(draw *models.Draw, limit int) []PrimeTimeAnalysis {
	analyses := ptsc.GetAllTeamPrimeTimeAnalysis(draw)
	
	// Sort by prime time ratio (descending)
	for i := 0; i < len(analyses)-1; i++ {
		for j := i + 1; j < len(analyses); j++ {
			if analyses[i].PrimeTimeRatio < analyses[j].PrimeTimeRatio {
				analyses[i], analyses[j] = analyses[j], analyses[i]
			}
		}
	}
	
	if limit > len(analyses) {
		limit = len(analyses)
	}
	
	return analyses[:limit]
}

// GetTeamsWithLeastPrimeTime returns teams with the least prime time games
func (ptsc *PrimeTimeSpreadConstraint) GetTeamsWithLeastPrimeTime(draw *models.Draw, limit int) []PrimeTimeAnalysis {
	analyses := ptsc.GetAllTeamPrimeTimeAnalysis(draw)
	
	// Sort by prime time ratio (ascending)
	for i := 0; i < len(analyses)-1; i++ {
		for j := i + 1; j < len(analyses); j++ {
			if analyses[i].PrimeTimeRatio > analyses[j].PrimeTimeRatio {
				analyses[i], analyses[j] = analyses[j], analyses[i]
			}
		}
	}
	
	if limit > len(analyses) {
		limit = len(analyses)
	}
	
	return analyses[:limit]
}

// GetDrawPrimeTimeStatistics returns overall prime time statistics
func (ptsc *PrimeTimeSpreadConstraint) GetDrawPrimeTimeStatistics(draw *models.Draw) PrimeTimeStatistics {
	analyses := ptsc.GetAllTeamPrimeTimeAnalysis(draw)
	
	stats := PrimeTimeStatistics{
		TotalTeams:           len(analyses),
		TotalMatches:         0,
		TotalPrimeTimeMatches: 0,
		AveragePrimeTimeRatio: 0.0,
		MinPrimeTimeRatio:     1.0,
		MaxPrimeTimeRatio:     0.0,
		TeamsWithinRange:      0,
		TeamsOutsideRange:     0,
	}
	
	totalRatio := 0.0
	
	for _, analysis := range analyses {
		stats.TotalMatches += analysis.TotalMatches
		stats.TotalPrimeTimeMatches += analysis.PrimeTimeMatches
		totalRatio += analysis.PrimeTimeRatio
		
		if analysis.PrimeTimeRatio < stats.MinPrimeTimeRatio {
			stats.MinPrimeTimeRatio = analysis.PrimeTimeRatio
		}
		if analysis.PrimeTimeRatio > stats.MaxPrimeTimeRatio {
			stats.MaxPrimeTimeRatio = analysis.PrimeTimeRatio
		}
		
		if analysis.WithinAcceptableRange {
			stats.TeamsWithinRange++
		} else {
			stats.TeamsOutsideRange++
		}
	}
	
	if len(analyses) > 0 {
		stats.AveragePrimeTimeRatio = totalRatio / float64(len(analyses))
	}
	
	return stats
}

// PrimeTimeStatistics contains overall prime time distribution statistics
type PrimeTimeStatistics struct {
	TotalTeams            int     `json:"total_teams"`
	TotalMatches          int     `json:"total_matches"`
	TotalPrimeTimeMatches int     `json:"total_prime_time_matches"`
	AveragePrimeTimeRatio float64 `json:"average_prime_time_ratio"`
	MinPrimeTimeRatio     float64 `json:"min_prime_time_ratio"`
	MaxPrimeTimeRatio     float64 `json:"max_prime_time_ratio"`
	TeamsWithinRange      int     `json:"teams_within_range"`
	TeamsOutsideRange     int     `json:"teams_outside_range"`
}

// GetRoundPrimeTimeDistribution returns prime time distribution by round
func (ptsc *PrimeTimeSpreadConstraint) GetRoundPrimeTimeDistribution(draw *models.Draw) map[int]RoundPrimeTimeInfo {
	roundInfo := make(map[int]RoundPrimeTimeInfo)
	
	for round := 1; round <= draw.Rounds; round++ {
		roundMatches := draw.GetMatchesByRound(round)
		
		info := RoundPrimeTimeInfo{
			Round:                round,
			TotalMatches:         0,
			PrimeTimeMatches:     0,
			RegularMatches:       0,
			PrimeTimeRatio:       0.0,
		}
		
		for _, match := range roundMatches {
			if !match.IsBye() {
				info.TotalMatches++
				if match.IsPrimeTime {
					info.PrimeTimeMatches++
				} else {
					info.RegularMatches++
				}
			}
		}
		
		if info.TotalMatches > 0 {
			info.PrimeTimeRatio = float64(info.PrimeTimeMatches) / float64(info.TotalMatches)
		}
		
		roundInfo[round] = info
	}
	
	return roundInfo
}

// RoundPrimeTimeInfo contains prime time information for a specific round
type RoundPrimeTimeInfo struct {
	Round            int     `json:"round"`
	TotalMatches     int     `json:"total_matches"`
	PrimeTimeMatches int     `json:"prime_time_matches"`
	RegularMatches   int     `json:"regular_matches"`
	PrimeTimeRatio   float64 `json:"prime_time_ratio"`
}

// SuggestPrimeTimeAdjustments suggests adjustments to improve distribution
func (ptsc *PrimeTimeSpreadConstraint) SuggestPrimeTimeAdjustments(draw *models.Draw) []PrimeTimeAdjustment {
	var adjustments []PrimeTimeAdjustment
	
	// Get teams with poor distribution
	poorDistribution := ptsc.GetTeamsWithPoorDistribution(draw)
	
	for _, analysis := range poorDistribution {
		if analysis.PrimeTimeRatio > ptsc.targetPrimeTimeRatio + ptsc.maxDeviation {
			// Team has too many prime time games
			adjustments = append(adjustments, PrimeTimeAdjustment{
				TeamID:     analysis.TeamID,
				Action:     "REDUCE",
				CurrentRatio: analysis.PrimeTimeRatio,
				TargetRatio:  ptsc.targetPrimeTimeRatio,
				Suggestion:   "Move some prime time games to regular time slots",
			})
		} else if analysis.PrimeTimeRatio < ptsc.targetPrimeTimeRatio - ptsc.maxDeviation {
			// Team has too few prime time games
			adjustments = append(adjustments, PrimeTimeAdjustment{
				TeamID:     analysis.TeamID,
				Action:     "INCREASE",
				CurrentRatio: analysis.PrimeTimeRatio,
				TargetRatio:  ptsc.targetPrimeTimeRatio,
				Suggestion:   "Move some regular games to prime time slots",
			})
		}
	}
	
	return adjustments
}

// PrimeTimeAdjustment represents a suggested adjustment to prime time distribution
type PrimeTimeAdjustment struct {
	TeamID       int     `json:"team_id"`
	Action       string  `json:"action"` // "INCREASE" or "REDUCE"
	CurrentRatio float64 `json:"current_ratio"`
	TargetRatio  float64 `json:"target_ratio"`
	Suggestion   string  `json:"suggestion"`
}