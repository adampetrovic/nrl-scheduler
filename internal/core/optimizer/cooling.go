package optimizer

import "math"

// CoolingSchedule defines the interface for temperature cooling strategies
type CoolingSchedule interface {
	NextTemperature(initialTemp float64, iteration int) float64
}

// ExponentialCooling implements exponential temperature cooling
// Temperature = T0 * (cooling_rate)^iteration
type ExponentialCooling struct {
	CoolingRate float64
}

// NewExponentialCooling creates a new exponential cooling schedule
func NewExponentialCooling(coolingRate float64) *ExponentialCooling {
	return &ExponentialCooling{
		CoolingRate: coolingRate,
	}
}

// NextTemperature calculates the next temperature using exponential cooling
func (ec *ExponentialCooling) NextTemperature(initialTemp float64, iteration int) float64 {
	return initialTemp * math.Pow(ec.CoolingRate, float64(iteration))
}

// LinearCooling implements linear temperature cooling
// Temperature = T0 - (cooling_rate * iteration)
type LinearCooling struct {
	CoolingRate float64
}

// NewLinearCooling creates a new linear cooling schedule
func NewLinearCooling(coolingRate float64) *LinearCooling {
	return &LinearCooling{
		CoolingRate: coolingRate,
	}
}

// NextTemperature calculates the next temperature using linear cooling
func (lc *LinearCooling) NextTemperature(initialTemp float64, iteration int) float64 {
	temp := initialTemp - (lc.CoolingRate * float64(iteration))
	if temp < 0 {
		return 0
	}
	return temp
}

// AdaptiveCooling adjusts cooling rate based on acceptance rate
type AdaptiveCooling struct {
	BaseCoolingRate    float64
	AcceptanceTarget   float64
	AdaptationFactor   float64
	currentCoolingRate float64
}

// NewAdaptiveCooling creates a new adaptive cooling schedule
func NewAdaptiveCooling(baseCoolingRate, acceptanceTarget, adaptationFactor float64) *AdaptiveCooling {
	return &AdaptiveCooling{
		BaseCoolingRate:    baseCoolingRate,
		AcceptanceTarget:   acceptanceTarget,
		AdaptationFactor:   adaptationFactor,
		currentCoolingRate: baseCoolingRate,
	}
}

// NextTemperature calculates the next temperature using adaptive cooling
func (ac *AdaptiveCooling) NextTemperature(initialTemp float64, iteration int) float64 {
	return initialTemp * math.Pow(ac.currentCoolingRate, float64(iteration))
}

// UpdateAcceptanceRate updates the cooling rate based on current acceptance rate
func (ac *AdaptiveCooling) UpdateAcceptanceRate(acceptanceRate float64) {
	if acceptanceRate > ac.AcceptanceTarget {
		// Acceptance rate too high - cool faster
		ac.currentCoolingRate *= (1.0 - ac.AdaptationFactor)
	} else if acceptanceRate < ac.AcceptanceTarget {
		// Acceptance rate too low - cool slower
		ac.currentCoolingRate *= (1.0 + ac.AdaptationFactor)
	}
	
	// Keep cooling rate within reasonable bounds
	if ac.currentCoolingRate < 0.8 {
		ac.currentCoolingRate = 0.8
	}
	if ac.currentCoolingRate > 0.999 {
		ac.currentCoolingRate = 0.999
	}
}

// LogarithmicCooling implements logarithmic temperature cooling
// Temperature = T0 / log(1 + iteration)
type LogarithmicCooling struct {
	ScalingFactor float64
}

// NewLogarithmicCooling creates a new logarithmic cooling schedule
func NewLogarithmicCooling(scalingFactor float64) *LogarithmicCooling {
	return &LogarithmicCooling{
		ScalingFactor: scalingFactor,
	}
}

// NextTemperature calculates the next temperature using logarithmic cooling
func (lgc *LogarithmicCooling) NextTemperature(initialTemp float64, iteration int) float64 {
	if iteration == 0 {
		return initialTemp
	}
	return initialTemp / (lgc.ScalingFactor * math.Log(1.0+float64(iteration)))
}

// GeometricCooling implements geometric temperature cooling with periodic reheating
// This can help escape local optima
type GeometricCooling struct {
	CoolingRate   float64
	ReheatFactor  float64
	ReheatPeriod  int
}

// NewGeometricCooling creates a new geometric cooling schedule with reheating
func NewGeometricCooling(coolingRate, reheatFactor float64, reheatPeriod int) *GeometricCooling {
	return &GeometricCooling{
		CoolingRate:  coolingRate,
		ReheatFactor: reheatFactor,
		ReheatPeriod: reheatPeriod,
	}
}

// NextTemperature calculates the next temperature using geometric cooling with reheating
func (gc *GeometricCooling) NextTemperature(initialTemp float64, iteration int) float64 {
	// Basic geometric cooling
	temp := initialTemp * math.Pow(gc.CoolingRate, float64(iteration))
	
	// Apply reheating if we're at a reheat period
	if gc.ReheatPeriod > 0 && iteration%gc.ReheatPeriod == 0 && iteration > 0 {
		temp *= gc.ReheatFactor
	}
	
	return temp
}

// CombinedCooling allows combining multiple cooling strategies
type CombinedCooling struct {
	Schedules []CoolingSchedule
	Weights   []float64
}

// NewCombinedCooling creates a new combined cooling schedule
func NewCombinedCooling(schedules []CoolingSchedule, weights []float64) *CombinedCooling {
	if len(schedules) != len(weights) {
		panic("schedules and weights must have the same length")
	}
	
	return &CombinedCooling{
		Schedules: schedules,
		Weights:   weights,
	}
}

// NextTemperature calculates the next temperature using weighted combination of schedules
func (cc *CombinedCooling) NextTemperature(initialTemp float64, iteration int) float64 {
	var totalTemp float64
	var totalWeight float64
	
	for i, schedule := range cc.Schedules {
		temp := schedule.NextTemperature(initialTemp, iteration)
		weight := cc.Weights[i]
		totalTemp += temp * weight
		totalWeight += weight
	}
	
	if totalWeight == 0 {
		return initialTemp
	}
	
	return totalTemp / totalWeight
}

// TemperatureScheduleConfig represents configuration for temperature schedules
type TemperatureScheduleConfig struct {
	Type           string                 `json:"type"`
	CoolingRate    float64               `json:"cooling_rate,omitempty"`
	ScalingFactor  float64               `json:"scaling_factor,omitempty"`
	ReheatFactor   float64               `json:"reheat_factor,omitempty"`
	ReheatPeriod   int                   `json:"reheat_period,omitempty"`
	AcceptanceTarget float64             `json:"acceptance_target,omitempty"`
	AdaptationFactor float64             `json:"adaptation_factor,omitempty"`
	Params         map[string]interface{} `json:"params,omitempty"`
}

// CreateCoolingSchedule creates a cooling schedule from configuration
func CreateCoolingSchedule(config TemperatureScheduleConfig) CoolingSchedule {
	switch config.Type {
	case "exponential":
		return NewExponentialCooling(config.CoolingRate)
	case "linear":
		return NewLinearCooling(config.CoolingRate)
	case "adaptive":
		return NewAdaptiveCooling(config.CoolingRate, config.AcceptanceTarget, config.AdaptationFactor)
	case "logarithmic":
		return NewLogarithmicCooling(config.ScalingFactor)
	case "geometric":
		return NewGeometricCooling(config.CoolingRate, config.ReheatFactor, config.ReheatPeriod)
	default:
		// Default to exponential cooling
		return NewExponentialCooling(0.99)
	}
}