package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adampetrovic/nrl-scheduler/internal/api"
	"github.com/adampetrovic/nrl-scheduler/pkg/types"
	
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Use in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	
	// Create basic schema for testing
	schema := `
	CREATE TABLE IF NOT EXISTS venues (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		city TEXT NOT NULL,
		capacity INTEGER NOT NULL,
		latitude REAL NOT NULL DEFAULT 0,
		longitude REAL NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS teams (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		short_name TEXT NOT NULL,
		city TEXT NOT NULL,
		venue_id INTEGER,
		latitude REAL NOT NULL DEFAULT 0,
		longitude REAL NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (venue_id) REFERENCES venues(id)
	);

	CREATE TABLE IF NOT EXISTS draws (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		season_year INTEGER NOT NULL,
		rounds INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'draft',
		constraint_config TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	
	_, err = db.Exec(schema)
	require.NoError(t, err)
	
	return db
}

func setupTestServer(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	server := api.NewServer(db)
	return server.GetRouter()
}

func TestHealthCheck(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	router := setupTestServer(db)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestVenueCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	router := setupTestServer(db)
	
	// Test Create Venue
	createReq := types.CreateVenueRequest{
		Name:      "Test Stadium",
		City:      "Test City",
		Capacity:  50000,
		Latitude:  -33.8688,
		Longitude: 151.2093,
	}
	
	body, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/venues", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var createResp types.VenueResponse
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	assert.NoError(t, err)
	assert.Equal(t, "Test Stadium", createResp.Name)
	assert.Equal(t, "Test City", createResp.City)
	assert.Equal(t, 50000, createResp.Capacity)
	
	// Test Get Venue
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/venues/1", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Test List Venues
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/venues", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var listResp types.PaginatedResponse
	err = json.Unmarshal(w.Body.Bytes(), &listResp)
	assert.NoError(t, err)
	assert.Equal(t, 1, listResp.Total)
}

func TestTeamCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	router := setupTestServer(db)
	
	// First create a venue
	venueReq := types.CreateVenueRequest{
		Name:      "Team Stadium",
		City:      "Team City",
		Capacity:  40000,
		Latitude:  -33.8688,
		Longitude: 151.2093,
	}
	
	body, _ := json.Marshal(venueReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/venues", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	var venueResp types.VenueResponse
	json.Unmarshal(w.Body.Bytes(), &venueResp)
	venueID := venueResp.ID
	
	// Test Create Team
	createReq := types.CreateTeamRequest{
		Name:      "Test Team",
		ShortName: "TST",
		City:      "Test City",
		VenueID:   &venueID,
		Latitude:  -33.8688,
		Longitude: 151.2093,
	}
	
	body, _ = json.Marshal(createReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/teams", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var createResp types.TeamResponse
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	assert.NoError(t, err)
	assert.Equal(t, "Test Team", createResp.Name)
	assert.Equal(t, "TST", createResp.ShortName)
	assert.Equal(t, "Test City", createResp.City)
	
	// Test List Teams
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/teams", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var listResp types.PaginatedResponse
	err = json.Unmarshal(w.Body.Bytes(), &listResp)
	assert.NoError(t, err)
	assert.Equal(t, 1, listResp.Total)
}

func TestDrawCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	router := setupTestServer(db)
	
	// Test Create Draw
	createReq := types.CreateDrawRequest{
		Name:       "Test Draw",
		SeasonYear: 2024,
		Rounds:     26,
	}
	
	body, _ := json.Marshal(createReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/draws", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var createResp types.DrawResponse
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	assert.NoError(t, err)
	assert.Equal(t, "Test Draw", createResp.Name)
	assert.Equal(t, 2024, createResp.SeasonYear)
	assert.Equal(t, 26, createResp.Rounds)
	assert.Equal(t, "draft", createResp.Status)
	
	// Test List Draws
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/draws", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var listResp types.PaginatedResponse
	err = json.Unmarshal(w.Body.Bytes(), &listResp)
	assert.NoError(t, err)
	assert.Equal(t, 1, listResp.Total)
}

func TestValidationErrors(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	
	router := setupTestServer(db)
	
	// Test invalid venue creation
	invalidReq := types.CreateVenueRequest{
		Name:     "", // Empty name should fail validation
		City:     "Test City",
		Capacity: 50000,
	}
	
	body, _ := json.Marshal(invalidReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/venues", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var errorResp types.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Contains(t, errorResp.Error, "Validation failed")
}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}