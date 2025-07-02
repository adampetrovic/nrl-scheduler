package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/adampetrovic/nrl-scheduler/internal/api/middleware"
	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
	"github.com/adampetrovic/nrl-scheduler/internal/storage"
	"github.com/adampetrovic/nrl-scheduler/pkg/types"
)

type TeamHandler struct {
	teamRepo storage.TeamRepository
}

func NewTeamHandler(teamRepo storage.TeamRepository) *TeamHandler {
	return &TeamHandler{
		teamRepo: teamRepo,
	}
}

func (h *TeamHandler) GetTeams(c *gin.Context) {
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

	teams, err := h.teamRepo.List(context.Background())
	if err != nil {
		middleware.InternalError(c, "Failed to retrieve teams")
		return
	}

	// Convert to response format
	teamResponses := make([]types.TeamResponse, len(teams))
	for i, team := range teams {
		teamResponses[i] = types.TeamToResponse(team, nil)
	}

	// Simple pagination (in production, you'd do this in the database)
	total := len(teamResponses)
	start := (params.Page - 1) * params.PerPage
	end := start + params.PerPage
	
	if start >= total {
		teamResponses = []types.TeamResponse{}
	} else if end > total {
		teamResponses = teamResponses[start:]
	} else {
		teamResponses = teamResponses[start:end]
	}

	totalPages := (total + params.PerPage - 1) / params.PerPage

	response := types.PaginatedResponse{
		Data:       teamResponses,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.BadRequest(c, "Invalid team ID")
		return
	}

	team, err := h.teamRepo.Get(context.Background(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			middleware.NotFound(c, "Team not found")
			return
		}
		middleware.InternalError(c, "Failed to retrieve team")
		return
	}

	response := types.TeamToResponse(team, nil)
	c.JSON(http.StatusOK, response)
}

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req types.CreateTeamRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		c.Error(err)
		return
	}

	team := &models.Team{
		Name:      req.Name,
		ShortName: req.ShortName,
		City:      req.City,
		VenueID:   req.VenueID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}

	if err := h.teamRepo.Create(context.Background(), team); err != nil {
		middleware.InternalError(c, "Failed to create team")
		return
	}

	response := types.TeamToResponse(team, nil)
	c.JSON(http.StatusCreated, response)
}

func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.BadRequest(c, "Invalid team ID")
		return
	}

	var req types.UpdateTeamRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		c.Error(err)
		return
	}

	team, err := h.teamRepo.Get(context.Background(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			middleware.NotFound(c, "Team not found")
			return
		}
		middleware.InternalError(c, "Failed to retrieve team")
		return
	}

	// Update fields if provided
	if req.Name != nil {
		team.Name = *req.Name
	}
	if req.ShortName != nil {
		team.ShortName = *req.ShortName
	}
	if req.City != nil {
		team.City = *req.City
	}
	if req.VenueID != nil {
		team.VenueID = req.VenueID
	}
	if req.Latitude != nil {
		team.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		team.Longitude = *req.Longitude
	}

	if err := h.teamRepo.Update(context.Background(), team); err != nil {
		middleware.InternalError(c, "Failed to update team")
		return
	}

	response := types.TeamToResponse(team, nil)
	c.JSON(http.StatusOK, response)
}

func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.BadRequest(c, "Invalid team ID")
		return
	}

	if err := h.teamRepo.Delete(context.Background(), id); err != nil {
		if err == storage.ErrNotFound {
			middleware.NotFound(c, "Team not found")
			return
		}
		middleware.InternalError(c, "Failed to delete team")
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Team deleted successfully",
	})
}