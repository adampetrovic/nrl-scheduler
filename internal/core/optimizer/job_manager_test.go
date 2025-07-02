package optimizer

import (
	"testing"
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/constraints"
)

func TestNewJobManager(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	optimizer := NewSimulatedAnnealing(100.0, 0.99, 100, engine)
	jm := NewJobManager(optimizer)

	if jm == nil {
		t.Error("Expected job manager to be created")
	}
	if jm.optimizer != optimizer {
		t.Error("Expected optimizer to be set")
	}
	if jm.jobs == nil {
		t.Error("Expected jobs map to be initialized")
	}
}

func TestStartOptimization(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	optimizer := NewSimulatedAnnealing(100.0, 0.99, 100, engine)
	jm := NewJobManager(optimizer)

	draw := createTestDraw()
	jobID, err := jm.StartOptimization(1, draw)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if jobID == "" {
		t.Error("Expected job ID to be returned")
	}

	// Verify job was created
	job, err := jm.GetJob(jobID)
	if err != nil {
		t.Errorf("Unexpected error getting job: %v", err)
	}
	if job == nil {
		t.Error("Expected job to be created")
	}
	if job.DrawID != 1 {
		t.Errorf("Expected draw ID 1, got %d", job.DrawID)
	}
	if job.Status != JobStatusPending && job.Status != JobStatusRunning {
		t.Errorf("Expected job to be pending or running, got %s", job.Status)
	}
}

func TestGetJob_NotFound(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	optimizer := NewSimulatedAnnealing(100.0, 0.99, 100, engine)
	jm := NewJobManager(optimizer)

	job, err := jm.GetJob("nonexistent")

	if err == nil {
		t.Error("Expected error for nonexistent job")
	}
	if job != nil {
		t.Error("Expected nil job for nonexistent job")
	}
}

func TestCancelJob(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	optimizer := NewSimulatedAnnealing(100.0, 0.99, 1000, engine) // Longer running
	jm := NewJobManager(optimizer)

	draw := createTestDraw()
	jobID, err := jm.StartOptimization(1, draw)
	if err != nil {
		t.Fatalf("Failed to start optimization: %v", err)
	}

	// Give the job a moment to start
	time.Sleep(10 * time.Millisecond)

	err = jm.CancelJob(jobID)
	if err != nil {
		t.Errorf("Unexpected error cancelling job: %v", err)
	}

	// Verify job status
	job, err := jm.GetJob(jobID)
	if err != nil {
		t.Errorf("Unexpected error getting job: %v", err)
	}

	// Wait a bit for the cancellation to process
	time.Sleep(50 * time.Millisecond)
	
	job, _ = jm.GetJob(jobID)
	if job.Status != JobStatusCancelled && job.Status != JobStatusCompleted {
		t.Errorf("Expected job to be cancelled or completed, got %s", job.Status)
	}
}

func TestCancelJob_NotFound(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	optimizer := NewSimulatedAnnealing(100.0, 0.99, 100, engine)
	jm := NewJobManager(optimizer)

	err := jm.CancelJob("nonexistent")

	if err == nil {
		t.Error("Expected error for nonexistent job")
	}
}

