package constraints

import (
	"fmt"
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// VenueAvailabilityConstraint ensures venues are not used on unavailable dates
type VenueAvailabilityConstraint struct {
	DateConstraint
	venueID int
}

// NewVenueAvailabilityConstraint creates a new venue availability constraint
func NewVenueAvailabilityConstraint(venueID int, unavailableDates []time.Time) *VenueAvailabilityConstraint {
	return &VenueAvailabilityConstraint{
		DateConstraint: NewDateConstraint(
			"VenueAvailability",
			fmt.Sprintf("Venue %d must not be used on specified unavailable dates", venueID),
			true, // This is a hard constraint
			unavailableDates,
		),
		venueID: venueID,
	}
}

// Validate checks if a match violates the venue availability constraint
func (vac *VenueAvailabilityConstraint) Validate(match *models.Match, draw *models.Draw) error {
	// Skip validation for bye matches
	if match.IsBye() {
		return nil
	}

	// Skip if this match doesn't use our venue
	if match.VenueID == nil || *match.VenueID != vac.venueID {
		return nil
	}

	// Skip if match doesn't have a date assigned yet
	if match.MatchDate == nil {
		return nil
	}

	// Check if the match date conflicts with unavailable dates
	if vac.IsDateUnavailable(*match.MatchDate) {
		return fmt.Errorf("venue %d is not available on %s",
			vac.venueID, match.MatchDate.Format("2006-01-02"))
	}

	return nil
}

// Score calculates how well the draw satisfies this constraint
func (vac *VenueAvailabilityConstraint) Score(draw *models.Draw) float64 {
	totalMatches := 0
	violatingMatches := 0

	for _, match := range draw.Matches {
		// Only consider matches at this venue
		if match.VenueID != nil && *match.VenueID == vac.venueID && match.MatchDate != nil {
			totalMatches++
			if vac.IsDateUnavailable(*match.MatchDate) {
				violatingMatches++
			}
		}
	}

	// If no matches at this venue, constraint is perfectly satisfied
	if totalMatches == 0 {
		return 1.0
	}

	// Return the percentage of non-violating matches
	return float64(totalMatches-violatingMatches) / float64(totalMatches)
}

// GetVenueID returns the venue ID this constraint applies to
func (vac *VenueAvailabilityConstraint) GetVenueID() int {
	return vac.venueID
}

// GetUnavailableDatesForVenue returns unavailable dates for this venue
func (vac *VenueAvailabilityConstraint) GetUnavailableDatesForVenue() []time.Time {
	return vac.GetUnavailableDates()
}
