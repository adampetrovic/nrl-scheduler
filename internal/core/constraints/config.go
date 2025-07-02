package constraints

import (
	"encoding/json"
	"fmt"
	"time"
)

// ConstraintConfig represents the JSON configuration for all constraints
type ConstraintConfig struct {
	Hard []HardConstraintConfig `json:"hard"`
	Soft []SoftConstraintConfig `json:"soft"`
}

// HardConstraintConfig represents configuration for hard constraints
type HardConstraintConfig struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}

// SoftConstraintConfig represents configuration for soft constraints
type SoftConstraintConfig struct {
	Type   string                 `json:"type"`
	Weight float64               `json:"weight"`
	Params map[string]interface{} `json:"params"`
}

// ConstraintFactory creates constraints from configuration
type ConstraintFactory struct{}

// NewConstraintFactory creates a new constraint factory
func NewConstraintFactory() *ConstraintFactory {
	return &ConstraintFactory{}
}

// CreateConstraintEngine creates a constraint engine from JSON configuration
func (cf *ConstraintFactory) CreateConstraintEngine(config ConstraintConfig) (*ConstraintEngine, error) {
	engine := NewConstraintEngine()
	
	// Create hard constraints
	for _, hardConfig := range config.Hard {
		constraint, err := cf.createHardConstraint(hardConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create hard constraint %s: %w", hardConfig.Type, err)
		}
		engine.AddHardConstraint(constraint)
	}
	
	// Create soft constraints
	for _, softConfig := range config.Soft {
		constraint, err := cf.createSoftConstraint(softConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create soft constraint %s: %w", softConfig.Type, err)
		}
		engine.AddSoftConstraint(constraint, softConfig.Weight)
	}
	
	return engine, nil
}

// createHardConstraint creates a hard constraint from configuration
func (cf *ConstraintFactory) createHardConstraint(config HardConstraintConfig) (Constraint, error) {
	switch config.Type {
	case "venue_availability":
		return cf.createVenueAvailabilityConstraint(config.Params)
		
	case "bye_constraint":
		return cf.createByeConstraint(config.Params)
		
	case "team_availability":
		return cf.createTeamAvailabilityConstraint(config.Params)
		
	case "double_up":
		return cf.createDoubleUpConstraint(config.Params)
		
	default:
		return nil, fmt.Errorf("unknown hard constraint type: %s", config.Type)
	}
}

// createSoftConstraint creates a soft constraint from configuration
func (cf *ConstraintFactory) createSoftConstraint(config SoftConstraintConfig) (Constraint, error) {
	switch config.Type {
	case "travel_minimization":
		return cf.createTravelMinimizationConstraint(config.Params)
		
	case "rest_period":
		return cf.createRestPeriodConstraint(config.Params)
		
	case "prime_time_spread":
		return cf.createPrimeTimeSpreadConstraint(config.Params)
		
	case "home_away_balance":
		return cf.createHomeAwayBalanceConstraint(config.Params)
		
	default:
		return nil, fmt.Errorf("unknown soft constraint type: %s", config.Type)
	}
}

// createVenueAvailabilityConstraint creates a venue availability constraint
func (cf *ConstraintFactory) createVenueAvailabilityConstraint(params map[string]interface{}) (Constraint, error) {
	venueID, ok := params["venue_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("venue_id parameter required and must be a number")
	}
	
	datesInterface, ok := params["unavailable_dates"]
	if !ok {
		return nil, fmt.Errorf("unavailable_dates parameter required")
	}
	
	dateStrings, ok := datesInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unavailable_dates must be an array")
	}
	
	var dates []time.Time
	for _, dateInterface := range dateStrings {
		dateStr, ok := dateInterface.(string)
		if !ok {
			return nil, fmt.Errorf("each date must be a string")
		}
		
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format %s (use YYYY-MM-DD): %w", dateStr, err)
		}
		dates = append(dates, date)
	}
	
	return NewVenueAvailabilityConstraint(int(venueID), dates), nil
}

// createByeConstraint creates a bye constraint
func (cf *ConstraintFactory) createByeConstraint(params map[string]interface{}) (Constraint, error) {
	// Bye constraint doesn't need parameters
	return NewByeConstraint(), nil
}

