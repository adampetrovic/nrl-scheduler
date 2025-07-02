package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// DrawRepository implements storage.DrawRepository using SQLite
type DrawRepository struct {
	db DBExecutor
}

// NewDrawRepository creates a new draw repository
func NewDrawRepository(db DBExecutor) *DrawRepository {
	return &DrawRepository{db: db}
}

// Create inserts a new draw
func (r *DrawRepository) Create(ctx context.Context, draw *models.Draw) error {
	query := `
		INSERT INTO draws (name, season_year, rounds, status, constraint_config)
		VALUES (?, ?, ?, ?, ?)
	`
	
	result, err := r.db.ExecContext(ctx, query,
		draw.Name, draw.SeasonYear, draw.Rounds, draw.Status, draw.ConstraintConfig)
	if err != nil {
		return fmt.Errorf("creating draw: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}

	draw.ID = int(id)
	return nil
}

// Get retrieves a draw by ID
func (r *DrawRepository) Get(ctx context.Context, id int) (*models.Draw, error) {
	query := `
		SELECT id, name, season_year, rounds, status, constraint_config, created_at, updated_at
		FROM draws
		WHERE id = ?
	`

	draw := &models.Draw{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&draw.ID, &draw.Name, &draw.SeasonYear, &draw.Rounds,
		&draw.Status, &draw.ConstraintConfig, &draw.CreatedAt, &draw.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("draw not found")
	}
	if err != nil {
		return nil, fmt.Errorf("getting draw: %w", err)
	}

	return draw, nil
}

// GetWithMatches retrieves a draw with all its matches
func (r *DrawRepository) GetWithMatches(ctx context.Context, id int) (*models.Draw, error) {
	// First get the draw
	draw, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Then get all matches for this draw
	query := `
		SELECT 
			m.id, m.draw_id, m.round, m.home_team_id, m.away_team_id, 
			m.venue_id, m.match_date, m.match_time, m.is_prime_time,
			m.created_at, m.updated_at
		FROM matches m
		WHERE m.draw_id = ?
		ORDER BY m.round, m.id
	`

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("getting matches for draw: %w", err)
	}
	defer rows.Close()

	var matches []*models.Match
	for rows.Next() {
		match := &models.Match{}
		var matchDate, matchTime sql.NullTime

		err := rows.Scan(
			&match.ID, &match.DrawID, &match.Round,
			&match.HomeTeamID, &match.AwayTeamID, &match.VenueID,
			&matchDate, &matchTime, &match.IsPrimeTime,
			&match.CreatedAt, &match.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning match: %w", err)
		}

		if matchDate.Valid {
			match.MatchDate = &matchDate.Time
		}
		if matchTime.Valid {
			match.MatchTime = &matchTime.Time
		}

		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating matches: %w", err)
	}

	draw.Matches = matches
	return draw, nil
}

// List retrieves all draws
func (r *DrawRepository) List(ctx context.Context) ([]*models.Draw, error) {
	query := `
		SELECT id, name, season_year, rounds, status, constraint_config, created_at, updated_at
		FROM draws
		ORDER BY season_year DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing draws: %w", err)
	}
	defer rows.Close()

	var draws []*models.Draw
	for rows.Next() {
		draw := &models.Draw{}
		err := rows.Scan(
			&draw.ID, &draw.Name, &draw.SeasonYear, &draw.Rounds,
			&draw.Status, &draw.ConstraintConfig, &draw.CreatedAt, &draw.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning draw: %w", err)
		}
		draws = append(draws, draw)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating draws: %w", err)
	}

	return draws, nil
}

// Update modifies an existing draw
func (r *DrawRepository) Update(ctx context.Context, draw *models.Draw) error {
	query := `
		UPDATE draws
		SET name = ?, season_year = ?, rounds = ?, status = ?, constraint_config = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		draw.Name, draw.SeasonYear, draw.Rounds, draw.Status, draw.ConstraintConfig, draw.ID)
	if err != nil {
		return fmt.Errorf("updating draw: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("draw not found")
	}

	return nil
}

// Delete removes a draw (matches are cascade deleted)
func (r *DrawRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM draws WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting draw: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("draw not found")
	}

	return nil
}