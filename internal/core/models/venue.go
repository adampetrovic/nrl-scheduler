package models

import (
	"errors"
	"time"
)

// Venue represents a sports venue
type Venue struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	City      string    `json:"city"`
	Capacity  int       `json:"capacity"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate ensures the venue has valid data
func (v *Venue) Validate() error {
	if v.Name == "" {
		return errors.New("venue name cannot be empty")
	}
	if v.City == "" {
		return errors.New("venue city cannot be empty")
	}
	if v.Capacity < 0 {
		return errors.New("venue capacity cannot be negative")
	}
	if v.Latitude < -90 || v.Latitude > 90 {
		return errors.New("venue latitude must be between -90 and 90")
	}
	if v.Longitude < -180 || v.Longitude > 180 {
		return errors.New("venue longitude must be between -180 and 180")
	}
	return nil
}