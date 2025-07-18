package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/adampetrovic/nrl-scheduler/internal/api/handlers"
	"github.com/adampetrovic/nrl-scheduler/internal/api/middleware"
	"github.com/adampetrovic/nrl-scheduler/internal/api/websocket"
	"github.com/adampetrovic/nrl-scheduler/internal/core/optimizer"
	"github.com/adampetrovic/nrl-scheduler/internal/storage/sqlite"
)

type Server struct {
	router          *gin.Engine
	db              *sql.DB
	repos           *sqlite.Repositories
	validate        *validator.Validate
	optimizerService *optimizer.Service
	wsHub           *websocket.Hub
}

func NewServer(db *sql.DB) *Server {
	repos := sqlite.NewRepositories(db)
	validate := validator.New()
	
	// Create WebSocket hub
	wsHub := websocket.NewHub()
	
	// Create optimizer service
	optimizerService := optimizer.NewService(repos)

	server := &Server{
		router:          gin.New(),
		db:              db,
		repos:           repos,
		validate:        validate,
		optimizerService: optimizerService,
		wsHub:           wsHub,
	}

	// Set up WebSocket broadcasting for the optimizer service
	optimizerService.SetWebSocketHub(wsHub)

	// Start WebSocket hub
	go wsHub.Run()

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

func (s *Server) setupMiddleware() {
	s.router.Use(gin.Logger())
	s.router.Use(gin.Recovery())
	s.router.Use(func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Next()
	})
	s.router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
	s.router.Use(middleware.ErrorHandler())
	s.router.Use(middleware.RequestValidator(s.validate))
}

func (s *Server) setupRoutes() {
	api := s.router.Group("/api/v1")

	// Teams endpoints
	teamHandler := handlers.NewTeamHandler(s.repos.Teams())
	api.GET("/teams", teamHandler.GetTeams)
	api.POST("/teams", teamHandler.CreateTeam)
	api.GET("/teams/:id", teamHandler.GetTeam)
	api.PUT("/teams/:id", teamHandler.UpdateTeam)
	api.DELETE("/teams/:id", teamHandler.DeleteTeam)

	// Venues endpoints
	venueHandler := handlers.NewVenueHandler(s.repos.Venues())
	api.GET("/venues", venueHandler.GetVenues)
	api.POST("/venues", venueHandler.CreateVenue)
	api.GET("/venues/:id", venueHandler.GetVenue)
	api.PUT("/venues/:id", venueHandler.UpdateVenue)
	api.DELETE("/venues/:id", venueHandler.DeleteVenue)

	// Draws endpoints
	drawHandler := handlers.NewDrawHandler(s.repos.Draws(), s.repos.Teams(), s.repos.Matches(), s.wsHub)
	api.GET("/draws", drawHandler.GetDraws)
	api.POST("/draws", drawHandler.CreateDraw)
	api.GET("/draws/:id", drawHandler.GetDraw)
	api.PUT("/draws/:id", drawHandler.UpdateDraw)
	api.DELETE("/draws/:id", drawHandler.DeleteDraw)
	api.GET("/draws/:id/matches", drawHandler.GetDrawMatches)

	// Draw generation endpoints
	api.POST("/draws/:id/generate", drawHandler.GenerateDraw)
	api.POST("/draws/:id/validate-constraints", drawHandler.ValidateConstraints)

	// Optimization endpoints
	optimizationHandler := handlers.NewOptimizationHandler(s.optimizerService, s.wsHub)
	optimizationHandler.RegisterRoutes(api)

	// WebSocket endpoint
	s.router.GET("/ws", func(c *gin.Context) {
		s.wsHub.ServeWS(c.Writer, c.Request)
	})

	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test WebSocket endpoint
	s.router.POST("/test/websocket", func(c *gin.Context) {
		if s.wsHub != nil {
			s.wsHub.BroadcastMessage("test_message", gin.H{
				"message": "Test WebSocket broadcast",
				"timestamp": gin.H{},
				"clients": s.wsHub.GetClientCount(),
			})
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Test message broadcasted",
				"clients": s.wsHub.GetClientCount(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket hub not available"})
		}
	})
}

func (s *Server) Run(addr string) error {
	log.Printf("Starting server on %s", addr)
	return s.router.Run(addr)
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

func (s *Server) GetWebSocketHub() *websocket.Hub {
	return s.wsHub
}
