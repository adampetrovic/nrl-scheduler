package optimizer

import (
	"math"
	"testing"
)

func TestExponentialCooling(t *testing.T) {
	cooling := NewExponentialCooling(0.95)
	initialTemp := 100.0

	// Test first iteration (should return initial temperature)
	temp0 := cooling.NextTemperature(initialTemp, 0)
	if temp0 != initialTemp {
		t.Errorf("Expected initial temperature %f, got %f", initialTemp, temp0)
	}

	// Test subsequent iterations
	temp1 := cooling.NextTemperature(initialTemp, 1)
	expected1 := initialTemp * 0.95
	if temp1 != expected1 {
		t.Errorf("Expected temperature %f, got %f", expected1, temp1)
	}

	temp10 := cooling.NextTemperature(initialTemp, 10)
	expected10 := initialTemp * math.Pow(0.95, 10)
	if math.Abs(temp10-expected10) > 1e-10 {
		t.Errorf("Expected temperature %f, got %f", expected10, temp10)
	}

	// Test that temperature decreases
	if temp1 >= temp0 {
		t.Error("Temperature should decrease over iterations")
	}
	if temp10 >= temp1 {
		t.Error("Temperature should continue decreasing")
	}
}

func TestLinearCooling(t *testing.T) {
	cooling := NewLinearCooling(5.0)
	initialTemp := 100.0

	// Test first iteration
	temp0 := cooling.NextTemperature(initialTemp, 0)
	if temp0 != initialTemp {
		t.Errorf("Expected initial temperature %f, got %f", initialTemp, temp0)
	}

	// Test subsequent iterations
	temp1 := cooling.NextTemperature(initialTemp, 1)
	expected1 := initialTemp - 5.0
	if temp1 != expected1 {
		t.Errorf("Expected temperature %f, got %f", expected1, temp1)
	}

	temp10 := cooling.NextTemperature(initialTemp, 10)
	expected10 := initialTemp - 50.0
	if temp10 != expected10 {
		t.Errorf("Expected temperature %f, got %f", expected10, temp10)
	}

	// Test that temperature doesn't go below zero
	temp100 := cooling.NextTemperature(initialTemp, 100)
	if temp100 != 0 {
		t.Errorf("Expected temperature 0, got %f", temp100)
	}
}

func TestAdaptiveCooling(t *testing.T) {
	cooling := NewAdaptiveCooling(0.95, 0.4, 0.1)
	initialTemp := 100.0

	// Test initial behavior
	temp0 := cooling.NextTemperature(initialTemp, 0)
	if temp0 != initialTemp {
		t.Errorf("Expected initial temperature %f, got %f", initialTemp, temp0)
	}

	// Test adaptation to high acceptance rate
	cooling.UpdateAcceptanceRate(0.8) // Higher than target 0.4
	temp1 := cooling.NextTemperature(initialTemp, 1)
	
	// Should cool faster now
	cooling2 := NewAdaptiveCooling(0.95, 0.4, 0.1)
	temp1_normal := cooling2.NextTemperature(initialTemp, 1)
	
	if temp1 >= temp1_normal {
		t.Error("Expected faster cooling with high acceptance rate")
	}

	// Test adaptation to low acceptance rate with fresh instance
	cooling3 := NewAdaptiveCooling(0.95, 0.4, 0.1)
	cooling3.UpdateAcceptanceRate(0.1) // Lower than target 0.4
	_ = cooling3.NextTemperature(initialTemp, 2)
	
	// Should have slower cooling rate now (rate should increase from 0.95, meaning it cools slower)
	if cooling3.currentCoolingRate <= 0.95 {
		t.Errorf("Expected cooling rate to increase from 0.95, got %f", cooling3.currentCoolingRate)
	}
}

func TestLogarithmicCooling(t *testing.T) {
	cooling := NewLogarithmicCooling(1.0)
	initialTemp := 100.0

	// Test first iteration
	temp0 := cooling.NextTemperature(initialTemp, 0)
	if temp0 != initialTemp {
		t.Errorf("Expected initial temperature %f, got %f", initialTemp, temp0)
	}

	// Test subsequent iterations
	temp1 := cooling.NextTemperature(initialTemp, 1)
	expected1 := initialTemp / (1.0 * math.Log(2.0))
	if math.Abs(temp1-expected1) > 1e-10 {
		t.Errorf("Expected temperature %f, got %f", expected1, temp1)
	}

	// Test that temperature decreases
	temp10 := cooling.NextTemperature(initialTemp, 10)
	if temp10 >= temp1 {
		t.Error("Temperature should decrease over iterations")
	}
}

