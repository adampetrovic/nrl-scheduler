package optimizer

import (
	"context"
	"fmt"

	"github.com/adampetrovic/nrl-scheduler/internal/core/constraints"
	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
	"github.com/adampetrovic/nrl-scheduler/internal/storage"
)

// Service provides optimization functionality integrated with the storage layer
type Service struct {
	repository       storage.Repositories
	constraintEngine *constraints.ConstraintEngine
	jobManager       *JobManager
}

// NewService creates a new optimizer service
func NewService(repository storage.Repositories) *Service {
	// Create constraint engine
	constraintEngine := constraints.NewConstraintEngine()
	
	// Create optimizer with default settings
	optimizer := NewSimulatedAnnealing(100.0, 0.99, 10000, constraintEngine)
	
	// Create job manager
	jobManager := NewJobManager(optimizer)
	
	return &Service{
		repository:       repository,
		constraintEngine: constraintEngine,
		jobManager:       jobManager,
	}
}

// OptimizeDraw starts optimization for a specific draw
func (s *Service) OptimizeDraw(drawID int, config OptimizationConfig) (string, error) {
	// Fetch the draw from storage
	draw, err := s.repository.Draws().GetWithMatches(context.Background(), drawID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch draw: %w", err)
	}
	
	// Load constraint configuration if present
	if err := s.loadConstraintConfig(draw); err != nil {
		return "", fmt.Errorf("failed to load constraint config: %w", err)
	}
	
	// Create optimizer with the provided config
	optimizer := NewSimulatedAnnealing(
		config.Temperature,
		config.CoolingRate,
		config.MaxIterations,
		s.constraintEngine,
	)
	
	// Set cooling schedule if specified
	if config.CoolingSchedule.Type != "" {
		optimizer.CoolingSchedule = CreateCoolingSchedule(config.CoolingSchedule)
	}
	
	// Update job manager with new optimizer
	s.jobManager.optimizer = optimizer
	
	// Mark draw as optimizing
	draw.Status = models.DrawStatusOptimizing
	if err := s.repository.Draws().Update(context.Background(), draw); err != nil {
		return "", fmt.Errorf("failed to update draw status: %w", err)
	}
	
	// Start optimization job
	jobID, err := s.jobManager.StartOptimization(drawID, draw)
	if err != nil {
		// Revert draw status on error
		draw.Status = models.DrawStatusDraft
		s.repository.Draws().Update(context.Background(), draw)
		return "", fmt.Errorf("failed to start optimization: %w", err)
	}
	
	return jobID, nil
}

// GetOptimizationJob returns information about an optimization job
func (s *Service) GetOptimizationJob(jobID string) (*OptimizationJob, error) {
	return s.jobManager.GetJob(jobID)
}

// CancelOptimization cancels a running optimization job
func (s *Service) CancelOptimization(jobID string) error {
	job, err := s.jobManager.GetJob(jobID)
	if err != nil {
		return err
	}
	
	// Cancel the job
	if err := s.jobManager.CancelJob(jobID); err != nil {
		return err
	}
	
	// Update draw status back to draft
	if job.Status == JobStatusRunning {
		draw, err := s.repository.Draws().Get(context.Background(), job.DrawID)
		if err == nil {
			draw.Status = models.DrawStatusDraft
			s.repository.Draws().Update(context.Background(), draw)
		}
	}
	
	return nil
}

// GetOptimizationResult returns the result of a completed optimization
func (s *Service) GetOptimizationResult(jobID string) (*OptimizationResult, error) {
	job, err := s.jobManager.GetJob(jobID)
	if err != nil {
		return nil, err
	}
	
	if job.Status != JobStatusCompleted {
		return nil, fmt.Errorf("optimization job has not completed")
	}
	
	if job.Result == nil {
		return nil, fmt.Errorf("optimization result not available")
	}
	
	return job.Result, nil
}

