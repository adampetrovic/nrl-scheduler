package constraints

import (
	"fmt"
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// TeamAvailabilityConstraint ensures teams are not scheduled on unavailable dates
type TeamAvailabilityConstraint struct {
	DateConstraint
	teamID int
}

// NewTeamAvailabilityConstraint creates a new team availability constraint
func NewTeamAvailabilityConstraint(teamID int, unavailableDates []time.Time) *TeamAvailabilityConstraint {
	return &TeamAvailabilityConstraint{
		DateConstraint: NewDateConstraint(
			"TeamAvailability",
			fmt.Sprintf("Team %d must not be scheduled on specified unavailable dates", teamID),
			true, // This is a hard constraint
			unavailableDates,
		),
		teamID: teamID,
	}
}

// Validate checks if a match violates the team availability constraint
func (tac *TeamAvailabilityConstraint) Validate(match *models.Match, draw *models.Draw) error {
	// Skip validation for bye matches
	if match.IsBye() {
		return nil
	}
	
	// Skip if this match doesn't involve our team
	if !match.HasTeam(tac.teamID) {
		return nil
	}
	
	// Skip if match doesn't have a date assigned yet
	if match.MatchDate == nil {
		return nil
	}
	
	// Check if the match date conflicts with unavailable dates
	if tac.IsDateUnavailable(*match.MatchDate) {
		return fmt.Errorf("team %d is not available on %s", 
			tac.teamID, match.MatchDate.Format("2006-01-02"))
	}
	
	return nil
}

// Score calculates how well the draw satisfies this constraint
func (tac *TeamAvailabilityConstraint) Score(draw *models.Draw) float64 {
	totalMatches := 0
	violatingMatches := 0
	
	for _, match := range draw.Matches {
		// Only consider matches involving this team
		if match.HasTeam(tac.teamID) && match.MatchDate != nil {
			totalMatches++
			if tac.IsDateUnavailable(*match.MatchDate) {
				violatingMatches++
			}
		}
	}
	
	// If no matches for this team, constraint is perfectly satisfied
	if totalMatches == 0 {
		return 1.0
	}
	
	// Return the percentage of non-violating matches
	return float64(totalMatches-violatingMatches) / float64(totalMatches)
}

// GetTeamID returns the team ID this constraint applies to
func (tac *TeamAvailabilityConstraint) GetTeamID() int {
	return tac.teamID
}

// GetUnavailableDatesForTeam returns unavailable dates for this team
func (tac *TeamAvailabilityConstraint) GetUnavailableDatesForTeam() []time.Time {
	return tac.GetUnavailableDates()
}

// ValidateTeamSchedule performs comprehensive validation for the team's entire schedule
func (tac *TeamAvailabilityConstraint) ValidateTeamSchedule(draw *models.Draw) []error {
	var errors []error
	
	teamMatches := draw.GetMatchesByTeam(tac.teamID)
	
	for _, match := range teamMatches {
		if err := tac.Validate(match, draw); err != nil {
			errors = append(errors, err)
		}
	}
	
	return errors
}

// GetConflictingMatches returns all matches that conflict with unavailable dates
func (tac *TeamAvailabilityConstraint) GetConflictingMatches(draw *models.Draw) []*models.Match {
	var conflictingMatches []*models.Match
	
	for _, match := range draw.Matches {
		if match.HasTeam(tac.teamID) && match.MatchDate != nil {
			if tac.IsDateUnavailable(*match.MatchDate) {
				conflictingMatches = append(conflictingMatches, match)
			}
		}
	}
	
	return conflictingMatches
}

// GetAvailableAlternatives suggests alternative dates for conflicting matches
func (tac *TeamAvailabilityConstraint) GetAvailableAlternatives(conflictingMatch *models.Match, 
	possibleDates []time.Time) []time.Time {
	
	var alternatives []time.Time
	
	for _, date := range possibleDates {
		if !tac.IsDateUnavailable(date) {
			alternatives = append(alternatives, date)
		}
	}
	
	return alternatives
}

// MultiTeamAvailabilityConstraint handles availability for multiple teams
type MultiTeamAvailabilityConstraint struct {
	BaseConstraint
	teamConstraints map[int]*TeamAvailabilityConstraint
}

// NewMultiTeamAvailabilityConstraint creates a constraint for multiple teams
func NewMultiTeamAvailabilityConstraint(teamAvailability map[int][]time.Time) *MultiTeamAvailabilityConstraint {
	constraints := make(map[int]*TeamAvailabilityConstraint)
	
	for teamID, unavailableDates := range teamAvailability {
		constraints[teamID] = NewTeamAvailabilityConstraint(teamID, unavailableDates)
	}
	
	return &MultiTeamAvailabilityConstraint{
		BaseConstraint: NewBaseConstraint(
			"MultiTeamAvailability",
			"Multiple teams must not be scheduled on their unavailable dates",
			true,
		),
		teamConstraints: constraints,
	}
}

// Validate checks if a match violates any team availability constraints
func (mtac *MultiTeamAvailabilityConstraint) Validate(match *models.Match, draw *models.Draw) error {
	for _, constraint := range mtac.teamConstraints {
		if err := constraint.Validate(match, draw); err != nil {
			return err
		}
	}
	return nil
}

// Score calculates the overall score across all team constraints
func (mtac *MultiTeamAvailabilityConstraint) Score(draw *models.Draw) float64 {
	if len(mtac.teamConstraints) == 0 {
		return 1.0
	}
	
	totalScore := 0.0
	for _, constraint := range mtac.teamConstraints {
		totalScore += constraint.Score(draw)
	}
	
	return totalScore / float64(len(mtac.teamConstraints))
}

// GetTeamConstraints returns individual team constraints
func (mtac *MultiTeamAvailabilityConstraint) GetTeamConstraints() map[int]*TeamAvailabilityConstraint {
	return mtac.teamConstraints
}

// AddTeamConstraint adds a new team availability constraint
func (mtac *MultiTeamAvailabilityConstraint) AddTeamConstraint(teamID int, unavailableDates []time.Time) {
	mtac.teamConstraints[teamID] = NewTeamAvailabilityConstraint(teamID, unavailableDates)
}

// RemoveTeamConstraint removes a team availability constraint
func (mtac *MultiTeamAvailabilityConstraint) RemoveTeamConstraint(teamID int) {
	delete(mtac.teamConstraints, teamID)
}

// GetAllConflictingMatches returns all matches that conflict with any team's availability
func (mtac *MultiTeamAvailabilityConstraint) GetAllConflictingMatches(draw *models.Draw) map[int][]*models.Match {
	conflicts := make(map[int][]*models.Match)
	
	for teamID, constraint := range mtac.teamConstraints {
		conflicting := constraint.GetConflictingMatches(draw)
		if len(conflicting) > 0 {
			conflicts[teamID] = conflicting
		}
	}
	
	return conflicts
}