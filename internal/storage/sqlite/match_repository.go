package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// MatchRepository implements storage.MatchRepository using SQLite
type MatchRepository struct {
	db    DBExecutor
	sqlDB *sql.DB // Keep reference for transaction operations
}

// NewMatchRepository creates a new match repository
func NewMatchRepository(db DBExecutor) *MatchRepository {
	var sqlDB *sql.DB
	if sdb, ok := db.(*sql.DB); ok {
		sqlDB = sdb
	}
	return &MatchRepository{db: db, sqlDB: sqlDB}
}

// Create inserts a new match
func (r *MatchRepository) Create(ctx context.Context, match *models.Match) error {
	query := `
		INSERT INTO matches (draw_id, round, home_team_id, away_team_id, venue_id, 
			match_date, match_time, is_prime_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.ExecContext(ctx, query,
		match.DrawID, match.Round, match.HomeTeamID, match.AwayTeamID,
		match.VenueID, match.MatchDate, match.MatchTime, match.IsPrimeTime)
	if err != nil {
		return fmt.Errorf("creating match: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %w", err)
	}

	match.ID = int(id)
	return nil
}

// CreateBatch inserts multiple matches in a single transaction
func (r *MatchRepository) CreateBatch(ctx context.Context, matches []*models.Match) error {
	if len(matches) == 0 {
		return nil
	}

	// If we don't have a sql.DB reference, fall back to individual creates
	if r.sqlDB == nil {
		for _, match := range matches {
			if err := r.Create(ctx, match); err != nil {
				return err
			}
		}
		return nil
	}

	tx, err := r.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO matches (draw_id, round, home_team_id, away_team_id, venue_id, 
			match_date, match_time, is_prime_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, match := range matches {
		result, err := stmt.ExecContext(ctx,
			match.DrawID, match.Round, match.HomeTeamID, match.AwayTeamID,
			match.VenueID, match.MatchDate, match.MatchTime, match.IsPrimeTime)
		if err != nil {
			return fmt.Errorf("creating match: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("getting last insert id: %w", err)
		}

		match.ID = int(id)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// Get retrieves a match by ID
func (r *MatchRepository) Get(ctx context.Context, id int) (*models.Match, error) {
	query := `
		SELECT id, draw_id, round, home_team_id, away_team_id, venue_id,
			match_date, match_time, is_prime_time, created_at, updated_at
		FROM matches
		WHERE id = ?
	`

	match := &models.Match{}
	var matchDate, matchTime sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&match.ID, &match.DrawID, &match.Round,
		&match.HomeTeamID, &match.AwayTeamID, &match.VenueID,
		&matchDate, &matchTime, &match.IsPrimeTime,
		&match.CreatedAt, &match.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("match not found")
	}
	if err != nil {
		return nil, fmt.Errorf("getting match: %w", err)
	}

	if matchDate.Valid {
		match.MatchDate = &matchDate.Time
	}
	if matchTime.Valid {
		match.MatchTime = &matchTime.Time
	}

	return match, nil
}

// GetWithRelations retrieves a match with teams and venue
func (r *MatchRepository) GetWithRelations(ctx context.Context, id int) (*models.Match, error) {
	query := `
		SELECT 
			m.id, m.draw_id, m.round, m.home_team_id, m.away_team_id, m.venue_id,
			m.match_date, m.match_time, m.is_prime_time, m.created_at, m.updated_at,
			ht.id, ht.name, ht.short_name, ht.city,
			at.id, at.name, at.short_name, at.city,
			v.id, v.name, v.city, v.capacity
		FROM matches m
		LEFT JOIN teams ht ON m.home_team_id = ht.id
		LEFT JOIN teams at ON m.away_team_id = at.id
		LEFT JOIN venues v ON m.venue_id = v.id
		WHERE m.id = ?
	`

	match := &models.Match{}
	var matchDate, matchTime sql.NullTime
	var homeTeam, awayTeam models.Team
	var venue models.Venue
	var homeTeamID, awayTeamID, venueID sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&match.ID, &match.DrawID, &match.Round,
		&homeTeamID, &awayTeamID, &venueID,
		&matchDate, &matchTime, &match.IsPrimeTime,
		&match.CreatedAt, &match.UpdatedAt,
		&homeTeam.ID, &homeTeam.Name, &homeTeam.ShortName, &homeTeam.City,
		&awayTeam.ID, &awayTeam.Name, &awayTeam.ShortName, &awayTeam.City,
		&venue.ID, &venue.Name, &venue.City, &venue.Capacity,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("match not found")
	}
	if err != nil {
		return nil, fmt.Errorf("getting match with relations: %w", err)
	}

	if matchDate.Valid {
		match.MatchDate = &matchDate.Time
	}
	if matchTime.Valid {
		match.MatchTime = &matchTime.Time
	}
	if homeTeamID.Valid {
		match.HomeTeamID = &[]int{int(homeTeamID.Int64)}[0]
		match.HomeTeam = &homeTeam
	}
	if awayTeamID.Valid {
		match.AwayTeamID = &[]int{int(awayTeamID.Int64)}[0]
		match.AwayTeam = &awayTeam
	}
	if venueID.Valid {
		match.VenueID = &[]int{int(venueID.Int64)}[0]
		match.Venue = &venue
	}

	return match, nil
}

// ListByDraw retrieves all matches for a draw
func (r *MatchRepository) ListByDraw(ctx context.Context, drawID int) ([]*models.Match, error) {
	query := `
		SELECT id, draw_id, round, home_team_id, away_team_id, venue_id,
			match_date, match_time, is_prime_time, created_at, updated_at
		FROM matches
		WHERE draw_id = ?
		ORDER BY round, id
	`

	return r.listMatches(ctx, query, drawID)
}

// ListByDrawWithRelations retrieves all matches for a draw with relations
func (r *MatchRepository) ListByDrawWithRelations(ctx context.Context, drawID int) ([]*models.Match, error) {
	query := `
		SELECT 
			m.id, m.draw_id, m.round, m.home_team_id, m.away_team_id, m.venue_id,
			m.match_date, m.match_time, m.is_prime_time, m.created_at, m.updated_at,
			ht.id, ht.name, ht.short_name, ht.city,
			at.id, at.name, at.short_name, at.city,
			v.id, v.name, v.city, v.capacity
		FROM matches m
		LEFT JOIN teams ht ON m.home_team_id = ht.id
		LEFT JOIN teams at ON m.away_team_id = at.id
		LEFT JOIN venues v ON m.venue_id = v.id
		WHERE m.draw_id = ?
		ORDER BY m.round, m.id
	`

	return r.listMatchesWithRelations(ctx, query, drawID)
}

// ListByRound retrieves all matches for a specific round
func (r *MatchRepository) ListByRound(ctx context.Context, drawID, round int) ([]*models.Match, error) {
	query := `
		SELECT id, draw_id, round, home_team_id, away_team_id, venue_id,
			match_date, match_time, is_prime_time, created_at, updated_at
		FROM matches
		WHERE draw_id = ? AND round = ?
		ORDER BY id
	`

	return r.listMatches(ctx, query, drawID, round)
}

// ListByTeam retrieves all matches for a specific team
func (r *MatchRepository) ListByTeam(ctx context.Context, drawID, teamID int) ([]*models.Match, error) {
	query := `
		SELECT id, draw_id, round, home_team_id, away_team_id, venue_id,
			match_date, match_time, is_prime_time, created_at, updated_at
		FROM matches
		WHERE draw_id = ? AND (home_team_id = ? OR away_team_id = ?)
		ORDER BY round, id
	`

	return r.listMatches(ctx, query, drawID, teamID, teamID)
}

// Update modifies an existing match
func (r *MatchRepository) Update(ctx context.Context, match *models.Match) error {
	query := `
		UPDATE matches
		SET round = ?, home_team_id = ?, away_team_id = ?, venue_id = ?,
			match_date = ?, match_time = ?, is_prime_time = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		match.Round, match.HomeTeamID, match.AwayTeamID, match.VenueID,
		match.MatchDate, match.MatchTime, match.IsPrimeTime, match.ID)
	if err != nil {
		return fmt.Errorf("updating match: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("match not found")
	}

	return nil
}

