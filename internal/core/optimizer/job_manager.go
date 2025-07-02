package optimizer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// JobStatus represents the status of an optimization job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusCancelled JobStatus = "cancelled"
	JobStatusFailed    JobStatus = "failed"
)

// OptimizationJob represents a running optimization job
type OptimizationJob struct {
	ID          string                `json:"id"`
	DrawID      int                   `json:"draw_id"`
	Status      JobStatus             `json:"status"`
	Progress    OptimizationProgress  `json:"progress"`
	Result      *OptimizationResult   `json:"result,omitempty"`
	Error       string                `json:"error,omitempty"`
	StartedAt   time.Time             `json:"started_at"`
	CompletedAt *time.Time            `json:"completed_at,omitempty"`
	CancelFunc  context.CancelFunc    `json:"-"`
}

// JobManager manages optimization jobs
type JobManager struct {
	jobs        map[string]*OptimizationJob
	mutex       sync.RWMutex
	optimizer   *SimulatedAnnealing
	broadcaster *OptimizationBroadcaster
}

// NewJobManager creates a new job manager
func NewJobManager(optimizer *SimulatedAnnealing) *JobManager {
	return &JobManager{
		jobs:      make(map[string]*OptimizationJob),
		optimizer: optimizer,
	}
}

// SetBroadcaster sets the WebSocket broadcaster for real-time updates
func (jm *JobManager) SetBroadcaster(broadcaster *OptimizationBroadcaster) {
	jm.broadcaster = broadcaster
}

// StartOptimization starts a new optimization job
func (jm *JobManager) StartOptimization(drawID int, draw *models.Draw) (string, error) {
	jobID := fmt.Sprintf("opt_%d_%d", drawID, time.Now().Unix())
	
	ctx, cancel := context.WithCancel(context.Background())
	
	job := &OptimizationJob{
		ID:         jobID,
		DrawID:     drawID,
		Status:     JobStatusPending,
		StartedAt:  time.Now(),
		CancelFunc: cancel,
	}
	
	jm.mutex.Lock()
	jm.jobs[jobID] = job
	jm.mutex.Unlock()
	
	// Start optimization in a goroutine
	go jm.runOptimization(ctx, job, draw)
	
	return jobID, nil
}

// runOptimization executes the optimization algorithm
func (jm *JobManager) runOptimization(ctx context.Context, job *OptimizationJob, draw *models.Draw) {
	jm.updateJobStatus(job.ID, JobStatusRunning)
	startTime := time.Now()
	
	// Create progress callback
	progressCallback := func(progress OptimizationProgress) {
		jm.updateJobProgress(job.ID, progress)
		
		// Broadcast progress update
		if jm.broadcaster != nil {
			jm.broadcaster.BroadcastOptimizationProgress(job.ID, job.DrawID, progress, jm.optimizer.MaxIterations)
		}
		
		// Check for cancellation
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
	
	// Run the optimization
	result, err := jm.optimizer.Optimize(draw, progressCallback)
	
	// Check if job was cancelled
	select {
	case <-ctx.Done():
		jm.updateJobStatus(job.ID, JobStatusCancelled)
		return
	default:
	}
	
	// Update job with result
	completedAt := time.Now()
	duration := completedAt.Sub(startTime)
	
	jm.mutex.Lock()
	if err != nil {
		job.Status = JobStatusFailed
		job.Error = err.Error()
		// Broadcast failure
		if jm.broadcaster != nil {
			jm.broadcaster.BroadcastOptimizationFailed(job.ID, job.DrawID, err)
		}
	} else {
		job.Status = JobStatusCompleted
		job.Result = result
		// Broadcast completion
		if jm.broadcaster != nil {
			jm.broadcaster.BroadcastOptimizationCompleted(job.ID, job.DrawID, result, duration)
		}
	}
	job.CompletedAt = &completedAt
	jm.mutex.Unlock()
}

// GetJob returns information about a specific job
func (jm *JobManager) GetJob(jobID string) (*OptimizationJob, error) {
	jm.mutex.RLock()
	defer jm.mutex.RUnlock()
	
	job, exists := jm.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job %s not found", jobID)
	}
	
	return job, nil
}

