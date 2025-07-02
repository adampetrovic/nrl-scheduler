package constraints

import (
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// RestPeriodConstraint ensures minimum rest days between matches
type RestPeriodConstraint struct {
	BaseConstraint
	minRestDays   int
	penaltyWeight float64
}

// NewRestPeriodConstraint creates a new rest period constraint
func NewRestPeriodConstraint(minRestDays int) *RestPeriodConstraint {
	return &RestPeriodConstraint{
		BaseConstraint: NewBaseConstraint(
			"RestPeriod",
			"Ensure minimum rest days between matches for player welfare",
			false, // This is a soft constraint
		),
		minRestDays:   minRestDays,
		penaltyWeight: 1.0,
	}
}

// Validate always returns nil for soft constraints
func (rpc *RestPeriodConstraint) Validate(match *models.Match, draw *models.Draw) error {
	// Soft constraints don't have hard validation failures
	return nil
}

// Score calculates how well the draw satisfies rest period requirements
func (rpc *RestPeriodConstraint) Score(draw *models.Draw) float64 {
	teams := rpc.getUniqueTeams(draw)
	if len(teams) == 0 {
		return 1.0
	}
	
	totalScore := 0.0
	
	for _, team := range teams {
		teamScore := rpc.scoreTeamRestPeriods(draw, team)
		totalScore += teamScore
	}
	
	return totalScore / float64(len(teams))
}

// scoreTeamRestPeriods calculates the rest period score for a specific team
func (rpc *RestPeriodConstraint) scoreTeamRestPeriods(draw *models.Draw, teamID int) float64 {
	teamMatches := rpc.getTeamMatchesWithDates(draw, teamID)
	if len(teamMatches) <= 1 {
		return 1.0 // Can't violate rest periods with 0 or 1 matches
	}
	
	violations := 0
	totalGaps := 0
	
	// Sort matches by date
	sortedMatches := rpc.sortMatchesByDate(teamMatches)
	
	// Check rest periods between consecutive matches
	for i := 1; i < len(sortedMatches); i++ {
		prevMatch := sortedMatches[i-1]
		currentMatch := sortedMatches[i]
		
		if prevMatch.MatchDate != nil && currentMatch.MatchDate != nil {
			restDays := rpc.calculateRestDays(*prevMatch.MatchDate, *currentMatch.MatchDate)
			totalGaps++
			
			if restDays < rpc.minRestDays {
				violations++
			}
		}
	}
	
	if totalGaps == 0 {
		return 1.0 // No gaps to evaluate
	}
	
	// Return percentage of adequate rest periods
	return float64(totalGaps-violations) / float64(totalGaps)
}

// getUniqueTeams extracts all unique team IDs from the draw
func (rpc *RestPeriodConstraint) getUniqueTeams(draw *models.Draw) []int {
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

// getTeamMatchesWithDates returns team matches that have scheduled dates
func (rpc *RestPeriodConstraint) getTeamMatchesWithDates(draw *models.Draw, teamID int) []*models.Match {
	var matches []*models.Match
	
	for _, match := range draw.Matches {
		if match.HasTeam(teamID) && match.MatchDate != nil {
			matches = append(matches, match)
		}
	}
	
	return matches
}

// sortMatchesByDate sorts matches by their scheduled date
func (rpc *RestPeriodConstraint) sortMatchesByDate(matches []*models.Match) []*models.Match {
	// Create a copy to avoid modifying the original slice
	sorted := make([]*models.Match, len(matches))
	copy(sorted, matches)
	
	// Simple bubble sort by date
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].MatchDate != nil && sorted[j].MatchDate != nil {
				if sorted[i].MatchDate.After(*sorted[j].MatchDate) {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	}
	
	return sorted
}

// calculateRestDays calculates the number of rest days between two match dates
func (rpc *RestPeriodConstraint) calculateRestDays(date1, date2 time.Time) int {
	// Ensure date1 is before date2
	if date1.After(date2) {
		date1, date2 = date2, date1
	}
	
	// Calculate the difference in days
	duration := date2.Sub(date1)
	days := int(duration.Hours() / 24)
	
	// Subtract 1 because the day of the second match doesn't count as rest
	return days - 1
}

// GetMinRestDays returns the minimum required rest days
func (rpc *RestPeriodConstraint) GetMinRestDays() int {
	return rpc.minRestDays
}

// SetPenaltyWeight sets the penalty weight for inadequate rest periods
func (rpc *RestPeriodConstraint) SetPenaltyWeight(weight float64) {
	rpc.penaltyWeight = weight
}

// AnalyzeTeamRestPeriods provides detailed rest period analysis for a team
func (rpc *RestPeriodConstraint) AnalyzeTeamRestPeriods(draw *models.Draw, teamID int) RestPeriodAnalysis {
	analysis := RestPeriodAnalysis{
		TeamID:              teamID,
		TotalMatches:        0,
		ScheduledMatches:    0,
		AdequateRestPeriods: 0,
		ShortRestPeriods:    0,
		RestPeriods:         []RestPeriod{},
	}
	
	teamMatches := draw.GetMatchesByTeam(teamID)
	analysis.TotalMatches = len(teamMatches)
	
	scheduledMatches := rpc.getTeamMatchesWithDates(draw, teamID)
	analysis.ScheduledMatches = len(scheduledMatches)
	
	if len(scheduledMatches) <= 1 {
		return analysis // Can't analyze rest periods with 0 or 1 scheduled matches
	}
	
	sortedMatches := rpc.sortMatchesByDate(scheduledMatches)
	
	// Analyze rest periods between consecutive matches
	for i := 1; i < len(sortedMatches); i++ {
		prevMatch := sortedMatches[i-1]
		currentMatch := sortedMatches[i]
		
		restDays := rpc.calculateRestDays(*prevMatch.MatchDate, *currentMatch.MatchDate)
		
		restPeriod := RestPeriod{
			FromMatchID:  prevMatch.ID,
			ToMatchID:    currentMatch.ID,
			FromDate:     *prevMatch.MatchDate,
			ToDate:       *currentMatch.MatchDate,
			RestDays:     restDays,
			IsAdequate:   restDays >= rpc.minRestDays,
		}
		
		analysis.RestPeriods = append(analysis.RestPeriods, restPeriod)
		
		if restPeriod.IsAdequate {
			analysis.AdequateRestPeriods++
		} else {
			analysis.ShortRestPeriods++
		}
	}
	
	return analysis
}

// RestPeriodAnalysis contains detailed rest period analysis for a team
type RestPeriodAnalysis struct {
	TeamID              int          `json:"team_id"`
	TotalMatches        int          `json:"total_matches"`
	ScheduledMatches    int          `json:"scheduled_matches"`
	AdequateRestPeriods int          `json:"adequate_rest_periods"`
	ShortRestPeriods    int          `json:"short_rest_periods"`
	RestPeriods         []RestPeriod `json:"rest_periods"`
}

// RestPeriod represents the rest period between two matches
type RestPeriod struct {
	FromMatchID int       `json:"from_match_id"`
	ToMatchID   int       `json:"to_match_id"`
	FromDate    time.Time `json:"from_date"`
	ToDate      time.Time `json:"to_date"`
	RestDays    int       `json:"rest_days"`
	IsAdequate  bool      `json:"is_adequate"`
}

// GetAllTeamRestAnalysis returns rest period analysis for all teams
func (rpc *RestPeriodConstraint) GetAllTeamRestAnalysis(draw *models.Draw) []RestPeriodAnalysis {
	teams := rpc.getUniqueTeams(draw)
	analyses := make([]RestPeriodAnalysis, len(teams))
	
	for i, teamID := range teams {
		analyses[i] = rpc.AnalyzeTeamRestPeriods(draw, teamID)
	}
	
	return analyses
}

// GetTeamsWithShortRest returns teams with inadequate rest periods
func (rpc *RestPeriodConstraint) GetTeamsWithShortRest(draw *models.Draw) []RestPeriodAnalysis {
	analyses := rpc.GetAllTeamRestAnalysis(draw)
	var teamsWithShortRest []RestPeriodAnalysis
	
	for _, analysis := range analyses {
		if analysis.ShortRestPeriods > 0 {
			teamsWithShortRest = append(teamsWithShortRest, analysis)
		}
	}
	
	return teamsWithShortRest
}

// GetShortestRestPeriods returns the shortest rest periods across all teams
func (rpc *RestPeriodConstraint) GetShortestRestPeriods(draw *models.Draw, limit int) []RestPeriod {
	var allRestPeriods []RestPeriod
	
	analyses := rpc.GetAllTeamRestAnalysis(draw)
	for _, analysis := range analyses {
		allRestPeriods = append(allRestPeriods, analysis.RestPeriods...)
	}
	
	// Sort by rest days (ascending)
	for i := 0; i < len(allRestPeriods)-1; i++ {
		for j := i + 1; j < len(allRestPeriods); j++ {
			if allRestPeriods[i].RestDays > allRestPeriods[j].RestDays {
				allRestPeriods[i], allRestPeriods[j] = allRestPeriods[j], allRestPeriods[i]
			}
		}
	}
	
	if limit > len(allRestPeriods) {
		limit = len(allRestPeriods)
	}
	
	return allRestPeriods[:limit]
}

// GetDrawRestStatistics returns overall rest period statistics for the draw
func (rpc *RestPeriodConstraint) GetDrawRestStatistics(draw *models.Draw) RestStatistics {
	analyses := rpc.GetAllTeamRestAnalysis(draw)
	
	stats := RestStatistics{
		TotalTeams:          len(analyses),
		TotalRestPeriods:    0,
		AdequateRestPeriods: 0,
		ShortRestPeriods:    0,
		AverageRestDays:     0.0,
		MinRestDays:         9999,
		MaxRestDays:         0,
	}
	
	totalRestDays := 0
	
	for _, analysis := range analyses {
		stats.TotalRestPeriods += len(analysis.RestPeriods)
		stats.AdequateRestPeriods += analysis.AdequateRestPeriods
		stats.ShortRestPeriods += analysis.ShortRestPeriods
		
		for _, restPeriod := range analysis.RestPeriods {
			totalRestDays += restPeriod.RestDays
			
			if restPeriod.RestDays < stats.MinRestDays {
				stats.MinRestDays = restPeriod.RestDays
			}
			if restPeriod.RestDays > stats.MaxRestDays {
				stats.MaxRestDays = restPeriod.RestDays
			}
		}
	}
	
	if stats.TotalRestPeriods > 0 {
		stats.AverageRestDays = float64(totalRestDays) / float64(stats.TotalRestPeriods)
	}
	
	if stats.MinRestDays == 9999 {
		stats.MinRestDays = 0
	}
	
	return stats
}

// RestStatistics contains overall rest period statistics
type RestStatistics struct {
	TotalTeams          int     `json:"total_teams"`
	TotalRestPeriods    int     `json:"total_rest_periods"`
	AdequateRestPeriods int     `json:"adequate_rest_periods"`
	ShortRestPeriods    int     `json:"short_rest_periods"`
	AverageRestDays     float64 `json:"average_rest_days"`
	MinRestDays         int     `json:"min_rest_days"`
	MaxRestDays         int     `json:"max_rest_days"`
}