// UpdateBatch updates multiple matches in a single transaction
func (r *MatchRepository) UpdateBatch(ctx context.Context, matches []*models.Match) error {
	if len(matches) == 0 {
		return nil
	}

	// If we don't have a sql.DB reference, fall back to individual updates
	if r.sqlDB == nil {
		for _, match := range matches {
			if err := r.Update(ctx, match); err != nil {
				return err
			}
		}
		return nil
	}

	tx, err := r.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE matches
		SET round = ?, home_team_id = ?, away_team_id = ?, venue_id = ?,
			match_date = ?, match_time = ?, is_prime_time = ?
		WHERE id = ?
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, match := range matches {
		result, err := stmt.ExecContext(ctx,
			match.Round, match.HomeTeamID, match.AwayTeamID, match.VenueID,
			match.MatchDate, match.MatchTime, match.IsPrimeTime, match.ID)
		if err != nil {
			return fmt.Errorf("updating match %d: %w", match.ID, err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("getting rows affected: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("match %d not found", match.ID)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// Delete removes a match
func (r *MatchRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM matches WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting match: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("match not found")
	}

	return nil
}

// DeleteByDraw removes all matches for a draw
func (r *MatchRepository) DeleteByDraw(ctx context.Context, drawID int) error {
	query := `DELETE FROM matches WHERE draw_id = ?`

	_, err := r.db.ExecContext(ctx, query, drawID)
	if err != nil {
		return fmt.Errorf("deleting matches by draw: %w", err)
	}

	return nil
}

// Helper methods

func (r *MatchRepository) listMatches(ctx context.Context, query string, args ...interface{}) ([]*models.Match, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing matches: %w", err)
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

	return matches, nil
}

func (r *MatchRepository) listMatchesWithRelations(ctx context.Context, query string, args ...interface{}) ([]*models.Match, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing matches with relations: %w", err)
	}
	defer rows.Close()

	var matches []*models.Match
	for rows.Next() {
		match := &models.Match{}
		var matchDate, matchTime sql.NullTime
		var homeTeam, awayTeam models.Team
		var venue models.Venue
		var homeTeamID, awayTeamID, venueID sql.NullInt64

		err := rows.Scan(
			&match.ID, &match.DrawID, &match.Round,
			&homeTeamID, &awayTeamID, &venueID,
			&matchDate, &matchTime, &match.IsPrimeTime,
			&match.CreatedAt, &match.UpdatedAt,
			&homeTeam.ID, &homeTeam.Name, &homeTeam.ShortName, &homeTeam.City,
			&awayTeam.ID, &awayTeam.Name, &awayTeam.ShortName, &awayTeam.City,
			&venue.ID, &venue.Name, &venue.City, &venue.Capacity,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning match with relations: %w", err)
		}

		if matchDate.Valid {
			match.MatchDate = &matchDate.Time
		}
		if matchTime.Valid {
			match.MatchTime = &matchTime.Time
		}
		if homeTeamID.Valid {
			match.HomeTeamID = &[]int{int(homeTeamID.Int64)}[0]
			match.HomeTeam = &homeTeam
		}
		if awayTeamID.Valid {
			match.AwayTeamID = &[]int{int(awayTeamID.Int64)}[0]
			match.AwayTeam = &awayTeam
		}
		if venueID.Valid {
			match.VenueID = &[]int{int(venueID.Int64)}[0]
			match.Venue = &venue
		}

		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating matches: %w", err)
	}

	return matches, nil
}