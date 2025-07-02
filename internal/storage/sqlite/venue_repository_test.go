package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

func TestVenueRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewVenueRepository(db.Conn())
	ctx := context.Background()

	venue := &models.Venue{
		Name:      "Suncorp Stadium",
		City:      "Brisbane",
		Capacity:  52500,
		Latitude:  -27.4649,
		Longitude: 153.0095,
	}

	err := repo.Create(ctx, venue)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if venue.ID == 0 {
		t.Error("Create() should set venue ID")
	}

	// Verify venue was created
	retrieved, err := repo.Get(ctx, venue.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.Name != venue.Name {
		t.Errorf("Name = %v, want %v", retrieved.Name, venue.Name)
	}
	if retrieved.City != venue.City {
		t.Errorf("City = %v, want %v", retrieved.City, venue.City)
	}
	if retrieved.Capacity != venue.Capacity {
		t.Errorf("Capacity = %v, want %v", retrieved.Capacity, venue.Capacity)
	}
}

func TestVenueRepository_Get(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewVenueRepository(db.Conn())
	ctx := context.Background()

	// Test non-existent venue
	_, err := repo.Get(ctx, 999)
	if err == nil {
		t.Error("Get() should return error for non-existent venue")
	}

	// Create and get venue
	venue := createTestVenue(t, repo)
	retrieved, err := repo.Get(ctx, venue.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.ID != venue.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, venue.ID)
	}
	if !retrieved.CreatedAt.IsZero() {
		// Should have timestamps
		if retrieved.CreatedAt.After(time.Now()) {
			t.Error("CreatedAt should not be in the future")
		}
	}
}

func TestVenueRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewVenueRepository(db.Conn())
	ctx := context.Background()

	// Test empty list
	venues, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(venues) != 0 {
		t.Errorf("List() should return empty list, got %d venues", len(venues))
	}

	// Create test venues
	_ = createTestVenue(t, repo)
	venue2 := &models.Venue{
		Name:      "ANZ Stadium",
		City:      "Sydney",
		Capacity:  84000,
		Latitude:  -33.8475,
		Longitude: 151.0636,
	}
	err = repo.Create(ctx, venue2)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// List venues
	venues, err = repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(venues) != 2 {
		t.Errorf("List() should return 2 venues, got %d", len(venues))
	}

	// Should be ordered by name (ANZ before Suncorp)
	if venues[0].Name != "ANZ Stadium" {
		t.Errorf("First venue name = %v, want ANZ Stadium", venues[0].Name)
	}
}

func TestVenueRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewVenueRepository(db.Conn())
	ctx := context.Background()

	// Test updating non-existent venue
	nonExistent := &models.Venue{ID: 999, Name: "Test"}
	err := repo.Update(ctx, nonExistent)
	if err == nil {
		t.Error("Update() should return error for non-existent venue")
	}

	// Create and update venue
	venue := createTestVenue(t, repo)
	originalName := venue.Name

	venue.Name = "Updated Stadium"
	venue.Capacity = 60000

	err = repo.Update(ctx, venue)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update
	updated, err := repo.Get(ctx, venue.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if updated.Name == originalName {
		t.Error("Update() should change venue name")
	}
	if updated.Name != "Updated Stadium" {
		t.Errorf("Name = %v, want Updated Stadium", updated.Name)
	}
	if updated.Capacity != 60000 {
		t.Errorf("Capacity = %v, want 60000", updated.Capacity)
	}
}

func TestVenueRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewVenueRepository(db.Conn())
	ctx := context.Background()

	// Test deleting non-existent venue
	err := repo.Delete(ctx, 999)
	if err == nil {
		t.Error("Delete() should return error for non-existent venue")
	}

	// Create and delete venue
	venue := createTestVenue(t, repo)

	err = repo.Delete(ctx, venue.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = repo.Get(ctx, venue.ID)
	if err == nil {
		t.Error("Get() should return error for deleted venue")
	}
}

func createTestVenue(t *testing.T, repo *VenueRepository) *models.Venue {
	venue := &models.Venue{
		Name:      "Suncorp Stadium",
		City:      "Brisbane",
		Capacity:  52500,
		Latitude:  -27.4649,
		Longitude: 153.0095,
	}

	err := repo.Create(context.Background(), venue)
	if err != nil {
		t.Fatalf("Failed to create test venue: %v", err)
	}

	return venue
}