// createTeamAvailabilityConstraint creates a team availability constraint
func (cf *ConstraintFactory) createTeamAvailabilityConstraint(params map[string]interface{}) (Constraint, error) {
	teamID, ok := params["team_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("team_id parameter required and must be a number")
	}
	
	datesInterface, ok := params["unavailable_dates"]
	if !ok {
		return nil, fmt.Errorf("unavailable_dates parameter required")
	}
	
	dateStrings, ok := datesInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unavailable_dates must be an array")
	}
	
	var dates []time.Time
	for _, dateInterface := range dateStrings {
		dateStr, ok := dateInterface.(string)
		if !ok {
			return nil, fmt.Errorf("each date must be a string")
		}
		
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date format %s (use YYYY-MM-DD): %w", dateStr, err)
		}
		dates = append(dates, date)
	}
	
	return NewTeamAvailabilityConstraint(int(teamID), dates), nil
}

// createDoubleUpConstraint creates a double-up constraint
func (cf *ConstraintFactory) createDoubleUpConstraint(params map[string]interface{}) (Constraint, error) {
	minRounds, ok := params["min_rounds_separation"].(float64)
	if !ok {
		return nil, fmt.Errorf("min_rounds_separation parameter required and must be a number")
	}
	
	return NewDoubleUpConstraint(int(minRounds)), nil
}

// createTravelMinimizationConstraint creates a travel minimization constraint
func (cf *ConstraintFactory) createTravelMinimizationConstraint(params map[string]interface{}) (Constraint, error) {
	maxConsecutive, ok := params["max_consecutive_away"].(float64)
	if !ok {
		return nil, fmt.Errorf("max_consecutive_away parameter required and must be a number")
	}
	
	return NewTravelMinimizationConstraint(int(maxConsecutive)), nil
}

// createRestPeriodConstraint creates a rest period constraint
func (cf *ConstraintFactory) createRestPeriodConstraint(params map[string]interface{}) (Constraint, error) {
	minDays, ok := params["min_rest_days"].(float64)
	if !ok {
		return nil, fmt.Errorf("min_rest_days parameter required and must be a number")
	}
	
	return NewRestPeriodConstraint(int(minDays)), nil
}

// createPrimeTimeSpreadConstraint creates a prime time spread constraint
func (cf *ConstraintFactory) createPrimeTimeSpreadConstraint(params map[string]interface{}) (Constraint, error) {
	targetRatio, ok := params["target_ratio"].(float64)
	if !ok {
		return nil, fmt.Errorf("target_ratio parameter required and must be a number")
	}
	
	maxDeviation, ok := params["max_deviation"].(float64)
	if !ok {
		return nil, fmt.Errorf("max_deviation parameter required and must be a number")
	}
	
	return NewPrimeTimeSpreadConstraint(targetRatio, maxDeviation), nil
}

// createHomeAwayBalanceConstraint creates a home/away balance constraint
func (cf *ConstraintFactory) createHomeAwayBalanceConstraint(params map[string]interface{}) (Constraint, error) {
	maxDeviation, ok := params["max_deviation"].(float64)
	if !ok {
		return nil, fmt.Errorf("max_deviation parameter required and must be a number")
	}
	
	return NewHomeAwayBalanceConstraint(maxDeviation), nil
}

// LoadConstraintConfigFromJSON loads constraint configuration from JSON bytes
func LoadConstraintConfigFromJSON(data []byte) (ConstraintConfig, error) {
	var config ConstraintConfig
	err := json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse JSON: %w", err)
	}
	
	return config, nil
}

// SaveConstraintConfigToJSON saves constraint configuration to JSON bytes
func SaveConstraintConfigToJSON(config ConstraintConfig) ([]byte, error) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return data, nil
}

