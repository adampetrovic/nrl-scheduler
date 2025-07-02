package draw

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/constraints"
	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// ConstraintAwareGenerator creates draws while respecting constraint rules
type ConstraintAwareGenerator struct {
	*Generator
	constraintEngine *constraints.ConstraintEngine
	factory          *constraints.ConstraintFactory
}

// NewConstraintAwareGenerator creates a new constraint-aware draw generator
func NewConstraintAwareGenerator(teams []*models.Team, rounds int, constraintConfig constraints.ConstraintConfig) (*ConstraintAwareGenerator, error) {
	baseGenerator, err := NewGenerator(teams, rounds)
	if err != nil {
		return nil, fmt.Errorf("failed to create base generator: %w", err)
	}
	
	factory := constraints.NewConstraintFactory()
	engine, err := factory.CreateConstraintEngine(constraintConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create constraint engine: %w", err)
	}
	
	return &ConstraintAwareGenerator{
		Generator:        baseGenerator,
		constraintEngine: engine,
		factory:          factory,
	}, nil
}

// NewConstraintAwareGeneratorFromJSON creates generator from JSON constraint config
func NewConstraintAwareGeneratorFromJSON(teams []*models.Team, rounds int, configJSON []byte) (*ConstraintAwareGenerator, error) {
	config, err := constraints.LoadConstraintConfigFromJSON(configJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to load constraint config: %w", err)
	}
	
	return NewConstraintAwareGenerator(teams, rounds, config)
}

// GenerateWithConstraints creates a draw and validates it against constraints
func (cag *ConstraintAwareGenerator) GenerateWithConstraints() (*models.Draw, []error, error) {
	// Generate the base draw
	draw, err := cag.GenerateRoundRobin()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate base draw: %w", err)
	}
	
	// Store constraint configuration in the draw
	if configJSON, err := constraints.SaveConstraintConfigToJSON(cag.getConstraintConfig()); err == nil {
		draw.ConstraintConfig = json.RawMessage(configJSON)
	}
	
	// Validate against constraints
	violations := cag.constraintEngine.ValidateDraw(draw)
	
	return draw, violations, nil
}

// GenerateDoubleWithConstraints creates a double round-robin draw with constraint validation
func (cag *ConstraintAwareGenerator) GenerateDoubleWithConstraints() (*models.Draw, []error, error) {
	// Generate the base double round-robin draw
	draw, err := cag.GenerateDoubleRoundRobin()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate base double draw: %w", err)
	}
	
	// Store constraint configuration in the draw
	if configJSON, err := constraints.SaveConstraintConfigToJSON(cag.getConstraintConfig()); err == nil {
		draw.ConstraintConfig = json.RawMessage(configJSON)
	}
	
	// Validate against constraints
	violations := cag.constraintEngine.ValidateDraw(draw)
	
	return draw, violations, nil
}

// ValidateDraw validates an existing draw against the configured constraints
func (cag *ConstraintAwareGenerator) ValidateDraw(draw *models.Draw) []error {
	return cag.constraintEngine.ValidateDraw(draw)
}

// ScoreDraw calculates the constraint satisfaction score for a draw
func (cag *ConstraintAwareGenerator) ScoreDraw(draw *models.Draw) float64 {
	return cag.constraintEngine.ScoreDraw(draw)
}

// AnalyzeDraw performs comprehensive constraint analysis
func (cag *ConstraintAwareGenerator) AnalyzeDraw(draw *models.Draw) []constraints.ConstraintViolation {
	return cag.constraintEngine.AnalyzeDraw(draw)
}

// getConstraintConfig reconstructs the constraint configuration
func (cag *ConstraintAwareGenerator) getConstraintConfig() constraints.ConstraintConfig {
	config := constraints.ConstraintConfig{
		Hard: []constraints.HardConstraintConfig{},
		Soft: []constraints.SoftConstraintConfig{},
	}
	
	// Add hard constraints
	for _, constraint := range cag.constraintEngine.GetHardConstraints() {
		hardConfig := constraints.HardConstraintConfig{
			Type:   cag.getConstraintType(constraint),
			Params: cag.getConstraintParams(constraint),
		}
		config.Hard = append(config.Hard, hardConfig)
	}
	
	// Add soft constraints
	for _, weighted := range cag.constraintEngine.GetSoftConstraints() {
		softConfig := constraints.SoftConstraintConfig{
			Type:   cag.getConstraintType(weighted.Constraint),
			Weight: weighted.Weight,
			Params: cag.getConstraintParams(weighted.Constraint),
		}
		config.Soft = append(config.Soft, softConfig)
	}
	
	return config
}

// getConstraintType maps constraint names to configuration types
func (cag *ConstraintAwareGenerator) getConstraintType(constraint constraints.Constraint) string {
	switch constraint.(type) {
	case *constraints.ByeConstraint:
		return "bye_constraint"
	case *constraints.DoubleUpConstraint:
		return "double_up"
	case *constraints.VenueAvailabilityConstraint:
		return "venue_availability"
	case *constraints.TeamAvailabilityConstraint:
		return "team_availability"
	case *constraints.TravelMinimizationConstraint:
		return "travel_minimization"
	case *constraints.RestPeriodConstraint:
		return "rest_period"
	case *constraints.PrimeTimeSpreadConstraint:
		return "prime_time_spread"
	case *constraints.HomeAwayBalanceConstraint:
		return "home_away_balance"
	default:
		return constraint.Name()
	}
}

