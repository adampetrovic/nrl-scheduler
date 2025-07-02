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

type VenueHandler struct {
	venueRepo storage.VenueRepository
}

func NewVenueHandler(venueRepo storage.VenueRepository) *VenueHandler {
	return &VenueHandler{
		venueRepo: venueRepo,
	}
}

func (h *VenueHandler) GetVenues(c *gin.Context) {
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

	venues, err := h.venueRepo.List(context.Background())
	if err != nil {
		middleware.InternalError(c, "Failed to retrieve venues")
		return
	}

	// Convert to response format
	venueResponses := make([]types.VenueResponse, len(venues))
	for i, venue := range venues {
		venueResponses[i] = types.VenueToResponse(venue)
	}

	// Simple pagination (in production, you'd do this in the database)
	total := len(venueResponses)
	start := (params.Page - 1) * params.PerPage
	end := start + params.PerPage
	
	if start >= total {
		venueResponses = []types.VenueResponse{}
	} else if end > total {
		venueResponses = venueResponses[start:]
	} else {
		venueResponses = venueResponses[start:end]
	}

	totalPages := (total + params.PerPage - 1) / params.PerPage

	response := types.PaginatedResponse{
		Data:       venueResponses,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

func (h *VenueHandler) GetVenue(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.BadRequest(c, "Invalid venue ID")
		return
	}

	venue, err := h.venueRepo.Get(context.Background(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			middleware.NotFound(c, "Venue not found")
			return
		}
		middleware.InternalError(c, "Failed to retrieve venue")
		return
	}

	response := types.VenueToResponse(venue)
	c.JSON(http.StatusOK, response)
}

func (h *VenueHandler) CreateVenue(c *gin.Context) {
	var req types.CreateVenueRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		c.Error(err)
		return
	}

	venue := &models.Venue{
		Name:      req.Name,
		City:      req.City,
		Capacity:  req.Capacity,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}

	if err := h.venueRepo.Create(context.Background(), venue); err != nil {
		middleware.InternalError(c, "Failed to create venue")
		return
	}

	response := types.VenueToResponse(venue)
	c.JSON(http.StatusCreated, response)
}

func (h *VenueHandler) UpdateVenue(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.BadRequest(c, "Invalid venue ID")
		return
	}

	var req types.UpdateVenueRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		c.Error(err)
		return
	}

	venue, err := h.venueRepo.Get(context.Background(), id)
	if err != nil {
		if err == storage.ErrNotFound {
			middleware.NotFound(c, "Venue not found")
			return
		}
		middleware.InternalError(c, "Failed to retrieve venue")
		return
	}

	// Update fields if provided
	if req.Name != nil {
		venue.Name = *req.Name
	}
	if req.City != nil {
		venue.City = *req.City
	}
	if req.Capacity != nil {
		venue.Capacity = *req.Capacity
	}
	if req.Latitude != nil {
		venue.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		venue.Longitude = *req.Longitude
	}

	if err := h.venueRepo.Update(context.Background(), venue); err != nil {
		middleware.InternalError(c, "Failed to update venue")
		return
	}

	response := types.VenueToResponse(venue)
	c.JSON(http.StatusOK, response)
}

func (h *VenueHandler) DeleteVenue(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.BadRequest(c, "Invalid venue ID")
		return
	}

	if err := h.venueRepo.Delete(context.Background(), id); err != nil {
		if err == storage.ErrNotFound {
			middleware.NotFound(c, "Venue not found")
			return
		}
		middleware.InternalError(c, "Failed to delete venue")
		return
	}

	c.JSON(http.StatusOK, types.SuccessResponse{
		Success: true,
		Message: "Venue deleted successfully",
	})
}