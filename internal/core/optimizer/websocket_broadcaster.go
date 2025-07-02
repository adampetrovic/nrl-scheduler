package optimizer

import (
	"time"
)

// WebSocketBroadcaster defines the interface for broadcasting WebSocket messages
type WebSocketBroadcaster interface {
	BroadcastMessage(messageType string, data interface{})
}

// OptimizationBroadcaster handles broadcasting optimization-related events
type OptimizationBroadcaster struct {
	wsHub WebSocketBroadcaster
}

// NewOptimizationBroadcaster creates a new optimization broadcaster
func NewOptimizationBroadcaster(wsHub WebSocketBroadcaster) *OptimizationBroadcaster {
	return &OptimizationBroadcaster{
		wsHub: wsHub,
	}
}

// BroadcastOptimizationProgress sends optimization progress updates
func (ob *OptimizationBroadcaster) BroadcastOptimizationProgress(jobID string, drawID int, progress OptimizationProgress, maxIterations int) {
	if ob.wsHub == nil {
		return
	}

	// Calculate percentage progress
	progressPercent := float64(progress.Iteration) / float64(maxIterations) * 100.0

	data := map[string]interface{}{
		"job_id":           jobID,
		"draw_id":          drawID,
		"iteration":        progress.Iteration,
		"max_iterations":   maxIterations,
		"current_score":    progress.CurrentScore,
		"best_score":       progress.BestScore,
		"temperature":      progress.Temperature,
		"progress":         progressPercent,
		"updated_at":       time.Now(),
	}

	ob.wsHub.BroadcastMessage("optimization_progress", data)
}

// BroadcastOptimizationCompleted sends optimization completion events
func (ob *OptimizationBroadcaster) BroadcastOptimizationCompleted(jobID string, drawID int, result *OptimizationResult, duration time.Duration) {
	if ob.wsHub == nil {
		return
	}

	data := map[string]interface{}{
		"job_id":      jobID,
		"draw_id":     drawID,
		"completed_at": time.Now(),
		"duration":    duration,
		"final_score": result.FinalScore,
		"iterations":  result.Iterations,
		"improvements": result.Improvements,
	}

	ob.wsHub.BroadcastMessage("optimization_completed", data)
}

// BroadcastOptimizationFailed sends optimization failure events
func (ob *OptimizationBroadcaster) BroadcastOptimizationFailed(jobID string, drawID int, err error) {
	if ob.wsHub == nil {
		return
	}

	data := map[string]interface{}{
		"job_id":   jobID,
		"draw_id":  drawID,
		"error":    err.Error(),
		"failed_at": time.Now(),
	}

	ob.wsHub.BroadcastMessage("optimization_failed", data)
}