func TestListJobs(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	optimizer := NewSimulatedAnnealing(100.0, 0.99, 100, engine)
	jm := NewJobManager(optimizer)

	draw := createTestDraw()
	
	// Start multiple jobs
	jobID1, _ := jm.StartOptimization(1, draw)
	jobID2, _ := jm.StartOptimization(2, draw)

	jobs, err := jm.ListJobs("")
	if err != nil {
		t.Errorf("Unexpected error listing jobs: %v", err)
	}

	if len(jobs) < 2 {
		t.Errorf("Expected at least 2 jobs, got %d", len(jobs))
	}

	// Verify our jobs are in the list
	found1, found2 := false, false
	for _, job := range jobs {
		if job.ID == jobID1 {
			found1 = true
		}
		if job.ID == jobID2 {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Error("Expected both jobs to be in the list")
	}
}

func TestGetJobsByDrawID(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	optimizer := NewSimulatedAnnealing(100.0, 0.99, 100, engine)
	jm := NewJobManager(optimizer)

	draw := createTestDraw()
	
	// Start jobs for different draws
	jobID1, _ := jm.StartOptimization(1, draw)
	jm.StartOptimization(2, draw)

	jobs, err := jm.GetJobsByDrawID(1)
	if err != nil {
		t.Errorf("Unexpected error getting jobs by draw ID: %v", err)
	}

	if len(jobs) != 1 {
		t.Errorf("Expected 1 job for draw 1, got %d", len(jobs))
	}

	if jobs[0].ID != jobID1 {
		t.Error("Expected job ID to match")
	}
}

func TestCleanupCompletedJobs(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	optimizer := NewSimulatedAnnealing(100.0, 0.99, 10, engine) // Quick completion
	jm := NewJobManager(optimizer)

	draw := createTestDraw()
	jobID, _ := jm.StartOptimization(1, draw)

	// Wait for job to complete
	time.Sleep(100 * time.Millisecond)

	// Manually set completion time to past
	job, _ := jm.GetJob(jobID)
	pastTime := time.Now().Add(-2 * time.Hour)
	job.CompletedAt = &pastTime

	// Cleanup jobs older than 1 hour
	jm.CleanupCompletedJobs(1 * time.Hour)

	// Job should be removed
	_, err := jm.GetJob(jobID)
	if err == nil {
		t.Error("Expected job to be cleaned up")
	}
}

func TestGetJobStatistics(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	optimizer := NewSimulatedAnnealing(100.0, 0.99, 100, engine)
	jm := NewJobManager(optimizer)

	draw := createTestDraw()
	
	// Start a job
	jm.StartOptimization(1, draw)

	stats := jm.GetJobStatistics()

	if stats.Total < 1 {
		t.Errorf("Expected at least 1 total job, got %d", stats.Total)
	}

	if stats.Pending+stats.Running+stats.Completed+stats.Cancelled+stats.Failed != stats.Total {
		t.Error("Job statistics don't add up to total")
	}
}

func TestOptimizationProgress(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	optimizer := NewSimulatedAnnealing(100.0, 0.99, 100, engine)
	jm := NewJobManager(optimizer)

	draw := createTestDraw()
	jobID, _ := jm.StartOptimization(1, draw)

	// Wait a bit for optimization to start
	time.Sleep(50 * time.Millisecond)

	job, err := jm.GetJob(jobID)
	if err != nil {
		t.Errorf("Unexpected error getting job: %v", err)
	}

	// Progress should be initialized
	if job.Progress.Iteration < 0 {
		t.Error("Expected non-negative iteration")
	}
	if job.Progress.Temperature < 0 {
		t.Error("Expected non-negative temperature")
	}
}

func TestJobTimeout(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	optimizer := NewSimulatedAnnealing(100.0, 0.99, 1, engine) // Very quick
	jm := NewJobManager(optimizer)

	draw := createTestDraw()
	jobID, _ := jm.StartOptimization(1, draw)

	// Wait for job to complete
	timeout := time.After(1 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Error("Job did not complete within timeout")
			return
		case <-ticker.C:
			job, _ := jm.GetJob(jobID)
			if job.Status == JobStatusCompleted || job.Status == JobStatusFailed {
				// Job completed successfully
				if job.CompletedAt == nil {
					t.Error("Expected completion time to be set")
				}
				return
			}
		}
	}
}

func TestConcurrentJobs(t *testing.T) {
	engine := constraints.NewConstraintEngine()
	optimizer := NewSimulatedAnnealing(100.0, 0.99, 100, engine)
	jm := NewJobManager(optimizer)

	draw := createTestDraw()
	
	// Start multiple jobs concurrently
	jobIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		jobID, err := jm.StartOptimization(i+1, draw)
		if err != nil {
			t.Errorf("Failed to start job %d: %v", i, err)
		}
		jobIDs[i] = jobID
	}

	// Verify all jobs were created
	for i, jobID := range jobIDs {
		job, err := jm.GetJob(jobID)
		if err != nil {
			t.Errorf("Failed to get job %d: %v", i, err)
		}
		if job.DrawID != i+1 {
			t.Errorf("Expected draw ID %d, got %d", i+1, job.DrawID)
		}
	}
}

func BenchmarkStartOptimization(b *testing.B) {
	engine := constraints.NewConstraintEngine()
	optimizer := NewSimulatedAnnealing(100.0, 0.99, 10, engine)
	jm := NewJobManager(optimizer)
	draw := createTestDraw()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jm.StartOptimization(i, draw)
	}
}

func BenchmarkGetJob(b *testing.B) {
	engine := constraints.NewConstraintEngine()
	optimizer := NewSimulatedAnnealing(100.0, 0.99, 10, engine)
	jm := NewJobManager(optimizer)
	draw := createTestDraw()

	jobID, _ := jm.StartOptimization(1, draw)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jm.GetJob(jobID)
	}
}