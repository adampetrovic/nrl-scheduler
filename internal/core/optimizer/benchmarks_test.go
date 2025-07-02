package optimizer

import (
	"testing"
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/constraints"
	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// createLargeDraw creates a draw with more matches for performance testing
func createLargeDraw(teams, rounds int) *models.Draw {
	draw := &models.Draw{
		ID:         1,
		Name:       "Large Test Draw",
		SeasonYear: 2025,
		Rounds:     rounds,
		Status:     models.DrawStatusDraft,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Matches:    make([]*models.Match, 0),
	}

	matchID := 1
	for round := 1; round <= rounds; round++ {
		for homeTeam := 1; homeTeam <= teams; homeTeam++ {
			for awayTeam := homeTeam + 1; awayTeam <= teams; awayTeam++ {
				venue := homeTeam // Home team's venue
				match := &models.Match{
					ID:         matchID,
					DrawID:     1,
					Round:      round,
					HomeTeamID: &homeTeam,
					AwayTeamID: &awayTeam,
					VenueID:    &venue,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
				draw.Matches = append(draw.Matches, match)
				matchID++
			}
		}
	}

	return draw
}

// BenchmarkOptimizeSmallDraw benchmarks optimization on a small draw
func BenchmarkOptimizeSmallDraw(b *testing.B) {
	engine := constraints.NewConstraintEngine()
	// Add some constraints to make it more realistic
	engine.AddSoftConstraint(constraints.NewHomeAwayBalanceConstraint(0.1), 0.8)
	engine.AddSoftConstraint(constraints.NewTravelMinimizationConstraint(3), 0.6)

	sa := NewSimulatedAnnealing(50.0, 0.98, 100, engine)
	draw := createTestDraw() // 4 matches

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sa.Optimize(draw, nil)
	}
}

// BenchmarkOptimizeMediumDraw benchmarks optimization on a medium-sized draw
func BenchmarkOptimizeMediumDraw(b *testing.B) {
	engine := constraints.NewConstraintEngine()
	engine.AddSoftConstraint(constraints.NewHomeAwayBalanceConstraint(0.1), 0.8)
	engine.AddSoftConstraint(constraints.NewTravelMinimizationConstraint(3), 0.6)

	sa := NewSimulatedAnnealing(50.0, 0.98, 100, engine)
	draw := createLargeDraw(6, 2) // 30 matches (6 teams, 2 rounds)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sa.Optimize(draw, nil)
	}
}

// BenchmarkOptimizeLargeDraw benchmarks optimization on a large draw
func BenchmarkOptimizeLargeDraw(b *testing.B) {
	engine := constraints.NewConstraintEngine()
	engine.AddSoftConstraint(constraints.NewHomeAwayBalanceConstraint(0.1), 0.8)
	engine.AddSoftConstraint(constraints.NewTravelMinimizationConstraint(3), 0.6)

	sa := NewSimulatedAnnealing(50.0, 0.98, 50, engine) // Fewer iterations for large draw
	draw := createLargeDraw(8, 2) // 56 matches (8 teams, 2 rounds)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sa.Optimize(draw, nil)
	}
}

// BenchmarkConstraintEvaluation benchmarks constraint evaluation
func BenchmarkConstraintEvaluation(b *testing.B) {
	engine := constraints.NewConstraintEngine()
	engine.AddHardConstraint(constraints.NewByeConstraint())
	engine.AddHardConstraint(constraints.NewDoubleUpConstraint(8))
	engine.AddSoftConstraint(constraints.NewHomeAwayBalanceConstraint(0.1), 0.8)
	engine.AddSoftConstraint(constraints.NewTravelMinimizationConstraint(3), 0.6)
	engine.AddSoftConstraint(constraints.NewRestPeriodConstraint(5), 0.7)
	engine.AddSoftConstraint(constraints.NewPrimeTimeSpreadConstraint(0.3, 0.1), 0.5)

	draw := createLargeDraw(6, 3) // 45 matches

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.ScoreDraw(draw)
	}
}

