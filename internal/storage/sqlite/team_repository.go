package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// TeamRepository implements storage.TeamRepository using SQLite
type TeamRepository struct {
	db DBExecutor
}

// NewTeamRepository creates a new team repository
func NewTeamRepository(db DBExecutor) *TeamRepository {
	return &TeamRepository{db: db}
}

// Create inserts a new team
func (r *TeamRepository) Create(ctx context.Context, team *models.Team) error {
	query := `
		INSERT INTO teams (name, short_name, city, venue_id, latitude, longitude)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.ExecContext(ctx, query,
		team.Name, team.ShortName, team.City, team.VenueID, team.Latitude, team.Longitude)
	if err != nil {
		return fmt.Errorf("creating team: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}

	team.ID = int(id)
	return nil
}

// Get retrieves a team by ID
func (r *TeamRepository) Get(ctx context.Context, id int) (*models.Team, error) {
	query := `
		SELECT id, name, short_name, city, venue_id, latitude, longitude, created_at, updated_at
		FROM teams
		WHERE id = ?
	`

	team := &models.Team{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&team.ID, &team.Name, &team.ShortName, &team.City, &team.VenueID,
		&team.Latitude, &team.Longitude, &team.CreatedAt, &team.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("team not found")
	}
	if err != nil {
		return nil, fmt.Errorf("getting team: %w", err)
	}

	return team, nil
}

// GetWithVenue retrieves a team with its venue
func (r *TeamRepository) GetWithVenue(ctx context.Context, id int) (*models.Team, error) {
	query := `
		SELECT 
			t.id, t.name, t.short_name, t.city, t.venue_id, t.latitude, t.longitude, 
			t.created_at, t.updated_at,
			v.id, v.name, v.city, v.capacity, v.latitude, v.longitude
		FROM teams t
		LEFT JOIN venues v ON t.venue_id = v.id
		WHERE t.id = ?
	`

	team := &models.Team{}
	var venue models.Venue
	var venueID sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&team.ID, &team.Name, &team.ShortName, &team.City, &venueID,
		&team.Latitude, &team.Longitude, &team.CreatedAt, &team.UpdatedAt,
		&venue.ID, &venue.Name, &venue.City, &venue.Capacity,
		&venue.Latitude, &venue.Longitude,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("team not found")
	}
	if err != nil {
		return nil, fmt.Errorf("getting team with venue: %w", err)
	}

	if venueID.Valid {
		team.VenueID = &[]int{int(venueID.Int64)}[0]
		team.Venue = &venue
	}

	return team, nil
}

// List retrieves all teams
func (r *TeamRepository) List(ctx context.Context) ([]*models.Team, error) {
	query := `
		SELECT id, name, short_name, city, venue_id, latitude, longitude, created_at, updated_at
		FROM teams
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing teams: %w", err)
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		team := &models.Team{}
		err := rows.Scan(
			&team.ID, &team.Name, &team.ShortName, &team.City, &team.VenueID,
			&team.Latitude, &team.Longitude, &team.CreatedAt, &team.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning team: %w", err)
		}
		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating teams: %w", err)
	}

	return teams, nil
}

// ListWithVenues retrieves all teams with their venues
func (r *TeamRepository) ListWithVenues(ctx context.Context) ([]*models.Team, error) {
	query := `
		SELECT 
			t.id, t.name, t.short_name, t.city, t.venue_id, t.latitude, t.longitude, 
			t.created_at, t.updated_at,
			v.id, v.name, v.city, v.capacity, v.latitude, v.longitude
		FROM teams t
		LEFT JOIN venues v ON t.venue_id = v.id
		ORDER BY t.name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing teams with venues: %w", err)
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		team := &models.Team{}
		var venue models.Venue
		var venueID sql.NullInt64

		err := rows.Scan(
			&team.ID, &team.Name, &team.ShortName, &team.City, &venueID,
			&team.Latitude, &team.Longitude, &team.CreatedAt, &team.UpdatedAt,
			&venue.ID, &venue.Name, &venue.City, &venue.Capacity,
			&venue.Latitude, &venue.Longitude,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning team with venue: %w", err)
		}

		if venueID.Valid {
			team.VenueID = &[]int{int(venueID.Int64)}[0]
			team.Venue = &venue
		}

		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating teams: %w", err)
	}

	return teams, nil
}

// Update modifies an existing team
func (r *TeamRepository) Update(ctx context.Context, team *models.Team) error {
	query := `
		UPDATE teams
		SET name = ?, short_name = ?, city = ?, venue_id = ?, latitude = ?, longitude = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		team.Name, team.ShortName, team.City, team.VenueID, 
		team.Latitude, team.Longitude, team.ID)
	if err != nil {
		return fmt.Errorf("updating team: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("team not found")
	}

	return nil
}

// Delete removes a team
func (r *TeamRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM teams WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting team: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("team not found")
	}

	return nil
}