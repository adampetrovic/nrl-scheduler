package models

import (
	"errors"
	"time"
)

// Team represents an NRL team
type Team struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	ShortName string    `json:"short_name"`
	City      string    `json:"city"`
	VenueID   *int      `json:"venue_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Venue *Venue `json:"venue,omitempty"`
}

// Validate ensures the team has valid data
func (t *Team) Validate() error {
	if t.Name == "" {
		return errors.New("team name cannot be empty")
	}
	if t.ShortName == "" {
		return errors.New("team short name cannot be empty")
	}
	if len(t.ShortName) > 3 {
		return errors.New("team short name cannot be longer than 3 characters")
	}
	if t.City == "" {
		return errors.New("team city cannot be empty")
	}
	if t.Latitude < -90 || t.Latitude > 90 {
		return errors.New("team latitude must be between -90 and 90")
	}
	if t.Longitude < -180 || t.Longitude > 180 {
		return errors.New("team longitude must be between -180 and 180")
	}
	return nil
}

// HasBye returns true if this team ID represents a bye
func (t *Team) HasBye() bool {
	return t == nil || t.ID == 0
}