// CancelJob cancels a running optimization job
func (jm *JobManager) CancelJob(jobID string) error {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	
	job, exists := jm.jobs[jobID]
	if !exists {
		return fmt.Errorf("job %s not found", jobID)
	}
	
	if job.Status == JobStatusRunning {
		job.CancelFunc()
		job.Status = JobStatusCancelled
		completedAt := time.Now()
		job.CompletedAt = &completedAt
	}
	
	return nil
}

// ListJobs returns all jobs, optionally filtered by status
func (jm *JobManager) ListJobs(status JobStatus) ([]*OptimizationJob, error) {
	jm.mutex.RLock()
	defer jm.mutex.RUnlock()
	
	var jobs []*OptimizationJob
	
	for _, job := range jm.jobs {
		if status == "" || job.Status == status {
			jobs = append(jobs, job)
		}
	}
	
	return jobs, nil
}

// GetJobsByDrawID returns all jobs for a specific draw
func (jm *JobManager) GetJobsByDrawID(drawID int) ([]*OptimizationJob, error) {
	jm.mutex.RLock()
	defer jm.mutex.RUnlock()
	
	var jobs []*OptimizationJob
	
	for _, job := range jm.jobs {
		if job.DrawID == drawID {
			jobs = append(jobs, job)
		}
	}
	
	return jobs, nil
}

// CleanupCompletedJobs removes completed jobs older than the specified duration
func (jm *JobManager) CleanupCompletedJobs(maxAge time.Duration) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	
	cutoff := time.Now().Add(-maxAge)
	
	for jobID, job := range jm.jobs {
		if job.CompletedAt != nil && job.CompletedAt.Before(cutoff) {
			delete(jm.jobs, jobID)
		}
	}
}

// updateJobStatus updates the status of a job
func (jm *JobManager) updateJobStatus(jobID string, status JobStatus) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	
	if job, exists := jm.jobs[jobID]; exists {
		job.Status = status
	}
}

// updateJobProgress updates the progress of a job
func (jm *JobManager) updateJobProgress(jobID string, progress OptimizationProgress) {
	jm.mutex.Lock()
	defer jm.mutex.Unlock()
	
	if job, exists := jm.jobs[jobID]; exists {
		job.Progress = progress
	}
}

// GetJobStatistics returns statistics about jobs
func (jm *JobManager) GetJobStatistics() JobStatistics {
	jm.mutex.RLock()
	defer jm.mutex.RUnlock()
	
	stats := JobStatistics{
		Total: len(jm.jobs),
	}
	
	for _, job := range jm.jobs {
		switch job.Status {
		case JobStatusPending:
			stats.Pending++
		case JobStatusRunning:
			stats.Running++
		case JobStatusCompleted:
			stats.Completed++
		case JobStatusCancelled:
			stats.Cancelled++
		case JobStatusFailed:
			stats.Failed++
		}
	}
	
	return stats
}

// JobStatistics contains statistics about optimization jobs
type JobStatistics struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Running   int `json:"running"`
	Completed int `json:"completed"`
	Cancelled int `json:"cancelled"`
	Failed    int `json:"failed"`
}

// OptimizationConfig contains configuration for optimization jobs
type OptimizationConfig struct {
	Temperature     float64                   `json:"temperature"`
	CoolingRate     float64                   `json:"cooling_rate"`
	MaxIterations   int                       `json:"max_iterations"`
	CoolingSchedule TemperatureScheduleConfig `json:"cooling_schedule"`
}

// DefaultOptimizationConfig returns a default configuration
func DefaultOptimizationConfig() OptimizationConfig {
	return OptimizationConfig{
		Temperature:   100.0,
		CoolingRate:   0.99,
		MaxIterations: 10000,
		CoolingSchedule: TemperatureScheduleConfig{
			Type:        "exponential",
			CoolingRate: 0.99,
		},
	}
}