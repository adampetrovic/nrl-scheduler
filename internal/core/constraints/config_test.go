package constraints

import (
	"encoding/json"
	"testing"
)

// TestConstraintFactory tests constraint creation from configuration
func TestConstraintFactory(t *testing.T) {
	factory := NewConstraintFactory()
	
	// Test creating venue availability constraint
	venueParams := map[string]interface{}{
		"venue_id": float64(1),
		"unavailable_dates": []interface{}{"2025-06-15", "2025-07-04"},
	}
	venueConfig := HardConstraintConfig{
		Type:   "venue_availability",
		Params: venueParams,
	}
	
	constraint, err := factory.createHardConstraint(venueConfig)
	if err != nil {
		t.Fatalf("Failed to create venue availability constraint: %v", err)
	}
	
	if constraint.Name() != "VenueAvailability" {
		t.Error("Wrong constraint name")
	}
	if !constraint.IsHard() {
		t.Error("Venue availability should be hard constraint")
	}
	
	// Test creating travel minimization constraint
	travelParams := map[string]interface{}{
		"max_consecutive_away": float64(3),
	}
	travelConfig := SoftConstraintConfig{
		Type:   "travel_minimization",
		Weight: 0.8,
		Params: travelParams,
	}
	
	softConstraint, err := factory.createSoftConstraint(travelConfig)
	if err != nil {
		t.Fatalf("Failed to create travel minimization constraint: %v", err)
	}
	
	if softConstraint.Name() != "TravelMinimization" {
		t.Error("Wrong constraint name")
	}
	if softConstraint.IsHard() {
		t.Error("Travel minimization should be soft constraint")
	}
}

// TestConstraintFactoryErrors tests error handling in constraint creation
func TestConstraintFactoryErrors(t *testing.T) {
	factory := NewConstraintFactory()
	
	// Test unknown constraint type
	unknownConfig := HardConstraintConfig{
		Type:   "unknown_constraint",
		Params: map[string]interface{}{},
	}
	
	_, err := factory.createHardConstraint(unknownConfig)
	if err == nil {
		t.Error("Should return error for unknown constraint type")
	}
	
	// Test missing required parameter
	venueConfigMissingParam := HardConstraintConfig{
		Type: "venue_availability",
		Params: map[string]interface{}{
			"unavailable_dates": []interface{}{"2025-06-15"},
			// Missing venue_id
		},
	}
	
	_, err = factory.createHardConstraint(venueConfigMissingParam)
	if err == nil {
		t.Error("Should return error for missing venue_id parameter")
	}
	
	// Test invalid date format
	venueConfigBadDate := HardConstraintConfig{
		Type: "venue_availability",
		Params: map[string]interface{}{
			"venue_id": float64(1),
			"unavailable_dates": []interface{}{"invalid-date"},
		},
	}
	
	_, err = factory.createHardConstraint(venueConfigBadDate)
	if err == nil {
		t.Error("Should return error for invalid date format")
	}
}

// TestConstraintEngineFromConfig tests creating constraint engine from configuration
func TestConstraintEngineFromConfig(t *testing.T) {
	factory := NewConstraintFactory()
	
	config := ConstraintConfig{
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
				Type:   "home_away_balance",
				Weight: 0.7,
				Params: map[string]interface{}{
					"max_deviation": float64(0.1),
				},
			},
		},
	}
	
	engine, err := factory.CreateConstraintEngine(config)
	if err != nil {
		t.Fatalf("Failed to create constraint engine: %v", err)
	}
	
	// Verify correct number of constraints
	if len(engine.GetHardConstraints()) != 2 {
		t.Errorf("Expected 2 hard constraints, got %d", len(engine.GetHardConstraints()))
	}
	if len(engine.GetSoftConstraints()) != 2 {
		t.Errorf("Expected 2 soft constraints, got %d", len(engine.GetSoftConstraints()))
	}
	
	// Verify constraint types
	hardConstraints := engine.GetHardConstraints()
	hardNames := make(map[string]bool)
	for _, constraint := range hardConstraints {
		hardNames[constraint.Name()] = true
	}
	
	if !hardNames["ByeConstraint"] {
		t.Error("Missing ByeConstraint")
	}
	if !hardNames["DoubleUpConstraint"] {
		t.Error("Missing DoubleUpConstraint")
	}
	
	// Verify soft constraint weights
	softConstraints := engine.GetSoftConstraints()
	for _, weighted := range softConstraints {
		if weighted.Constraint.Name() == "TravelMinimization" && weighted.Weight != 0.8 {
			t.Error("Wrong weight for travel minimization constraint")
		}
		if weighted.Constraint.Name() == "HomeAwayBalance" && weighted.Weight != 0.7 {
			t.Error("Wrong weight for home/away balance constraint")
		}
	}
}

