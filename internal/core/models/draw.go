package models

import (
	"encoding/json"
	"errors"
	"time"
)

// DrawStatus represents the status of a draw
type DrawStatus string

const (
	DrawStatusDraft      DrawStatus = "draft"
	DrawStatusOptimizing DrawStatus = "optimizing"
	DrawStatusCompleted  DrawStatus = "completed"
)

// Draw represents a season draw/schedule
type Draw struct {
	ID               int             `json:"id"`
	Name             string          `json:"name"`
	SeasonYear       int             `json:"season_year"`
	Rounds           int             `json:"rounds"`
	Status           DrawStatus      `json:"status"`
	ConstraintConfig json.RawMessage `json:"constraint_config,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`

	// Relations
	Matches []*Match `json:"matches,omitempty"`
}

// Validate ensures the draw has valid data
func (d *Draw) Validate() error {
	if d.Name == "" {
		return errors.New("draw name cannot be empty")
	}
	if d.SeasonYear < 2000 || d.SeasonYear > 2100 {
		return errors.New("season year must be between 2000 and 2100")
	}
	if d.Rounds < 1 || d.Rounds > 52 {
		return errors.New("rounds must be between 1 and 52")
	}
	if !d.isValidStatus() {
		return errors.New("invalid draw status")
	}
	return nil
}

func (d *Draw) isValidStatus() bool {
	switch d.Status {
	case DrawStatusDraft, DrawStatusOptimizing, DrawStatusCompleted:
		return true
	default:
		return false
	}
}

// GetMatchesByRound returns all matches for a specific round
func (d *Draw) GetMatchesByRound(round int) []*Match {
	var matches []*Match
	for _, m := range d.Matches {
		if m.Round == round {
			matches = append(matches, m)
		}
	}
	return matches
}

// GetMatchesByTeam returns all matches for a specific team
func (d *Draw) GetMatchesByTeam(teamID int) []*Match {
	var matches []*Match
	for _, m := range d.Matches {
		if (m.HomeTeamID != nil && *m.HomeTeamID == teamID) ||
			(m.AwayTeamID != nil && *m.AwayTeamID == teamID) {
			matches = append(matches, m)
		}
	}
	return matches
}

// IsComplete returns true if all matches have been scheduled
func (d *Draw) IsComplete() bool {
	if len(d.Matches) == 0 {
		return false
	}

	for _, m := range d.Matches {
		if m.MatchDate == nil {
			return false
		}
	}
	return true
}