func TestGeometricCooling(t *testing.T) {
	cooling := NewGeometricCooling(0.95, 2.0, 5)
	initialTemp := 100.0

	// Test normal cooling
	temp1 := cooling.NextTemperature(initialTemp, 1)
	expected1 := initialTemp * 0.95
	if temp1 != expected1 {
		t.Errorf("Expected temperature %f, got %f", expected1, temp1)
	}

	// Test reheating at period
	temp5 := cooling.NextTemperature(initialTemp, 5)
	temp5_normal := initialTemp * math.Pow(0.95, 5)
	expected5 := temp5_normal * 2.0 // With reheating
	if temp5 != expected5 {
		t.Errorf("Expected reheated temperature %f, got %f", expected5, temp5)
	}

	// Test that reheating occurred
	temp4 := cooling.NextTemperature(initialTemp, 4)
	if temp5 <= temp4 {
		t.Error("Expected reheating to increase temperature")
	}
}

func TestCombinedCooling(t *testing.T) {
	exp := NewExponentialCooling(0.95)
	linear := NewLinearCooling(5.0)
	
	schedules := []CoolingSchedule{exp, linear}
	weights := []float64{0.7, 0.3}
	
	cooling := NewCombinedCooling(schedules, weights)
	initialTemp := 100.0

	temp1 := cooling.NextTemperature(initialTemp, 1)
	
	expTemp := exp.NextTemperature(initialTemp, 1)
	linearTemp := linear.NextTemperature(initialTemp, 1)
	expected := (expTemp*0.7 + linearTemp*0.3) / (0.7 + 0.3)
	
	if math.Abs(temp1-expected) > 1e-10 {
		t.Errorf("Expected combined temperature %f, got %f", expected, temp1)
	}
}

func TestCombinedCooling_PanicsOnMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for mismatched schedules and weights")
		}
	}()

	exp := NewExponentialCooling(0.95)
	schedules := []CoolingSchedule{exp}
	weights := []float64{0.7, 0.3} // More weights than schedules

	NewCombinedCooling(schedules, weights)
}

func TestCreateCoolingSchedule(t *testing.T) {
	testCases := []struct {
		name     string
		config   TemperatureScheduleConfig
		expected string
	}{
		{
			name: "exponential",
			config: TemperatureScheduleConfig{
				Type:        "exponential",
				CoolingRate: 0.95,
			},
			expected: "exponential",
		},
		{
			name: "linear",
			config: TemperatureScheduleConfig{
				Type:        "linear",
				CoolingRate: 5.0,
			},
			expected: "linear",
		},
		{
			name: "adaptive",
			config: TemperatureScheduleConfig{
				Type:             "adaptive",
				CoolingRate:      0.95,
				AcceptanceTarget: 0.4,
				AdaptationFactor: 0.1,
			},
			expected: "adaptive",
		},
		{
			name: "logarithmic",
			config: TemperatureScheduleConfig{
				Type:          "logarithmic",
				ScalingFactor: 1.0,
			},
			expected: "logarithmic",
		},
		{
			name: "geometric",
			config: TemperatureScheduleConfig{
				Type:         "geometric",
				CoolingRate:  0.95,
				ReheatFactor: 2.0,
				ReheatPeriod: 5,
			},
			expected: "geometric",
		},
		{
			name: "unknown_defaults_to_exponential",
			config: TemperatureScheduleConfig{
				Type: "unknown",
			},
			expected: "exponential",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schedule := CreateCoolingSchedule(tc.config)
			if schedule == nil {
				t.Error("Expected cooling schedule to be created")
			}

			// Test that it works by calling NextTemperature
			temp := schedule.NextTemperature(100.0, 1)
			if temp < 0 {
				t.Error("Expected non-negative temperature")
			}
		})
	}
}