// BenchmarkDrawCopy benchmarks draw copying performance
func BenchmarkDrawCopy(b *testing.B) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)
	draw := createLargeDraw(8, 2) // 56 matches

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sa.copyDraw(draw)
	}
}

// BenchmarkNeighborGeneration benchmarks neighbor generation
func BenchmarkNeighborGeneration(b *testing.B) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)
	draw := createLargeDraw(6, 3) // 45 matches

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sa.generateNeighbor(draw)
	}
}

// BenchmarkOperations benchmarks individual operations
func BenchmarkSwapMatches(b *testing.B) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)
	draw := createLargeDraw(6, 3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset draw state for consistent benchmarking
		testDraw := sa.copyDraw(draw)
		sa.swapMatches(testDraw)
	}
}

func BenchmarkRescheduleMatch(b *testing.B) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)
	draw := createLargeDraw(6, 3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testDraw := sa.copyDraw(draw)
		sa.rescheduleMatch(testDraw)
	}
}

func BenchmarkSwapVenues(b *testing.B) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)
	draw := createLargeDraw(6, 3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testDraw := sa.copyDraw(draw)
		sa.swapVenues(testDraw)
	}
}

func BenchmarkSwapHomeAway(b *testing.B) {
	engine := constraints.NewConstraintEngine()
	sa := NewSimulatedAnnealing(100.0, 0.99, 100, engine)
	draw := createLargeDraw(6, 3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testDraw := sa.copyDraw(draw)
		sa.swapHomeAway(testDraw)
	}
}

// BenchmarkCoolingSchedules benchmarks different cooling schedules
func BenchmarkExponentialCooling(b *testing.B) {
	cooling := NewExponentialCooling(0.99)
	initialTemp := 100.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cooling.NextTemperature(initialTemp, i%1000)
	}
}

func BenchmarkLinearCooling(b *testing.B) {
	cooling := NewLinearCooling(0.1)
	initialTemp := 100.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cooling.NextTemperature(initialTemp, i%1000)
	}
}

func BenchmarkAdaptiveCooling(b *testing.B) {
	cooling := NewAdaptiveCooling(0.99, 0.4, 0.1)
	initialTemp := 100.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cooling.NextTemperature(initialTemp, i%1000)
	}
}

func BenchmarkLogarithmicCooling(b *testing.B) {
	cooling := NewLogarithmicCooling(1.0)
	initialTemp := 100.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cooling.NextTemperature(initialTemp, i%1000+1) // +1 to avoid log(1)
	}
}

// BenchmarkJobManager benchmarks job management operations
func BenchmarkJobManagerOperations(b *testing.B) {
	engine := constraints.NewConstraintEngine()
	optimizer := NewSimulatedAnnealing(50.0, 0.99, 10, engine) // Quick jobs
	jm := NewJobManager(optimizer)
	draw := createTestDraw()

	b.Run("StartOptimization", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jm.StartOptimization(i, draw)
		}
	})

	// Start a job for other benchmarks
	jobID, _ := jm.StartOptimization(1, draw)

	b.Run("GetJob", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jm.GetJob(jobID)
		}
	})

	b.Run("GetJobStatistics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jm.GetJobStatistics()
		}
	})
}

// BenchmarkMemoryUsage provides insights into memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	engine := constraints.NewConstraintEngine()
	engine.AddSoftConstraint(constraints.NewHomeAwayBalanceConstraint(0.1), 0.8)
	
	sa := NewSimulatedAnnealing(100.0, 0.98, 100, engine)

	b.Run("SmallDraw", func(b *testing.B) {
		draw := createTestDraw()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sa.copyDraw(draw)
		}
	})

	b.Run("MediumDraw", func(b *testing.B) {
		draw := createLargeDraw(6, 2)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sa.copyDraw(draw)
		}
	})

	b.Run("LargeDraw", func(b *testing.B) {
		draw := createLargeDraw(8, 2)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sa.copyDraw(draw)
		}
	})
}