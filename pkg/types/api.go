package types

import (
	"encoding/json"
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/constraints"
	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
	"github.com/adampetrovic/nrl-scheduler/internal/core/optimizer"
)

// Team API types
type CreateTeamRequest struct {
	Name      string  `json:"name" validate:"required,min=1,max=100"`
	ShortName string  `json:"short_name" validate:"required,min=1,max=3"`
	City      string  `json:"city" validate:"required,min=1,max=100"`
	VenueID   *int    `json:"venue_id,omitempty"`
	Latitude  float64 `json:"latitude" validate:"min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"min=-180,max=180"`
}

type UpdateTeamRequest struct {
	Name      *string  `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	ShortName *string  `json:"short_name,omitempty" validate:"omitempty,min=1,max=3"`
	City      *string  `json:"city,omitempty" validate:"omitempty,min=1,max=100"`
	VenueID   *int     `json:"venue_id,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	Longitude *float64 `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
}

type TeamResponse struct {
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	ShortName string         `json:"short_name"`
	City      string         `json:"city"`
	VenueID   *int           `json:"venue_id"`
	Venue     *VenueResponse `json:"venue,omitempty"`
	Latitude  float64        `json:"latitude"`
	Longitude float64        `json:"longitude"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Venue API types
type CreateVenueRequest struct {
	Name      string  `json:"name" validate:"required,min=1,max=100"`
	City      string  `json:"city" validate:"required,min=1,max=100"`
	Capacity  int     `json:"capacity" validate:"required,min=1,max=200000"`
	Latitude  float64 `json:"latitude" validate:"min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"min=-180,max=180"`
}

