package models

import (
	"errors"
	"time"
)

// Match represents a single match in a draw
type Match struct {
	ID          int        `json:"id"`
	DrawID      int        `json:"draw_id"`
	Round       int        `json:"round"`
	HomeTeamID  *int       `json:"home_team_id"`
	AwayTeamID  *int       `json:"away_team_id"`
	VenueID     *int       `json:"venue_id"`
	MatchDate   *time.Time `json:"match_date"`
	MatchTime   *time.Time `json:"match_time"`
	IsPrimeTime bool       `json:"is_prime_time"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relations
	HomeTeam *Team  `json:"home_team,omitempty"`
	AwayTeam *Team  `json:"away_team,omitempty"`
	Venue    *Venue `json:"venue,omitempty"`
}

// Validate ensures the match has valid data
func (m *Match) Validate() error {
	if m.DrawID <= 0 {
		return errors.New("match must belong to a draw")
	}
	if m.Round <= 0 {
		return errors.New("match round must be positive")
	}

	// Check if it's a bye (both teams nil) or a regular match
	if m.HomeTeamID == nil && m.AwayTeamID == nil {
		// This is a bye - valid
		return nil
	}

	// For regular matches, both teams must be set
	if m.HomeTeamID == nil || m.AwayTeamID == nil {
		return errors.New("match must have both home and away teams or be a bye")
	}

	// Teams cannot play against themselves
	if *m.HomeTeamID == *m.AwayTeamID {
		return errors.New("team cannot play against itself")
	}

	// Regular matches should have a venue
	if m.VenueID == nil {
		return errors.New("match must have a venue")
	}

	return nil
}

// IsBye returns true if this match represents a bye
func (m *Match) IsBye() bool {
	return m.HomeTeamID == nil && m.AwayTeamID == nil
}

// HasTeam returns true if the match involves the specified team
func (m *Match) HasTeam(teamID int) bool {
	if m.IsBye() {
		return false
	}
	return (m.HomeTeamID != nil && *m.HomeTeamID == teamID) ||
		(m.AwayTeamID != nil && *m.AwayTeamID == teamID)
}

// IsScheduled returns true if the match has a date assigned
func (m *Match) IsScheduled() bool {
	return m.MatchDate != nil
}

// GetOpponent returns the opponent team ID for the given team
func (m *Match) GetOpponent(teamID int) (*int, error) {
	if m.IsBye() {
		return nil, errors.New("bye matches have no opponent")
	}

	if m.HomeTeamID != nil && *m.HomeTeamID == teamID {
		return m.AwayTeamID, nil
	}
	if m.AwayTeamID != nil && *m.AwayTeamID == teamID {
		return m.HomeTeamID, nil
	}

	return nil, errors.New("team not in this match")
}

// IsHomeGame returns true if the specified team is playing at home
func (m *Match) IsHomeGame(teamID int) (bool, error) {
	if !m.HasTeam(teamID) {
		return false, errors.New("team not in this match")
	}
	return m.HomeTeamID != nil && *m.HomeTeamID == teamID, nil
}