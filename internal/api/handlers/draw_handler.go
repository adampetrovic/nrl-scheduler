package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/adampetrovic/nrl-scheduler/internal/api/middleware"
	"github.com/adampetrovic/nrl-scheduler/internal/api/websocket"
	"github.com/adampetrovic/nrl-scheduler/internal/core/constraints"
	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
	"github.com/adampetrovic/nrl-scheduler/internal/storage"
	"github.com/adampetrovic/nrl-scheduler/pkg/types"
)

type DrawHandler struct {
	drawRepo  storage.DrawRepository
	teamRepo  storage.TeamRepository
	matchRepo storage.MatchRepository
	wsHub     *websocket.Hub
}

func NewDrawHandler(drawRepo storage.DrawRepository, teamRepo storage.TeamRepository, matchRepo storage.MatchRepository, wsHub *websocket.Hub) *DrawHandler {
	return &DrawHandler{
		drawRepo:  drawRepo,
		teamRepo:  teamRepo,
		matchRepo: matchRepo,
		wsHub:     wsHub,
	}
}

func (h *DrawHandler) GetDraws(c *gin.Context) {
	var params types.ListQueryParams
	if err := middleware.BindQueryAndValidate(c, &params); err != nil {
		middleware.BadRequest(c, "Invalid query parameters")
		return
	}

	// Set defaults
	if params.Page == 0 {
		params.Page = 1
	}
	if params.PerPage == 0 {
		params.PerPage = 20
	}

	draws, err := h.drawRepo.List(context.Background())
	if err != nil {
		log.Printf("Error retrieving draws: %v", err)
		middleware.InternalError(c, "Failed to retrieve draws")
		return
	}

	// Convert to response format
	drawResponses := make([]types.DrawResponse, len(draws))
	for i, draw := range draws {
		log.Printf("Converting draw %d: %+v", draw.ID, draw)
		drawResponses[i] = types.DrawToResponse(draw)
	}

	// Simple pagination
	total := len(drawResponses)
	start := (params.Page - 1) * params.PerPage
	end := start + params.PerPage
	
	if start >= total {
		drawResponses = []types.DrawResponse{}
	} else if end > total {
		drawResponses = drawResponses[start:]
	} else {
		drawResponses = drawResponses[start:end]
	}

	totalPages := (total + params.PerPage - 1) / params.PerPage

	response := types.PaginatedResponse{
		Data:       drawResponses,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

func (h *DrawHandler) GetDraw(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.BadRequest(c, "Invalid draw ID")
		return
	}

	drawModel, err := h.drawRepo.Get(context.Background(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			middleware.NotFound(c, "Draw not found")
			return
		}
		middleware.InternalError(c, "Failed to retrieve draw")
		return
	}

	response := types.DrawToResponse(drawModel)
	c.JSON(http.StatusOK, response)
}

func (h *DrawHandler) CreateDraw(c *gin.Context) {
	var req types.CreateDrawRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		c.Error(err)
		return
	}

	// Convert constraint config to JSON if provided
	var constraintConfigJSON json.RawMessage
	if req.ConstraintConfig != nil {
		var err error
		constraintConfigJSON, err = json.Marshal(req.ConstraintConfig)
		if err != nil {
			middleware.BadRequest(c, "Invalid constraint configuration")
			return
		}
	}

	drawModel := &models.Draw{
		Name:             req.Name,
		SeasonYear:       req.SeasonYear,
		Rounds:           req.Rounds,
		Status:           models.DrawStatusDraft,
		ConstraintConfig: constraintConfigJSON,
	}

	if err := h.drawRepo.Create(context.Background(), drawModel); err != nil {
		middleware.InternalError(c, "Failed to create draw")
		return
	}

	// Broadcast draw creation event
	if h.wsHub != nil {
		h.wsHub.BroadcastMessage(websocket.DrawCreated, websocket.DrawEventData{
			Draw:      drawModel,
			Timestamp: time.Now(),
		})
	}

	response := types.DrawToResponse(drawModel)
	c.JSON(http.StatusCreated, response)
}

func (h *DrawHandler) UpdateDraw(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.BadRequest(c, "Invalid draw ID")
		return
	}

	var req types.UpdateDrawRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		c.Error(err)
		return
	}

	drawModel, err := h.drawRepo.Get(context.Background(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			middleware.NotFound(c, "Draw not found")
			return
		}
		middleware.InternalError(c, "Failed to retrieve draw")
		return
	}

	// Update fields if provided
	if req.Name != nil {
		drawModel.Name = *req.Name
	}
	if req.SeasonYear != nil {
		drawModel.SeasonYear = *req.SeasonYear
	}
	if req.Rounds != nil {
		drawModel.Rounds = *req.Rounds
	}
	if req.ConstraintConfig != nil {
		var err error
		drawModel.ConstraintConfig, err = json.Marshal(req.ConstraintConfig)
		if err != nil {
			middleware.BadRequest(c, "Invalid constraint configuration")
			return
		}
	}

	if err := h.drawRepo.Update(context.Background(), drawModel); err != nil {
		middleware.InternalError(c, "Failed to update draw")
		return
	}

	// Broadcast draw update event
	if h.wsHub != nil {
		h.wsHub.BroadcastMessage(websocket.DrawUpdated, websocket.DrawEventData{
			Draw:      drawModel,
			Timestamp: time.Now(),
		})
	}

	response := types.DrawToResponse(drawModel)
	c.JSON(http.StatusOK, response)
}