type UpdateVenueRequest struct {
	Name      *string  `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	City      *string  `json:"city,omitempty" validate:"omitempty,min=1,max=100"`
	Capacity  *int     `json:"capacity,omitempty" validate:"omitempty,min=1,max=200000"`
	Latitude  *float64 `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	Longitude *float64 `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
}

type VenueResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	City      string    `json:"city"`
	Capacity  int       `json:"capacity"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Draw API types
type CreateDrawRequest struct {
	Name             string                       `json:"name" validate:"required,min=1,max=100"`
	SeasonYear       int                          `json:"season_year" validate:"required,min=2000,max=2100"`
	Rounds           int                          `json:"rounds" validate:"required,min=1,max=52"`
	ConstraintConfig *constraints.ConstraintConfig `json:"constraint_config,omitempty"`
}

type UpdateDrawRequest struct {
	Name             *string                       `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	SeasonYear       *int                          `json:"season_year,omitempty" validate:"omitempty,min=2000,max=2100"`
	Rounds           *int                          `json:"rounds,omitempty" validate:"omitempty,min=1,max=52"`
	ConstraintConfig *constraints.ConstraintConfig `json:"constraint_config,omitempty"`
}

type DrawResponse struct {
	ID               int               `json:"id"`
	Name             string            `json:"name"`
	SeasonYear       int               `json:"season_year"`
	Rounds           int               `json:"rounds"`
	Status           string            `json:"status"`
	ConstraintConfig interface{}       `json:"constraint_config,omitempty"`
	MatchCount       int               `json:"match_count"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// Match API types
type MatchResponse struct {
	ID          int             `json:"id"`
	DrawID      int             `json:"draw_id"`
	Round       int             `json:"round"`
	HomeTeam    *TeamResponse   `json:"home_team,omitempty"`
	AwayTeam    *TeamResponse   `json:"away_team,omitempty"`
	Venue       *VenueResponse  `json:"venue,omitempty"`
	ScheduledAt *time.Time      `json:"scheduled_at,omitempty"`
	IsBye       bool            `json:"is_bye"`
	Created     time.Time       `json:"created"`
	Updated     time.Time       `json:"updated"`
}

// Draw generation types
type GenerateDrawRequest struct {
	Constraints *constraints.ConstraintConfig `json:"constraints,omitempty"`
	Options     *GenerationOptions            `json:"options,omitempty"`
}

type GenerationOptions struct {
	Seed           *int64 `json:"seed,omitempty"`
	MaxAttempts    *int   `json:"max_attempts,omitempty"`
	ValidateAfter  *bool  `json:"validate_after,omitempty"`
}

type GenerateDrawResponse struct {
	Success        bool                       `json:"success"`
	MatchCount     int                        `json:"match_count"`
	Violations     []ConstraintViolation      `json:"violations,omitempty"`
	Message        string                     `json:"message"`
	GeneratedAt    time.Time                  `json:"generated_at"`
	GenerationTime time.Duration              `json:"generation_time"`
}

// Constraint validation types
type ValidateConstraintsRequest struct {
	Constraints *constraints.ConstraintConfig `json:"constraints,omitempty"`
}

type ValidateConstraintsResponse struct {
	IsValid    bool                  `json:"is_valid"`
	Violations []ConstraintViolation `json:"violations"`
	Score      float64               `json:"score"`
}

type ConstraintViolation struct {
	Type        string            `json:"type"`
	Severity    string            `json:"severity"` // "hard" or "soft"
	Description string            `json:"description"`
	MatchID     *int              `json:"match_id,omitempty"`
	Round       *int              `json:"round,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// Optimization API types
type TemperatureScheduleRequest struct {
	Type             string                 `json:"type"`
	CoolingRate      float64               `json:"cooling_rate,omitempty"`
	ScalingFactor    float64               `json:"scaling_factor,omitempty"`
	ReheatFactor     float64               `json:"reheat_factor,omitempty"`
	ReheatPeriod     int                   `json:"reheat_period,omitempty"`
	AcceptanceTarget float64               `json:"acceptance_target,omitempty"`
	AdaptationFactor float64               `json:"adaptation_factor,omitempty"`
	Params           map[string]interface{} `json:"params,omitempty"`
}

type StartOptimizationRequest struct {
	Temperature     float64                     `json:"temperature" validate:"required,min=0.1,max=1000"`
	CoolingRate     float64                     `json:"cooling_rate" validate:"required,min=0.1,max=0.999"`
	MaxIterations   int                         `json:"max_iterations" validate:"required,min=100,max=1000000"`
	CoolingSchedule *TemperatureScheduleRequest `json:"cooling_schedule,omitempty"`
}

type StartOptimizationResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

type OptimizationStatusResponse struct {
	JobID       string                      `json:"job_id"`
	DrawID      int                         `json:"draw_id"`
	Status      string                      `json:"status"`
	Progress    optimizer.OptimizationProgress `json:"progress"`
	StartedAt   time.Time                   `json:"started_at"`
	CompletedAt *time.Time                  `json:"completed_at,omitempty"`
	Error       *string                     `json:"error,omitempty"`
}

type OptimizationJobsResponse struct {
	Jobs []*optimizer.OptimizationJob `json:"jobs"`
}

type ConstraintValidationResponse struct {
	DrawID     int                             `json:"draw_id"`
	IsValid    bool                            `json:"is_valid"`
	Violations []constraints.ConstraintViolation `json:"violations"`
}

type DrawScoreResponse struct {
	DrawID int     `json:"draw_id"`
	Score  float64 `json:"score"`
}

// Generic API response types
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

// Query parameters
type ListQueryParams struct {
	Page     int    `form:"page" validate:"omitempty,min=1"`
	PerPage  int    `form:"per_page" validate:"omitempty,min=1,max=100"`
	Search   string `form:"search" validate:"omitempty,max=200"`
	SortBy   string `form:"sort_by" validate:"omitempty,oneof=id name created updated"`
	SortDir  string `form:"sort_dir" validate:"omitempty,oneof=asc desc"`
	IsActive *bool  `form:"is_active"`
}

// Conversion helpers
func TeamToResponse(team *models.Team, venue *models.Venue) TeamResponse {
	resp := TeamResponse{
		ID:        team.ID,
		Name:      team.Name,
		ShortName: team.ShortName,
		City:      team.City,
		VenueID:   team.VenueID,
		Latitude:  team.Latitude,
		Longitude: team.Longitude,
		CreatedAt: team.CreatedAt,
		UpdatedAt: team.UpdatedAt,
	}
	
	if venue != nil {
		resp.Venue = &VenueResponse{
			ID:        venue.ID,
			Name:      venue.Name,
			City:      venue.City,
			Capacity:  venue.Capacity,
			Latitude:  venue.Latitude,
			Longitude: venue.Longitude,
			CreatedAt: venue.CreatedAt,
			UpdatedAt: venue.UpdatedAt,
		}
	}
	
	return resp
}

func VenueToResponse(venue *models.Venue) VenueResponse {
	return VenueResponse{
		ID:        venue.ID,
		Name:      venue.Name,
		City:      venue.City,
		Capacity:  venue.Capacity,
		Latitude:  venue.Latitude,
		Longitude: venue.Longitude,
		CreatedAt: venue.CreatedAt,
		UpdatedAt: venue.UpdatedAt,
	}
}

func DrawToResponse(draw *models.Draw) DrawResponse {
	var constraintConfig interface{}
	if len(draw.ConstraintConfig) > 0 {
		// Try to unmarshal the raw JSON
		var config constraints.ConstraintConfig
		if err := json.Unmarshal(draw.ConstraintConfig, &config); err == nil {
			constraintConfig = config
		} else {
			constraintConfig = string(draw.ConstraintConfig)
		}
	}
	
	matchCount := 0
	if draw.Matches != nil {
		matchCount = len(draw.Matches)
	}
	
	return DrawResponse{
		ID:               draw.ID,
		Name:             draw.Name,
		SeasonYear:       draw.SeasonYear,
		Rounds:           draw.Rounds,
		Status:           string(draw.Status),
		ConstraintConfig: constraintConfig,
		MatchCount:       matchCount,
		CreatedAt:        draw.CreatedAt,
		UpdatedAt:        draw.UpdatedAt,
	}
}

func MatchToResponse(match *models.Match, homeTeam, awayTeam *models.Team, venue *models.Venue) MatchResponse {
	resp := MatchResponse{
		ID:          match.ID,
		DrawID:      match.DrawID,
		Round:       match.Round,
		ScheduledAt: match.MatchDate,
		IsBye:       match.IsBye(),
		Created:     match.CreatedAt,
		Updated:     match.UpdatedAt,
	}
	
	if homeTeam != nil {
		team := TeamToResponse(homeTeam, nil)
		resp.HomeTeam = &team
	}
	
	if awayTeam != nil {
		team := TeamToResponse(awayTeam, nil)
		resp.AwayTeam = &team
	}
	
	if venue != nil {
		v := VenueToResponse(venue)
		resp.Venue = &v
	}
	
	return resp
}