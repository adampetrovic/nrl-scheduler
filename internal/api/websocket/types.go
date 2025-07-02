package websocket

import (
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/constraints"
	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
	"github.com/adampetrovic/nrl-scheduler/internal/core/optimizer"
)

// Message types for WebSocket communication
const (
	// Optimization events
	OptimizationStarted   = "optimization_started"
	OptimizationProgress  = "optimization_progress"
	OptimizationCompleted = "optimization_completed"
	OptimizationFailed    = "optimization_failed"
	OptimizationCancelled = "optimization_cancelled"

	// Draw events
	DrawCreated        = "draw_created"
	DrawUpdated        = "draw_updated"
	DrawDeleted        = "draw_deleted"
	DrawGenerated      = "draw_generated"
	DrawStatusChanged  = "draw_status_changed"

	// Match events
	MatchUpdated = "match_updated"
	MatchCreated = "match_created"
	MatchDeleted = "match_deleted"

	// Constraint events
	ConstraintViolation = "constraint_violation"
	ConstraintsValidated = "constraints_validated"

	// System events
	SystemStatus = "system_status"
	ClientCount  = "client_count"
)

// OptimizationStartedData represents the data for optimization started events
type OptimizationStartedData struct {
	JobID       string    `json:"job_id"`
	DrawID      int       `json:"draw_id"`
	StartedAt   time.Time `json:"started_at"`
	Config      optimizer.OptimizationConfig `json:"config"`
}

// OptimizationProgressData represents the data for optimization progress events
type OptimizationProgressData struct {
	JobID           string    `json:"job_id"`
	DrawID          int       `json:"draw_id"`
	Iteration       int       `json:"iteration"`
	MaxIterations   int       `json:"max_iterations"`
	CurrentScore    float64   `json:"current_score"`
	BestScore       float64   `json:"best_score"`
	Temperature     float64   `json:"temperature"`
	Progress        float64   `json:"progress"`
	EstimatedTimeRemaining time.Duration `json:"estimated_time_remaining,omitempty"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// OptimizationCompletedData represents the data for optimization completed events
type OptimizationCompletedData struct {
	JobID         string    `json:"job_id"`
	DrawID        int       `json:"draw_id"`
	CompletedAt   time.Time `json:"completed_at"`
	Duration      time.Duration `json:"duration"`
	FinalScore    float64   `json:"final_score"`
	Iterations    int       `json:"iterations"`
	Improvements  int       `json:"improvements"`
}

// OptimizationFailedData represents the data for optimization failed events
type OptimizationFailedData struct {
	JobID     string    `json:"job_id"`
	DrawID    int       `json:"draw_id"`
	Error     string    `json:"error"`
	FailedAt  time.Time `json:"failed_at"`
}

// OptimizationCancelledData represents the data for optimization cancelled events
type OptimizationCancelledData struct {
	JobID       string    `json:"job_id"`
	DrawID      int       `json:"draw_id"`
	CancelledAt time.Time `json:"cancelled_at"`
	Reason      string    `json:"reason,omitempty"`
}

// DrawEventData represents the data for draw-related events
type DrawEventData struct {
	Draw      *models.Draw `json:"draw"`
	Timestamp time.Time    `json:"timestamp"`
	UserID    string       `json:"user_id,omitempty"`
}

// MatchEventData represents the data for match-related events
type MatchEventData struct {
	Match     *models.Match `json:"match"`
	DrawID    int           `json:"draw_id"`
	Timestamp time.Time     `json:"timestamp"`
	UserID    string        `json:"user_id,omitempty"`
}

// ConstraintViolationData represents the data for constraint violation events
type ConstraintViolationData struct {
	DrawID      int                               `json:"draw_id"`
	Violations  []constraints.ConstraintViolation `json:"violations"`
	TotalCount  int                               `json:"total_count"`
	Severity    string                            `json:"severity"`
	Timestamp   time.Time                         `json:"timestamp"`
}

// ConstraintsValidatedData represents the data for constraint validation events
type ConstraintsValidatedData struct {
	DrawID        int                               `json:"draw_id"`
	IsValid       bool                              `json:"is_valid"`
	Violations    []constraints.ConstraintViolation `json:"violations"`
	Score         float64                           `json:"score"`
	ValidatedAt   time.Time                         `json:"validated_at"`
}

// SystemStatusData represents the data for system status events
type SystemStatusData struct {
	Status             string    `json:"status"`
	ActiveOptimizations int      `json:"active_optimizations"`
	ConnectedClients   int       `json:"connected_clients"`
	Timestamp          time.Time `json:"timestamp"`
	Memory             struct {
		Used      uint64  `json:"used"`
		Available uint64  `json:"available"`
		Percent   float64 `json:"percent"`
	} `json:"memory,omitempty"`
}

// ClientCountData represents the data for client count events
type ClientCountData struct {
	Count     int       `json:"count"`
	Timestamp time.Time `json:"timestamp"`
}