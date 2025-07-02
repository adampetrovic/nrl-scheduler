package optimizer

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/constraints"
	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// SimulatedAnnealing implements the simulated annealing optimization algorithm
type SimulatedAnnealing struct {
	Temperature      float64
	CoolingRate      float64
	MaxIterations    int
	ConstraintEngine *constraints.ConstraintEngine
	CoolingSchedule  CoolingSchedule
}

// OptimizationResult contains the results of an optimization run
type OptimizationResult struct {
	InitialScore    float64       `json:"initial_score"`
	FinalScore      float64       `json:"final_score"`
	Iterations      int           `json:"iterations"`
	Improvements    int           `json:"improvements"`
	Duration        time.Duration `json:"duration"`
	BestDraw        *models.Draw  `json:"best_draw,omitempty"`
}

// OptimizationProgress tracks the current state of optimization
type OptimizationProgress struct {
	Iteration       int     `json:"iteration"`
	Temperature     float64 `json:"temperature"`
	CurrentScore    float64 `json:"current_score"`
	BestScore       float64 `json:"best_score"`
	AcceptanceRate  float64 `json:"acceptance_rate"`
	EstimatedTime   string  `json:"estimated_time"`
}

// ProgressCallback is called during optimization to report progress
type ProgressCallback func(progress OptimizationProgress)

// NewSimulatedAnnealing creates a new simulated annealing optimizer
func NewSimulatedAnnealing(temperature, coolingRate float64, maxIterations int, constraintEngine *constraints.ConstraintEngine) *SimulatedAnnealing {
	return &SimulatedAnnealing{
		Temperature:      temperature,
		CoolingRate:      coolingRate,
		MaxIterations:    maxIterations,
		ConstraintEngine: constraintEngine,
		CoolingSchedule:  NewExponentialCooling(coolingRate),
	}
}

// Optimize runs the simulated annealing algorithm on the given draw
func (sa *SimulatedAnnealing) Optimize(draw *models.Draw, callback ProgressCallback) (*OptimizationResult, error) {
	if draw == nil {
		return nil, fmt.Errorf("draw cannot be nil")
	}

	if len(draw.Matches) == 0 {
		return nil, fmt.Errorf("draw has no matches to optimize")
	}

	startTime := time.Now()
	
	// Create a copy of the draw to work with
	currentDraw := sa.copyDraw(draw)
	bestDraw := sa.copyDraw(draw)
	
	currentScore := sa.ConstraintEngine.ScoreDraw(currentDraw)
	bestScore := currentScore
	initialScore := currentScore
	
	temperature := sa.Temperature
	improvements := 0
	acceptances := 0
	
	rand.Seed(time.Now().UnixNano())
	
	for i := 0; i < sa.MaxIterations; i++ {
		// Create a neighbor solution by applying a random modification
		neighbor, err := sa.generateNeighbor(currentDraw)
		if err != nil {
			continue // Skip this iteration if neighbor generation fails
		}
		
		neighborScore := sa.ConstraintEngine.ScoreDraw(neighbor)
		
		// Calculate acceptance probability
		accepted := false
		if neighborScore > currentScore {
			// Better solution - always accept
			accepted = true
			improvements++
		} else if temperature > 0 {
			// Worse solution - accept with probability based on temperature
			delta := neighborScore - currentScore
			probability := math.Exp(delta / temperature)
			if rand.Float64() < probability {
				accepted = true
			}
		}
		
		if accepted {
			currentDraw = neighbor
			currentScore = neighborScore
			acceptances++
			
			// Update best solution if this is the best we've seen
			if currentScore > bestScore {
				bestDraw = sa.copyDraw(currentDraw)
				bestScore = currentScore
			}
		}
		
		// Update temperature
		temperature = sa.CoolingSchedule.NextTemperature(sa.Temperature, i)
		
		// Report progress if callback provided
		if callback != nil && i%100 == 0 {
			acceptanceRate := float64(acceptances) / float64(i+1)
			elapsed := time.Since(startTime)
			remaining := time.Duration(float64(elapsed) * float64(sa.MaxIterations-i) / float64(i+1))
			
			progress := OptimizationProgress{
				Iteration:      i,
				Temperature:    temperature,
				CurrentScore:   currentScore,
				BestScore:      bestScore,
				AcceptanceRate: acceptanceRate,
				EstimatedTime:  remaining.String(),
			}
			callback(progress)
		}
	}
	
	duration := time.Since(startTime)
	
	result := &OptimizationResult{
		InitialScore: initialScore,
		FinalScore:   bestScore,
		Iterations:   sa.MaxIterations,
		Improvements: improvements,
		Duration:     duration,
		BestDraw:     bestDraw,
	}
	
	return result, nil
}

// generateNeighbor creates a neighbor solution by applying a random modification
func (sa *SimulatedAnnealing) generateNeighbor(draw *models.Draw) (*models.Draw, error) {
	neighbor := sa.copyDraw(draw)
	
	// Choose a random modification operation
	operations := []func(*models.Draw) error{
		sa.swapMatches,
		sa.rescheduleMatch,
		sa.swapVenues,
		sa.swapHomeAway,
	}
	
	operation := operations[rand.Intn(len(operations))]
	err := operation(neighbor)
	if err != nil {
		return nil, err
	}
	
	return neighbor, nil
}

// copyDraw creates a deep copy of a draw
func (sa *SimulatedAnnealing) copyDraw(original *models.Draw) *models.Draw {
	copy := &models.Draw{
		ID:               original.ID,
		Name:             original.Name,
		SeasonYear:       original.SeasonYear,
		Rounds:           original.Rounds,
		Status:           original.Status,
		ConstraintConfig: original.ConstraintConfig,
		CreatedAt:        original.CreatedAt,
		UpdatedAt:        original.UpdatedAt,
		Matches:          make([]*models.Match, len(original.Matches)),
	}
	
	// Deep copy matches
	for i, match := range original.Matches {
		copy.Matches[i] = &models.Match{
			ID:          match.ID,
			DrawID:      match.DrawID,
			Round:       match.Round,
			HomeTeamID:  copyIntPtr(match.HomeTeamID),
			AwayTeamID:  copyIntPtr(match.AwayTeamID),
			VenueID:     copyIntPtr(match.VenueID),
			MatchDate:   copyTimePtr(match.MatchDate),
			MatchTime:   copyTimePtr(match.MatchTime),
			IsPrimeTime: match.IsPrimeTime,
			CreatedAt:   match.CreatedAt,
			UpdatedAt:   match.UpdatedAt,
		}
	}
	
	return copy
}

// Helper functions for copying pointers
func copyIntPtr(ptr *int) *int {
	if ptr == nil {
		return nil
	}
	val := *ptr
	return &val
}

func copyTimePtr(ptr *time.Time) *time.Time {
	if ptr == nil {
		return nil
	}
	val := *ptr
	return &val
}