// ApplyOptimizationResult applies the optimized draw to storage
func (s *Service) ApplyOptimizationResult(jobID string) error {
	job, err := s.jobManager.GetJob(jobID)
	if err != nil {
		return err
	}
	
	if job.Status != JobStatusCompleted || job.Result == nil {
		return fmt.Errorf("optimization job not completed or result not available")
	}
	
	// Update draw with optimized matches
	optimizedDraw := job.Result.BestDraw
	optimizedDraw.Status = models.DrawStatusCompleted
	
	if err := s.repository.Draws().Update(context.Background(), optimizedDraw); err != nil {
		return fmt.Errorf("failed to update draw: %w", err)
	}
	
	// Update all matches
	for _, match := range optimizedDraw.Matches {
		if err := s.repository.Matches().Update(context.Background(), match); err != nil {
			return fmt.Errorf("failed to update match %d: %w", match.ID, err)
		}
	}
	
	return nil
}

// ValidateDrawConstraints validates a draw against all configured constraints
func (s *Service) ValidateDrawConstraints(drawID int) ([]constraints.ConstraintViolation, error) {
	draw, err := s.repository.Draws().GetWithMatches(context.Background(), drawID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch draw: %w", err)
	}
	
	// Load constraint configuration
	if err := s.loadConstraintConfig(draw); err != nil {
		return nil, fmt.Errorf("failed to load constraint config: %w", err)
	}
	
	// Analyze the draw
	violations := s.constraintEngine.AnalyzeDraw(draw)
	return violations, nil
}

// ScoreDraw calculates the constraint satisfaction score for a draw
func (s *Service) ScoreDraw(drawID int) (float64, error) {
	draw, err := s.repository.Draws().GetWithMatches(context.Background(), drawID)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch draw: %w", err)
	}
	
	// Load constraint configuration
	if err := s.loadConstraintConfig(draw); err != nil {
		return 0, fmt.Errorf("failed to load constraint config: %w", err)
	}
	
	// Calculate score
	score := s.constraintEngine.ScoreDraw(draw)
	return score, nil
}

// ListOptimizationJobs returns optimization jobs, optionally filtered by draw ID
func (s *Service) ListOptimizationJobs(drawID int) ([]*OptimizationJob, error) {
	if drawID > 0 {
		return s.jobManager.GetJobsByDrawID(drawID)
	}
	return s.jobManager.ListJobs("")
}

// GetJobStatistics returns statistics about optimization jobs
func (s *Service) GetJobStatistics() JobStatistics {
	return s.jobManager.GetJobStatistics()
}

// loadConstraintConfig loads and configures constraints from the draw's configuration
func (s *Service) loadConstraintConfig(draw *models.Draw) error {
	if draw.ConstraintConfig == nil {
		// Use default constraints if none specified
		return s.loadDefaultConstraints()
	}
	
	// Parse constraint configuration from JSON
	config, err := constraints.LoadConstraintConfigFromJSON(draw.ConstraintConfig)
	if err != nil {
		return fmt.Errorf("failed to parse constraint config: %w", err)
	}
	
	// Create constraint engine from configuration
	factory := constraints.NewConstraintFactory()
	engine, err := factory.CreateConstraintEngine(config)
	if err != nil {
		return fmt.Errorf("failed to create constraint engine: %w", err)
	}
	
	s.constraintEngine = engine
	return nil
}

// loadDefaultConstraints loads a default set of NRL constraints
func (s *Service) loadDefaultConstraints() error {
	// Get default NRL constraint configuration
	config := constraints.GetDefaultNRLConstraintConfig()
	
	// Create constraint engine from configuration
	factory := constraints.NewConstraintFactory()
	engine, err := factory.CreateConstraintEngine(config)
	if err != nil {
		return fmt.Errorf("failed to create default constraint engine: %w", err)
	}
	
	s.constraintEngine = engine
	return nil
}

// GetConstraintEngine returns the constraint engine for direct access
func (s *Service) GetConstraintEngine() *constraints.ConstraintEngine {
	return s.constraintEngine
}

// GetJobManager returns the job manager for direct access
func (s *Service) GetJobManager() *JobManager {
	return s.jobManager
}

// SetOptimizationConfig updates the optimizer configuration
func (s *Service) SetOptimizationConfig(config OptimizationConfig) {
	optimizer := NewSimulatedAnnealing(
		config.Temperature,
		config.CoolingRate,
		config.MaxIterations,
		s.constraintEngine,
	)
	
	if config.CoolingSchedule.Type != "" {
		optimizer.CoolingSchedule = CreateCoolingSchedule(config.CoolingSchedule)
	}
	
	s.jobManager.optimizer = optimizer
}