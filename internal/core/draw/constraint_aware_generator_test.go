package draw

import (
	"encoding/json"
	"testing"

	"github.com/adampetrovic/nrl-scheduler/internal/core/constraints"
	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// TestConstraintAwareGenerator tests the constraint-aware generator
func TestConstraintAwareGenerator(t *testing.T) {
	teams := createConstraintTestTeams()
	config := constraints.GetDefaultNRLConstraintConfig()
	
	generator, err := NewConstraintAwareGenerator(teams, 10, config)
	if err != nil {
		t.Fatalf("Failed to create constraint-aware generator: %v", err)
	}
	
	// Test generation with constraints
	draw, violations, err := generator.GenerateWithConstraints()
	if err != nil {
		t.Fatalf("Failed to generate draw with constraints: %v", err)
	}
	
	if draw == nil {
		t.Fatal("Draw should not be nil")
	}
	
	if len(draw.Matches) == 0 {
		t.Error("Draw should have matches")
	}
	
	// Verify constraint configuration is stored
	if len(draw.ConstraintConfig) == 0 {
		t.Error("Draw should store constraint configuration")
	}
	
	// Test validation - with default constraints, there might be some violations
	// but the draw should still be generated
	t.Logf("Generated draw with %d violations", len(violations))
	
	// Test scoring
	score := generator.ScoreDraw(draw)
	if score < 0 || score > 1 {
		t.Errorf("Score should be between 0 and 1, got %f", score)
	}
}

// TestConstraintAwareGeneratorFromJSON tests creation from JSON config
func TestConstraintAwareGeneratorFromJSON(t *testing.T) {
	teams := createConstraintTestTeams()
	
	// Create JSON configuration
	config := constraints.ConstraintConfig{
		Hard: []constraints.HardConstraintConfig{
			{
				Type:   "bye_constraint",
				Params: map[string]interface{}{},
			},
		},
		Soft: []constraints.SoftConstraintConfig{
			{
				Type:   "travel_minimization",
				Weight: 0.8,
				Params: map[string]interface{}{
					"max_consecutive_away": float64(3),
				},
			},
		},
	}
	
	configJSON, err := constraints.SaveConstraintConfigToJSON(config)
	if err != nil {
		t.Fatalf("Failed to save config to JSON: %v", err)
	}
	
	generator, err := NewConstraintAwareGeneratorFromJSON(teams, 6, configJSON)
	if err != nil {
		t.Fatalf("Failed to create generator from JSON: %v", err)
	}
	
	// Test generation
	draw, violations, err := generator.GenerateWithConstraints()
	if err != nil {
		t.Fatalf("Failed to generate draw: %v", err)
	}
	
	if draw == nil {
		t.Fatal("Draw should not be nil")
	}
	
	t.Logf("Generated draw from JSON config with %d violations", len(violations))
}

// TestGenerateWithAnalysis tests comprehensive generation with analysis
func TestGenerateWithAnalysis(t *testing.T) {
	teams := createConstraintTestTeams()
	config := constraints.GetDefaultNRLConstraintConfig()
	
	generator, err := NewConstraintAwareGenerator(teams, 8, config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	result, err := generator.GenerateWithAnalysis()
	if err != nil {
		t.Fatalf("Failed to generate with analysis: %v", err)
	}
	
	if result.Draw == nil {
		t.Fatal("Result should contain draw")
	}
	
	if result.Score < 0 || result.Score > 1 {
		t.Errorf("Score should be between 0 and 1, got %f", result.Score)
	}
	
	// Analysis should contain some results
	if result.Analysis == nil {
		t.Error("Result should contain analysis")
	}
	
	t.Logf("Generated draw with score %f, %d hard violations, %d soft violations",
		result.Score, result.HardViolations, result.SoftViolations)
}

// TestDoubleRoundRobinWithConstraints tests double round-robin generation
func TestDoubleRoundRobinWithConstraints(t *testing.T) {
	teams := createConstraintTestTeams()
	config := constraints.GetDefaultNRLConstraintConfig()
	
	generator, err := NewConstraintAwareGenerator(teams, 10, config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	result, err := generator.GenerateDoubleWithAnalysis()
	if err != nil {
		t.Fatalf("Failed to generate double round-robin: %v", err)
	}
	
	if result.Draw == nil {
		t.Fatal("Result should contain draw")
	}
	
	// Should have matches for the full rounds
	if len(result.Draw.Matches) == 0 {
		t.Error("Double round-robin should have matches")
	}
	
	t.Logf("Double round-robin: %d matches, score %f", 
		len(result.Draw.Matches), result.Score)
}

// TestUpdateConstraints tests updating constraints on existing generator
func TestUpdateConstraints(t *testing.T) {
	teams := createConstraintTestTeams()
	initialConfig := constraints.ConstraintConfig{
		Hard: []constraints.HardConstraintConfig{
			{
				Type:   "bye_constraint",
				Params: map[string]interface{}{},
			},
		},
	}
	
	generator, err := NewConstraintAwareGenerator(teams, 6, initialConfig)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	// Generate initial draw
	initialResult, err := generator.GenerateWithAnalysis()
	if err != nil {
		t.Fatalf("Failed to generate initial draw: %v", err)
	}
	
	// Update constraints to add soft constraints
	newConfig := constraints.ConstraintConfig{
		Hard: []constraints.HardConstraintConfig{
			{
				Type:   "bye_constraint",
				Params: map[string]interface{}{},
			},
		},
		Soft: []constraints.SoftConstraintConfig{
			{
				Type:   "travel_minimization",
				Weight: 0.8,
				Params: map[string]interface{}{
					"max_consecutive_away": float64(2),
				},
			},
		},
	}
	
	err = generator.UpdateConstraints(newConfig)
	if err != nil {
		t.Fatalf("Failed to update constraints: %v", err)
	}
	
	// Generate new draw with updated constraints
	newResult, err := generator.GenerateWithAnalysis()
	if err != nil {
		t.Fatalf("Failed to generate draw with updated constraints: %v", err)
	}
	
	// Verify both draws were generated
	if initialResult.Draw == nil || newResult.Draw == nil {
		t.Fatal("Both draws should be generated")
	}
	
	t.Logf("Initial score: %f, Updated score: %f", 
		initialResult.Score, newResult.Score)
}

// TestConstraintValidation tests constraint validation
func TestConstraintValidation(t *testing.T) {
	teams := createConstraintTestTeams()
	config := constraints.GetDefaultNRLConstraintConfig()
	
	generator, err := NewConstraintAwareGenerator(teams, 6, config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	// Create a test draw
	draw, _, err := generator.GenerateWithConstraints()
	if err != nil {
		t.Fatalf("Failed to generate test draw: %v", err)
	}
	
	// Test validating the draw
	violations := generator.ValidateDraw(draw)
	t.Logf("Draw has %d constraint violations", len(violations))
	
	// Test analyzing the draw
	analysis := generator.AnalyzeDraw(draw)
	t.Logf("Draw analysis contains %d items", len(analysis))
	
	// Verify analysis structure
	for _, violation := range analysis {
		if violation.ConstraintName == "" {
			t.Error("Analysis item should have constraint name")
		}
		if violation.Severity == "" {
			t.Error("Analysis item should have severity")
		}
	}
}

// TestExportConstraintConfig tests exporting constraint configuration
func TestExportConstraintConfig(t *testing.T) {
	teams := createConstraintTestTeams()
	config := constraints.GetDefaultNRLConstraintConfig()
	
	generator, err := NewConstraintAwareGenerator(teams, 6, config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	// Export configuration
	exportedJSON, err := generator.ExportConstraintConfig()
	if err != nil {
		t.Fatalf("Failed to export constraint config: %v", err)
	}
	
	// Verify it's valid JSON
	var exportedConfig constraints.ConstraintConfig
	err = json.Unmarshal(exportedJSON, &exportedConfig)
	if err != nil {
		t.Fatalf("Exported JSON should be valid: %v", err)
	}
	
	// Verify it can be used to create a new generator
	newGenerator, err := NewConstraintAwareGeneratorFromJSON(teams, 6, exportedJSON)
	if err != nil {
		t.Fatalf("Should be able to create generator from exported config: %v", err)
	}
	
	// Test generation with exported config
	_, _, err = newGenerator.GenerateWithConstraints()
	if err != nil {
		t.Fatalf("Should be able to generate with exported config: %v", err)
	}
}

// TestDefaultNRLGenerator tests the default NRL generator
func TestDefaultNRLGenerator(t *testing.T) {
	teams := createConstraintTestTeams()
	
	generator, err := GetDefaultNRLGenerator(teams, 8)
	if err != nil {
		t.Fatalf("Failed to create default NRL generator: %v", err)
	}
	
	result, err := generator.GenerateWithAnalysis()
	if err != nil {
		t.Fatalf("Failed to generate with default NRL constraints: %v", err)
	}
	
	if result.Draw == nil {
		t.Fatal("Default generator should produce draw")
	}
	
	// Should have some constraints applied
	engine := generator.GetConstraintEngine()
	if len(engine.GetHardConstraints()) == 0 && len(engine.GetSoftConstraints()) == 0 {
		t.Error("Default NRL generator should have constraints")
	}
	
	t.Logf("Default NRL generator produced draw with score %f", result.Score)
}

// TestConstraintTypeInfo tests getting constraint type information
func TestConstraintTypeInfo(t *testing.T) {
	teams := createConstraintTestTeams()
	config := constraints.GetDefaultNRLConstraintConfig()
	
	generator, err := NewConstraintAwareGenerator(teams, 6, config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	typeInfo := generator.GetConstraintTypeInfo()
	
	// Should contain information about all constraint types
	expectedTypes := []string{
		"venue_availability",
		"bye_constraint",
		"team_availability", 
		"double_up",
		"travel_minimization",
		"rest_period",
		"prime_time_spread",
		"home_away_balance",
	}
	
	for _, expectedType := range expectedTypes {
		info, exists := typeInfo[expectedType]
		if !exists {
			t.Errorf("Missing type info for: %s", expectedType)
			continue
		}
		
		if info.Description == "" {
			t.Errorf("Type %s should have description", expectedType)
		}
	}
}

// TestValidateConstraintConfig tests constraint configuration validation
func TestValidateConstraintConfig(t *testing.T) {
	teams := createConstraintTestTeams()
	config := constraints.GetDefaultNRLConstraintConfig()
	
	generator, err := NewConstraintAwareGenerator(teams, 6, config)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}
	
	// Test valid configuration
	validConfig := constraints.ConstraintConfig{
		Hard: []constraints.HardConstraintConfig{
			{
				Type:   "bye_constraint",
				Params: map[string]interface{}{},
			},
		},
		Soft: []constraints.SoftConstraintConfig{
			{
				Type:   "travel_minimization",
				Weight: 0.8,
				Params: map[string]interface{}{
					"max_consecutive_away": float64(3),
				},
			},
		},
	}
	
	err = generator.ValidateConstraintConfig(validConfig)
	if err != nil {
		t.Errorf("Valid config should pass validation: %v", err)
	}
	
	// Test invalid configuration
	invalidConfig := constraints.ConstraintConfig{
		Soft: []constraints.SoftConstraintConfig{
			{
				Type:   "travel_minimization",
				Weight: 1.5, // Invalid weight > 1
				Params: map[string]interface{}{
					"max_consecutive_away": float64(3),
				},
			},
		},
	}
	
	err = generator.ValidateConstraintConfig(invalidConfig)
	if err == nil {
		t.Error("Invalid config should fail validation")
	}
}

// createConstraintTestTeams creates a set of teams for testing
func createConstraintTestTeams() []*models.Team {
	return []*models.Team{
		{ID: 1, Name: "Brisbane Broncos", VenueID: &[]int{1}[0]},
		{ID: 2, Name: "Melbourne Storm", VenueID: &[]int{2}[0]},
		{ID: 3, Name: "Sydney Roosters", VenueID: &[]int{3}[0]},
		{ID: 4, Name: "Penrith Panthers", VenueID: &[]int{4}[0]},
		{ID: 5, Name: "Parramatta Eels", VenueID: &[]int{5}[0]},
		{ID: 6, Name: "Canterbury Bulldogs", VenueID: &[]int{6}[0]},
	}
}

// Benchmark tests for performance
func BenchmarkConstraintAwareGeneration(b *testing.B) {
	teams := createConstraintTestTeams()
	config := constraints.GetDefaultNRLConstraintConfig()
	
	generator, err := NewConstraintAwareGenerator(teams, 10, config)
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := generator.GenerateWithConstraints()
		if err != nil {
			b.Fatalf("Generation failed: %v", err)
		}
	}
}

func BenchmarkConstraintValidation(b *testing.B) {
	teams := createConstraintTestTeams()
	config := constraints.GetDefaultNRLConstraintConfig()
	
	generator, err := NewConstraintAwareGenerator(teams, 10, config)
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}
	
	// Generate a test draw
	draw, _, err := generator.GenerateWithConstraints()
	if err != nil {
		b.Fatalf("Failed to generate test draw: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.ValidateDraw(draw)
	}
}

func BenchmarkDrawScoring(b *testing.B) {
	teams := createConstraintTestTeams()
	config := constraints.GetDefaultNRLConstraintConfig()
	
	generator, err := NewConstraintAwareGenerator(teams, 10, config)
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}
	
	// Generate a test draw
	draw, _, err := generator.GenerateWithConstraints()
	if err != nil {
		b.Fatalf("Failed to generate test draw: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.ScoreDraw(draw)
	}
}