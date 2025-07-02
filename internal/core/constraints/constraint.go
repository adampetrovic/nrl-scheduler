package constraints

import (
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// Constraint defines the interface for all constraints
type Constraint interface {
	// Validate checks if a match violates this constraint
	Validate(match *models.Match, draw *models.Draw) error

	// Score returns a score for how well the draw satisfies this constraint
	// Higher scores are better. Returns 0.0 to 1.0 for soft constraints
	Score(draw *models.Draw) float64

	// IsHard returns true if this is a hard constraint (must be satisfied)
	IsHard() bool

	// Name returns the constraint name
	Name() string

	// Description returns a human-readable description
	Description() string
}

// WeightedConstraint wraps a soft constraint with a weight
type WeightedConstraint struct {
	Constraint Constraint
	Weight     float64
}

// ConstraintEngine manages and evaluates all constraints
type ConstraintEngine struct {
	hardConstraints []Constraint
	softConstraints []WeightedConstraint
}

// NewConstraintEngine creates a new constraint engine
func NewConstraintEngine() *ConstraintEngine {
	return &ConstraintEngine{
		hardConstraints: make([]Constraint, 0),
		softConstraints: make([]WeightedConstraint, 0),
	}
}

// AddHardConstraint adds a hard constraint to the engine
func (ce *ConstraintEngine) AddHardConstraint(constraint Constraint) {
	if constraint.IsHard() {
		ce.hardConstraints = append(ce.hardConstraints, constraint)
	}
}

// AddSoftConstraint adds a soft constraint with weight to the engine
func (ce *ConstraintEngine) AddSoftConstraint(constraint Constraint, weight float64) {
	if !constraint.IsHard() {
		ce.softConstraints = append(ce.softConstraints, WeightedConstraint{
			Constraint: constraint,
			Weight:     weight,
		})
	}
}

// ValidateMatch checks if a match violates any hard constraints
func (ce *ConstraintEngine) ValidateMatch(match *models.Match, draw *models.Draw) error {
	for _, constraint := range ce.hardConstraints {
		if err := constraint.Validate(match, draw); err != nil {
			return err
		}
	}
	return nil
}

// ValidateDraw checks if the entire draw violates any hard constraints
func (ce *ConstraintEngine) ValidateDraw(draw *models.Draw) []error {
	var errors []error

	for _, match := range draw.Matches {
		if err := ce.ValidateMatch(match, draw); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// ScoreDraw calculates the total score for a draw considering all constraints
func (ce *ConstraintEngine) ScoreDraw(draw *models.Draw) float64 {
	// First check hard constraints - if any fail, return 0
	if violations := ce.ValidateDraw(draw); len(violations) > 0 {
		return 0.0
	}

	// Calculate weighted score from soft constraints
	var totalScore float64
	var totalWeight float64

	for _, weighted := range ce.softConstraints {
		score := weighted.Constraint.Score(draw)
		totalScore += score * weighted.Weight
		totalWeight += weighted.Weight
	}

	// Return normalized score
	if totalWeight == 0 {
		return 1.0 // No soft constraints means perfect score
	}

	return totalScore / totalWeight
}

// GetHardConstraints returns all hard constraints
func (ce *ConstraintEngine) GetHardConstraints() []Constraint {
	return ce.hardConstraints
}

// GetSoftConstraints returns all soft constraints with weights
func (ce *ConstraintEngine) GetSoftConstraints() []WeightedConstraint {
	return ce.softConstraints
}

// ConstraintViolation represents a constraint violation
type ConstraintViolation struct {
	ConstraintName string
	MatchID        int
	Round          int
	Description    string
	Severity       ViolationSeverity
}

// ViolationSeverity indicates how severe a constraint violation is
type ViolationSeverity string

const (
	SeverityHard    ViolationSeverity = "hard"
	SeveritySoft    ViolationSeverity = "soft"
	SeverityWarning ViolationSeverity = "warning"
)

// AnalyzeDraw performs comprehensive constraint analysis
func (ce *ConstraintEngine) AnalyzeDraw(draw *models.Draw) []ConstraintViolation {
	var violations []ConstraintViolation

	// Check hard constraints
	for _, constraint := range ce.hardConstraints {
		for _, match := range draw.Matches {
			if err := constraint.Validate(match, draw); err != nil {
				violations = append(violations, ConstraintViolation{
					ConstraintName: constraint.Name(),
					MatchID:        match.ID,
					Round:          match.Round,
					Description:    err.Error(),
					Severity:       SeverityHard,
				})
			}
		}

		// Check overall draw score for this constraint
		if score := constraint.Score(draw); score < 0.5 {
			violations = append(violations, ConstraintViolation{
				ConstraintName: constraint.Name(),
				MatchID:        0,
				Round:          0,
				Description:    "Overall constraint satisfaction below threshold",
				Severity:       SeverityWarning,
			})
		}
	}

	// Check soft constraints
	for _, weighted := range ce.softConstraints {
		if score := weighted.Constraint.Score(draw); score < 0.3 {
			violations = append(violations, ConstraintViolation{
				ConstraintName: weighted.Constraint.Name(),
				MatchID:        0,
				Round:          0,
				Description:    "Soft constraint poorly satisfied",
				Severity:       SeveritySoft,
			})
		}
	}

	return violations
}

// BaseConstraint provides common functionality for constraints
type BaseConstraint struct {
	name        string
	description string
	isHard      bool
}

// NewBaseConstraint creates a base constraint
func NewBaseConstraint(name, description string, isHard bool) BaseConstraint {
	return BaseConstraint{
		name:        name,
		description: description,
		isHard:      isHard,
	}
}

// Name returns the constraint name
func (bc BaseConstraint) Name() string {
	return bc.name
}

// Description returns the constraint description
func (bc BaseConstraint) Description() string {
	return bc.description
}

// IsHard returns whether this is a hard constraint
func (bc BaseConstraint) IsHard() bool {
	return bc.isHard
}

// DateConstraint provides helper methods for date-related constraints
type DateConstraint struct {
	BaseConstraint
	unavailableDates []time.Time
}

// NewDateConstraint creates a date-based constraint
func NewDateConstraint(name, description string, isHard bool, unavailableDates []time.Time) DateConstraint {
	return DateConstraint{
		BaseConstraint:   NewBaseConstraint(name, description, isHard),
		unavailableDates: unavailableDates,
	}
}

// IsDateUnavailable checks if a given date is in the unavailable list
func (dc DateConstraint) IsDateUnavailable(date time.Time) bool {
	for _, unavailable := range dc.unavailableDates {
		if unavailable.Year() == date.Year() &&
			unavailable.YearDay() == date.YearDay() {
			return true
		}
	}
	return false
}

// GetUnavailableDates returns the list of unavailable dates
func (dc DateConstraint) GetUnavailableDates() []time.Time {
	return dc.unavailableDates
}