func (h *DrawHandler) DeleteDraw(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.BadRequest(c, "Invalid draw ID")
		return
	}

	if err := h.drawRepo.Delete(context.Background(), id); err != nil {
		if err == storage.ErrNotFound {
			middleware.NotFound(c, "Draw not found")
			return
		}
		middleware.InternalError(c, "Failed to delete draw")
		return
	}

	// Broadcast draw deletion event
	if h.wsHub != nil {
		h.wsHub.BroadcastMessage(websocket.DrawDeleted, websocket.DrawEventData{
			Draw:      &models.Draw{ID: id}, // Just ID for deletion
			Timestamp: time.Now(),
		})
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Draw deleted successfully",
	})
}

func (h *DrawHandler) GetDrawMatches(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.BadRequest(c, "Invalid draw ID")
		return
	}

	drawModel, err := h.drawRepo.Get(context.Background(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			middleware.NotFound(c, "Draw not found")
			return
		}
		middleware.InternalError(c, "Failed to retrieve draw")
		return
	}

	// For now, return the matches from the draw model
	// In a full implementation, you might fetch from a match repository
	matchResponses := make([]types.MatchResponse, len(drawModel.Matches))
	for i, match := range drawModel.Matches {
		var homeTeam, awayTeam *models.Team
		var venue *models.Venue
		
		if match.HomeTeamID != nil {
			homeTeam, _ = h.teamRepo.Get(context.Background(), *match.HomeTeamID)
		}
		if match.AwayTeamID != nil {
			awayTeam, _ = h.teamRepo.Get(context.Background(), *match.AwayTeamID)
		}
		// Placeholder venue - would fetch from venue repo if VenueID exists
		
		matchResponses[i] = types.MatchToResponse(match, homeTeam, awayTeam, venue)
	}

	c.JSON(http.StatusOK, matchResponses)
}

func (h *DrawHandler) GenerateDraw(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.BadRequest(c, "Invalid draw ID")
		return
	}

	var req types.GenerateDrawRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		c.Error(err)
		return
	}

	drawModel, err := h.drawRepo.Get(context.Background(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			middleware.NotFound(c, "Draw not found")
			return
		}
		middleware.InternalError(c, "Failed to retrieve draw")
		return
	}

	// TODO: Implement actual draw generation
	// For now, just change status to optimizing
	drawModel.Status = models.DrawStatusOptimizing
	
	if err := h.drawRepo.Update(context.Background(), drawModel); err != nil {
		middleware.InternalError(c, "Failed to update draw status")
		return
	}

	response := types.GenerateDrawResponse{
		Success:        true,
		MatchCount:     0,
		Violations:     []types.ConstraintViolation{},
		Message:        "Draw generation started (placeholder implementation)",
		GeneratedAt:    time.Now(),
		GenerationTime: time.Millisecond,
	}

	c.JSON(http.StatusOK, response)
}

func (h *DrawHandler) ValidateConstraints(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.BadRequest(c, "Invalid draw ID")
		return
	}

	var req types.ValidateConstraintsRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		c.Error(err)
		return
	}

	drawModel, err := h.drawRepo.Get(context.Background(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			middleware.NotFound(c, "Draw not found")
			return
		}
		middleware.InternalError(c, "Failed to retrieve draw")
		return
	}

	if drawModel.Status == models.DrawStatusDraft {
		middleware.BadRequest(c, "Draw has not been generated yet")
		return
	}

	// TODO: Implement actual constraint validation
	// For now, return a simple placeholder response
	response := types.ValidateConstraintsResponse{
		IsValid:    true,
		Violations: []types.ConstraintViolation{},
		Score:      1.0,
	}

	// Broadcast constraint validation event
	if h.wsHub != nil {
		h.wsHub.BroadcastMessage(websocket.ConstraintsValidated, websocket.ConstraintsValidatedData{
			DrawID:      id,
			IsValid:     response.IsValid,
			Violations:  []constraints.ConstraintViolation{}, // Convert from types.ConstraintViolation if needed
			Score:       response.Score,
			ValidatedAt: time.Now(),
		})
	}

	c.JSON(http.StatusOK, response)
}