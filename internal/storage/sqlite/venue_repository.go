package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// VenueRepository implements storage.VenueRepository using SQLite
type VenueRepository struct {
	db DBExecutor
}

// NewVenueRepository creates a new venue repository
func NewVenueRepository(db DBExecutor) *VenueRepository {
	return &VenueRepository{db: db}
}

// Create inserts a new venue
func (r *VenueRepository) Create(ctx context.Context, venue *models.Venue) error {
	query := `
		INSERT INTO venues (name, city, capacity, latitude, longitude)
		VALUES (?, ?, ?, ?, ?)
	`
	
	result, err := r.db.ExecContext(ctx, query,
		venue.Name, venue.City, venue.Capacity, venue.Latitude, venue.Longitude)
	if err != nil {
		return fmt.Errorf("creating venue: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}

	venue.ID = int(id)
	return nil
}

// Get retrieves a venue by ID
func (r *VenueRepository) Get(ctx context.Context, id int) (*models.Venue, error) {
	query := `
		SELECT id, name, city, capacity, latitude, longitude, created_at, updated_at
		FROM venues
		WHERE id = ?
	`

	venue := &models.Venue{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&venue.ID, &venue.Name, &venue.City, &venue.Capacity,
		&venue.Latitude, &venue.Longitude, &venue.CreatedAt, &venue.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("venue not found")
	}
	if err != nil {
		return nil, fmt.Errorf("getting venue: %w", err)
	}

	return venue, nil
}

// List retrieves all venues
func (r *VenueRepository) List(ctx context.Context) ([]*models.Venue, error) {
	query := `
		SELECT id, name, city, capacity, latitude, longitude, created_at, updated_at
		FROM venues
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing venues: %w", err)
	}
	defer rows.Close()

	var venues []*models.Venue
	for rows.Next() {
		venue := &models.Venue{}
		err := rows.Scan(
			&venue.ID, &venue.Name, &venue.City, &venue.Capacity,
			&venue.Latitude, &venue.Longitude, &venue.CreatedAt, &venue.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning venue: %w", err)
		}
		venues = append(venues, venue)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating venues: %w", err)
	}

	return venues, nil
}

// Update modifies an existing venue
func (r *VenueRepository) Update(ctx context.Context, venue *models.Venue) error {
	query := `
		UPDATE venues
		SET name = ?, city = ?, capacity = ?, latitude = ?, longitude = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		venue.Name, venue.City, venue.Capacity, venue.Latitude, venue.Longitude, venue.ID)
	if err != nil {
		return fmt.Errorf("updating venue: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("venue not found")
	}

	return nil
}

// Delete removes a venue
func (r *VenueRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM venues WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting venue: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("venue not found")
	}

	return nil
}