// getConstraintParams extracts parameters from a constraint (basic implementation)
func (cag *ConstraintAwareGenerator) getConstraintParams(constraint constraints.Constraint) map[string]interface{} {
	params := make(map[string]interface{})
	
	// This is a simplified implementation - in a full system you'd want
	// constraints to export their parameters properly
	switch c := constraint.(type) {
	case *constraints.DoubleUpConstraint:
		params["min_rounds_separation"] = c.GetMinRoundsSeparation()
	case *constraints.TravelMinimizationConstraint:
		params["max_consecutive_away"] = c.GetMaxConsecutiveAway()
	case *constraints.RestPeriodConstraint:
		params["min_rest_days"] = c.GetMinRestDays()
	case *constraints.PrimeTimeSpreadConstraint:
		params["target_ratio"] = c.GetTargetPrimeTimeRatio()
		params["max_deviation"] = c.GetMaxDeviation()
	case *constraints.HomeAwayBalanceConstraint:
		params["max_deviation"] = c.GetMaxDeviation()
	case *constraints.VenueAvailabilityConstraint:
		params["venue_id"] = c.GetVenueID()
		params["unavailable_dates"] = cag.formatDates(c.GetUnavailableDatesForVenue())
	case *constraints.TeamAvailabilityConstraint:
		params["team_id"] = c.GetTeamID()
		params["unavailable_dates"] = cag.formatDates(c.GetUnavailableDatesForTeam())
	}
	
	return params
}

// formatDates converts time.Time slice to string slice for JSON serialization
func (cag *ConstraintAwareGenerator) formatDates(dates []time.Time) []string {
	formatted := make([]string, len(dates))
	for i, date := range dates {
		formatted[i] = date.Format("2006-01-02")
	}
	return formatted
}

// GenerationResult contains the result of constraint-aware generation
type GenerationResult struct {
	Draw           *models.Draw                     `json:"draw"`
	Score          float64                          `json:"score"`
	Violations     []error                          `json:"violations"`
	Analysis       []constraints.ConstraintViolation `json:"analysis"`
	HardViolations int                              `json:"hard_violations"`
	SoftViolations int                              `json:"soft_violations"`
}

// GenerateWithAnalysis creates a draw and provides comprehensive analysis
func (cag *ConstraintAwareGenerator) GenerateWithAnalysis() (*GenerationResult, error) {
	// Generate the draw with constraint validation
	draw, violations, err := cag.GenerateWithConstraints()
	if err != nil {
		return nil, err
	}
	
	// Calculate score
	score := cag.ScoreDraw(draw)
	
	// Perform detailed analysis
	analysis := cag.AnalyzeDraw(draw)
	
	// Count violation types
	hardViolations := 0
	softViolations := 0
	for _, violation := range analysis {
		switch violation.Severity {
		case constraints.SeverityHard:
			hardViolations++
		case constraints.SeveritySoft:
			softViolations++
		}
	}
	
	return &GenerationResult{
		Draw:           draw,
		Score:          score,
		Violations:     violations,
		Analysis:       analysis,
		HardViolations: hardViolations,
		SoftViolations: softViolations,
	}, nil
}

// GenerateDoubleWithAnalysis creates a double round-robin draw with analysis
func (cag *ConstraintAwareGenerator) GenerateDoubleWithAnalysis() (*GenerationResult, error) {
	// Generate the draw with constraint validation
	draw, violations, err := cag.GenerateDoubleWithConstraints()
	if err != nil {
		return nil, err
	}
	
	// Calculate score
	score := cag.ScoreDraw(draw)
	
	// Perform detailed analysis
	analysis := cag.AnalyzeDraw(draw)
	
	// Count violation types
	hardViolations := 0
	softViolations := 0
	for _, violation := range analysis {
		switch violation.Severity {
		case constraints.SeverityHard:
			hardViolations++
		case constraints.SeveritySoft:
			softViolations++
		}
	}
	
	return &GenerationResult{
		Draw:           draw,
		Score:          score,
		Violations:     violations,
		Analysis:       analysis,
		HardViolations: hardViolations,
		SoftViolations: softViolations,
	}, nil
}

// GetConstraintEngine returns the constraint engine for advanced operations
func (cag *ConstraintAwareGenerator) GetConstraintEngine() *constraints.ConstraintEngine {
	return cag.constraintEngine
}

// UpdateConstraints updates the constraint configuration
func (cag *ConstraintAwareGenerator) UpdateConstraints(config constraints.ConstraintConfig) error {
	engine, err := cag.factory.CreateConstraintEngine(config)
	if err != nil {
		return fmt.Errorf("failed to create new constraint engine: %w", err)
	}
	
	cag.constraintEngine = engine
	return nil
}

// GetConstraintTypeInfo returns information about available constraint types
func (cag *ConstraintAwareGenerator) GetConstraintTypeInfo() map[string]constraints.ConstraintTypeInfo {
	return constraints.GetConstraintTypeInfo()
}

// ValidateConstraintConfig validates a constraint configuration without applying it
func (cag *ConstraintAwareGenerator) ValidateConstraintConfig(config constraints.ConstraintConfig) error {
	return constraints.ValidateConstraintConfig(config)
}

// ExportConstraintConfig exports the current constraint configuration as JSON
func (cag *ConstraintAwareGenerator) ExportConstraintConfig() ([]byte, error) {
	config := cag.getConstraintConfig()
	return constraints.SaveConstraintConfigToJSON(config)
}

// GetDefaultNRLGenerator creates a generator with default NRL constraints
func GetDefaultNRLGenerator(teams []*models.Team, rounds int) (*ConstraintAwareGenerator, error) {
	config := constraints.GetDefaultNRLConstraintConfig()
	return NewConstraintAwareGenerator(teams, rounds, config)
}