package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/adampetrovic/nrl-scheduler/internal/core/optimizer"
	"github.com/adampetrovic/nrl-scheduler/pkg/types"
)

// OptimizationHandler handles optimization-related HTTP requests
type OptimizationHandler struct {
	optimizerService *optimizer.Service
}

// NewOptimizationHandler creates a new optimization handler
func NewOptimizationHandler(optimizerService *optimizer.Service) *OptimizationHandler {
	return &OptimizationHandler{
		optimizerService: optimizerService,
	}
}

// StartOptimization starts optimization for a specific draw
// POST /api/v1/optimize/:drawId/start
func (h *OptimizationHandler) StartOptimization(c *gin.Context) {
	drawIDStr := c.Param("drawId")
	drawID, err := strconv.Atoi(drawIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid draw ID",
			Details: map[string]string{
				"draw_id": "must be a valid integer",
			},
		})
		return
	}

	var request types.StartOptimizationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid request body",
			Details: map[string]string{
				"json": err.Error(),
			},
		})
		return
	}

	// Convert request to optimization config
	config := optimizer.OptimizationConfig{
		Temperature:   request.Temperature,
		CoolingRate:   request.CoolingRate,
		MaxIterations: request.MaxIterations,
	}

	if request.CoolingSchedule != nil {
		config.CoolingSchedule = optimizer.TemperatureScheduleConfig{
			Type:             request.CoolingSchedule.Type,
			CoolingRate:      request.CoolingSchedule.CoolingRate,
			ScalingFactor:    request.CoolingSchedule.ScalingFactor,
			ReheatFactor:     request.CoolingSchedule.ReheatFactor,
			ReheatPeriod:     request.CoolingSchedule.ReheatPeriod,
			AcceptanceTarget: request.CoolingSchedule.AcceptanceTarget,
			AdaptationFactor: request.CoolingSchedule.AdaptationFactor,
			Params:           request.CoolingSchedule.Params,
		}
	}

	jobID, err := h.optimizerService.OptimizeDraw(drawID, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to start optimization",
			Details: map[string]string{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusAccepted, types.StartOptimizationResponse{
		JobID: jobID,
		Status: "started",
	})
}

// GetOptimizationStatus returns the status of an optimization job
// GET /api/v1/optimize/:jobId/status
func (h *OptimizationHandler) GetOptimizationStatus(c *gin.Context) {
	jobID := c.Param("jobId")

	job, err := h.optimizerService.GetOptimizationJob(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: "Optimization job not found",
			Details: map[string]string{
				"job_id": jobID,
			},
		})
		return
	}

	response := types.OptimizationStatusResponse{
		JobID:       job.ID,
		DrawID:      job.DrawID,
		Status:      string(job.Status),
		Progress:    job.Progress,
		StartedAt:   job.StartedAt,
		CompletedAt: job.CompletedAt,
	}

	if job.Error != "" {
		response.Error = &job.Error
	}

	c.JSON(http.StatusOK, response)
}

// CancelOptimization cancels a running optimization job
// POST /api/v1/optimize/:jobId/cancel
func (h *OptimizationHandler) CancelOptimization(c *gin.Context) {
	jobID := c.Param("jobId")

	err := h.optimizerService.CancelOptimization(jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to cancel optimization",
			Details: map[string]string{
				"job_id": jobID,
				"error":  err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "cancelled",
		"job_id": jobID,
	})
}

// GetOptimizationResult returns the result of a completed optimization
// GET /api/v1/optimize/:jobId/result
func (h *OptimizationHandler) GetOptimizationResult(c *gin.Context) {
	jobID := c.Param("jobId")

	result, err := h.optimizerService.GetOptimizationResult(jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, types.ErrorResponse{
			Error: "Optimization result not available",
			Details: map[string]string{
				"job_id": jobID,
				"error":  err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ApplyOptimizationResult applies the optimized draw to storage
// POST /api/v1/optimize/:jobId/apply
func (h *OptimizationHandler) ApplyOptimizationResult(c *gin.Context) {
	jobID := c.Param("jobId")

	err := h.optimizerService.ApplyOptimizationResult(jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to apply optimization result",
			Details: map[string]string{
				"job_id": jobID,
				"error":  err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "applied",
		"job_id": jobID,
	})
}

// ValidateDrawConstraints validates a draw against all configured constraints
// GET /api/v1/draws/:drawId/validate-constraints
func (h *OptimizationHandler) ValidateDrawConstraints(c *gin.Context) {
	drawIDStr := c.Param("drawId")
	drawID, err := strconv.Atoi(drawIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid draw ID",
			Details: map[string]string{
				"draw_id": "must be a valid integer",
			},
		})
		return
	}

	violations, err := h.optimizerService.ValidateDrawConstraints(drawID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to validate constraints",
			Details: map[string]string{
				"error": err.Error(),
			},
		})
		return
	}

	response := types.ConstraintValidationResponse{
		DrawID:     drawID,
		IsValid:    len(violations) == 0,
		Violations: violations,
	}

	c.JSON(http.StatusOK, response)
}