// GetDefaultNRLConstraintConfig returns a default constraint configuration for NRL
func GetDefaultNRLConstraintConfig() ConstraintConfig {
	return ConstraintConfig{
		Hard: []HardConstraintConfig{
			{
				Type:   "bye_constraint",
				Params: map[string]interface{}{},
			},
			{
				Type: "double_up",
				Params: map[string]interface{}{
					"min_rounds_separation": float64(8),
				},
			},
		},
		Soft: []SoftConstraintConfig{
			{
				Type:   "travel_minimization",
				Weight: 0.8,
				Params: map[string]interface{}{
					"max_consecutive_away": float64(3),
				},
			},
			{
				Type:   "rest_period",
				Weight: 0.9,
				Params: map[string]interface{}{
					"min_rest_days": float64(5),
				},
			},
			{
				Type:   "prime_time_spread",
				Weight: 0.7,
				Params: map[string]interface{}{
					"target_ratio":   float64(0.3),
					"max_deviation": float64(0.1),
				},
			},
			{
				Type:   "home_away_balance",
				Weight: 0.8,
				Params: map[string]interface{}{
					"max_deviation": float64(0.1),
				},
			},
		},
	}
}

// ValidateConstraintConfig validates a constraint configuration
func ValidateConstraintConfig(config ConstraintConfig) error {
	factory := NewConstraintFactory()
	
	// Validate hard constraints
	for i, hardConfig := range config.Hard {
		if hardConfig.Type == "" {
			return fmt.Errorf("hard constraint %d: type cannot be empty", i)
		}
		
		_, err := factory.createHardConstraint(hardConfig)
		if err != nil {
			return fmt.Errorf("hard constraint %d (%s): %w", i, hardConfig.Type, err)
		}
	}
	
	// Validate soft constraints
	for i, softConfig := range config.Soft {
		if softConfig.Type == "" {
			return fmt.Errorf("soft constraint %d: type cannot be empty", i)
		}
		
		if softConfig.Weight < 0 || softConfig.Weight > 1 {
			return fmt.Errorf("soft constraint %d (%s): weight must be between 0 and 1", i, softConfig.Type)
		}
		
		_, err := factory.createSoftConstraint(softConfig)
		if err != nil {
			return fmt.Errorf("soft constraint %d (%s): %w", i, softConfig.Type, err)
		}
	}
	
	return nil
}

// GetConstraintTypeInfo returns information about available constraint types
func GetConstraintTypeInfo() map[string]ConstraintTypeInfo {
	return map[string]ConstraintTypeInfo{
		"venue_availability": {
			Type:        "hard",
			Description: "Ensures venues are not used on unavailable dates",
			Parameters: map[string]string{
				"venue_id":          "int - ID of the venue",
				"unavailable_dates": "[]string - Array of dates in YYYY-MM-DD format",
			},
		},
		"bye_constraint": {
			Type:        "hard",
			Description: "Ensures each team gets exactly one bye per full round-robin",
			Parameters:  map[string]string{},
		},
		"team_availability": {
			Type:        "hard",
			Description: "Ensures teams are not scheduled on unavailable dates",
			Parameters: map[string]string{
				"team_id":           "int - ID of the team",
				"unavailable_dates": "[]string - Array of dates in YYYY-MM-DD format",
			},
		},
		"double_up": {
			Type:        "hard",
			Description: "Teams cannot play each other twice within X rounds",
			Parameters: map[string]string{
				"min_rounds_separation": "int - Minimum rounds between same matchups",
			},
		},
		"travel_minimization": {
			Type:        "soft",
			Description: "Minimize consecutive away games to reduce travel burden",
			Parameters: map[string]string{
				"max_consecutive_away": "int - Maximum consecutive away games allowed",
			},
		},
		"rest_period": {
			Type:        "soft",
			Description: "Ensure minimum rest days between matches for player welfare",
			Parameters: map[string]string{
				"min_rest_days": "int - Minimum rest days between matches",
			},
		},
		"prime_time_spread": {
			Type:        "soft",
			Description: "Distribute prime-time games fairly across all teams",
			Parameters: map[string]string{
				"target_ratio":   "float - Target ratio of prime time games (0.0-1.0)",
				"max_deviation": "float - Maximum allowed deviation from target",
			},
		},
		"home_away_balance": {
			Type:        "soft",
			Description: "Balance home and away games fairly for all teams",
			Parameters: map[string]string{
				"max_deviation": "float - Maximum deviation from 50/50 balance",
			},
		},
	}
}

// ConstraintTypeInfo contains information about a constraint type
type ConstraintTypeInfo struct {
	Type        string            `json:"type"`        // "hard" or "soft"
	Description string            `json:"description"`
	Parameters  map[string]string `json:"parameters"`
}