// TestJSONSerialization tests JSON loading and saving
func TestJSONSerialization(t *testing.T) {
	// Create test configuration
	originalConfig := ConstraintConfig{
		Hard: []HardConstraintConfig{
			{
				Type: "team_availability",
				Params: map[string]interface{}{
					"team_id": float64(1),
					"unavailable_dates": []interface{}{"2025-06-15", "2025-07-04"},
				},
			},
		},
		Soft: []SoftConstraintConfig{
			{
				Type:   "rest_period",
				Weight: 0.9,
				Params: map[string]interface{}{
					"min_rest_days": float64(5),
				},
			},
		},
	}
	
	// Save to JSON
	jsonData, err := SaveConstraintConfigToJSON(originalConfig)
	if err != nil {
		t.Fatalf("Failed to save config to JSON: %v", err)
	}
	
	// Load from JSON
	loadedConfig, err := LoadConstraintConfigFromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to load config from JSON: %v", err)
	}
	
	// Compare configurations
	if len(loadedConfig.Hard) != len(originalConfig.Hard) {
		t.Error("Hard constraints count mismatch after JSON round-trip")
	}
	if len(loadedConfig.Soft) != len(originalConfig.Soft) {
		t.Error("Soft constraints count mismatch after JSON round-trip")
	}
	
	// Check specific values
	if loadedConfig.Hard[0].Type != "team_availability" {
		t.Error("Hard constraint type mismatch")
	}
	if loadedConfig.Soft[0].Weight != 0.9 {
		t.Error("Soft constraint weight mismatch")
	}
}

// TestDefaultNRLConfig tests the default NRL configuration
func TestDefaultNRLConfig(t *testing.T) {
	config := GetDefaultNRLConstraintConfig()
	
	// Verify it has some constraints
	if len(config.Hard) == 0 {
		t.Error("Default NRL config should have hard constraints")
	}
	if len(config.Soft) == 0 {
		t.Error("Default NRL config should have soft constraints")
	}
	
	// Verify it's a valid configuration
	factory := NewConstraintFactory()
	_, err := factory.CreateConstraintEngine(config)
	if err != nil {
		t.Fatalf("Default NRL config should be valid: %v", err)
	}
}

// TestConstraintConfigValidation tests configuration validation
func TestConstraintConfigValidation(t *testing.T) {
	// Test valid configuration
	validConfig := ConstraintConfig{
		Hard: []HardConstraintConfig{
			{
				Type:   "bye_constraint",
				Params: map[string]interface{}{},
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
		},
	}
	
	err := ValidateConstraintConfig(validConfig)
	if err != nil {
		t.Errorf("Valid config should pass validation: %v", err)
	}
	
	// Test invalid weight
	invalidWeightConfig := ConstraintConfig{
		Soft: []SoftConstraintConfig{
			{
				Type:   "travel_minimization",
				Weight: 1.5, // Invalid weight > 1
				Params: map[string]interface{}{
					"max_consecutive_away": float64(3),
				},
			},
		},
	}
	
	err = ValidateConstraintConfig(invalidWeightConfig)
	if err == nil {
		t.Error("Should reject config with weight > 1")
	}
	
	// Test empty constraint type
	emptyTypeConfig := ConstraintConfig{
		Hard: []HardConstraintConfig{
			{
				Type:   "", // Empty type
				Params: map[string]interface{}{},
			},
		},
	}
	
	err = ValidateConstraintConfig(emptyTypeConfig)
	if err == nil {
		t.Error("Should reject config with empty constraint type")
	}
}

// TestConstraintTypeInfo tests constraint type information
func TestConstraintTypeInfo(t *testing.T) {
	info := GetConstraintTypeInfo()
	
	// Check that all known constraint types are present
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
		typeInfo, exists := info[expectedType]
		if !exists {
			t.Errorf("Missing constraint type info for: %s", expectedType)
			continue
		}
		
		if typeInfo.Description == "" {
			t.Errorf("Constraint type %s should have description", expectedType)
		}
		
		if typeInfo.Type != "hard" && typeInfo.Type != "soft" {
			t.Errorf("Constraint type %s should be 'hard' or 'soft'", expectedType)
		}
	}
}

// TestComplexConfiguration tests a complex real-world configuration
func TestComplexConfiguration(t *testing.T) {
	factory := NewConstraintFactory()
	
	// Create a complex configuration similar to what might be used in production
	config := ConstraintConfig{
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
			{
				Type: "venue_availability",
				Params: map[string]interface{}{
					"venue_id": float64(1),
					"unavailable_dates": []interface{}{
						"2025-06-15", "2025-07-04", "2025-12-25",
					},
				},
			},
			{
				Type: "team_availability",
				Params: map[string]interface{}{
					"team_id": float64(5),
					"unavailable_dates": []interface{}{
						"2025-08-10", "2025-09-15",
					},
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
	
	// Validate the configuration
	err := ValidateConstraintConfig(config)
	if err != nil {
		t.Fatalf("Complex config should be valid: %v", err)
	}
	
	// Create constraint engine
	engine, err := factory.CreateConstraintEngine(config)
	if err != nil {
		t.Fatalf("Failed to create engine from complex config: %v", err)
	}
	
	// Verify all constraints were created
	if len(engine.GetHardConstraints()) != 4 {
		t.Errorf("Expected 4 hard constraints, got %d", len(engine.GetHardConstraints()))
	}
	if len(engine.GetSoftConstraints()) != 4 {
		t.Errorf("Expected 4 soft constraints, got %d", len(engine.GetSoftConstraints()))
	}
	
	// Test JSON serialization of complex config
	jsonData, err := SaveConstraintConfigToJSON(config)
	if err != nil {
		t.Fatalf("Failed to serialize complex config: %v", err)
	}
	
	// Verify JSON is valid
	var jsonCheck interface{}
	err = json.Unmarshal(jsonData, &jsonCheck)
	if err != nil {
		t.Fatalf("Generated JSON should be valid: %v", err)
	}
	
	// Test round-trip
	loadedConfig, err := LoadConstraintConfigFromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to load serialized complex config: %v", err)
	}
	
	// Create engine from loaded config
	_, err = factory.CreateConstraintEngine(loadedConfig)
	if err != nil {
		t.Fatalf("Failed to create engine from loaded config: %v", err)
	}
}