// ScoreDraw calculates the constraint satisfaction score for a draw
// GET /api/v1/draws/:drawId/score
func (h *OptimizationHandler) ScoreDraw(c *gin.Context) {
	drawIDStr := c.Param("drawId")
	drawID, err := strconv.Atoi(drawIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid draw ID",
			Details: map[string]string{
				"draw_id": "must be a valid integer",
			},
		})
		return
	}

	score, err := h.optimizerService.ScoreDraw(drawID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to calculate draw score",
			Details: map[string]string{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, types.DrawScoreResponse{
		DrawID: drawID,
		Score:  score,
	})
}

// ListOptimizationJobs returns optimization jobs, optionally filtered by draw ID
// GET /api/v1/optimize/jobs
func (h *OptimizationHandler) ListOptimizationJobs(c *gin.Context) {
	drawIDStr := c.Query("draw_id")
	var drawID int
	var err error

	if drawIDStr != "" {
		drawID, err = strconv.Atoi(drawIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error: "Invalid draw ID filter",
				Details: map[string]string{
					"draw_id": "must be a valid integer",
				},
			})
			return
		}
	}

	jobs, err := h.optimizerService.ListOptimizationJobs(drawID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: "Failed to list optimization jobs",
			Details: map[string]string{
				"error": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, types.OptimizationJobsResponse{
		Jobs: jobs,
	})
}

// GetJobStatistics returns statistics about optimization jobs
// GET /api/v1/optimize/statistics
func (h *OptimizationHandler) GetJobStatistics(c *gin.Context) {
	stats := h.optimizerService.GetJobStatistics()
	c.JSON(http.StatusOK, stats)
}

// GetOptimizationConfig returns the current optimization configuration
// GET /api/v1/optimize/config
func (h *OptimizationHandler) GetOptimizationConfig(c *gin.Context) {
	config := optimizer.DefaultOptimizationConfig()
	c.JSON(http.StatusOK, config)
}

// SetOptimizationConfig updates the optimization configuration
// PUT /api/v1/optimize/config
func (h *OptimizationHandler) SetOptimizationConfig(c *gin.Context) {
	var config optimizer.OptimizationConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{
			Error: "Invalid configuration",
			Details: map[string]string{
				"json": err.Error(),
			},
		})
		return
	}

	h.optimizerService.SetOptimizationConfig(config)
	c.JSON(http.StatusOK, gin.H{
		"status": "updated",
		"config": config,
	})
}

// RegisterRoutes registers optimization routes with the Gin router
func (h *OptimizationHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Optimization job management
	router.POST("/optimize/:drawId/start", h.StartOptimization)
	router.GET("/optimize/:jobId/status", h.GetOptimizationStatus)
	router.POST("/optimize/:jobId/cancel", h.CancelOptimization)
	router.GET("/optimize/:jobId/result", h.GetOptimizationResult)
	router.POST("/optimize/:jobId/apply", h.ApplyOptimizationResult)

	// Draw validation and scoring
	router.GET("/draws/:drawId/validate-constraints", h.ValidateDrawConstraints)
	router.GET("/draws/:drawId/score", h.ScoreDraw)

	// Job listing and statistics
	router.GET("/optimize/jobs", h.ListOptimizationJobs)
	router.GET("/optimize/statistics", h.GetJobStatistics)

	// Configuration
	router.GET("/optimize/config", h.GetOptimizationConfig)
	router.PUT("/optimize/config", h.SetOptimizationConfig)
}