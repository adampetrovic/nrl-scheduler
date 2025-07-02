package sqlite

import (
	"context"
	"database/sql"

	"github.com/adampetrovic/nrl-scheduler/internal/storage"
)

// Repositories implements storage.Repositories using SQLite
type Repositories struct {
	db           *sql.DB
	tx           *sql.Tx
	venues       *VenueRepository
	teams        *TeamRepository
	draws        *DrawRepository
	matches      *MatchRepository
}

// NewRepositories creates a new repositories instance
func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		db:      db,
		venues:  NewVenueRepository(db),
		teams:   NewTeamRepository(db),
		draws:   NewDrawRepository(db),
		matches: NewMatchRepository(db),
	}
}

// Venues returns the venue repository
func (r *Repositories) Venues() storage.VenueRepository {
	return r.venues
}

// Teams returns the team repository
func (r *Repositories) Teams() storage.TeamRepository {
	return r.teams
}

// Draws returns the draw repository
func (r *Repositories) Draws() storage.DrawRepository {
	return r.draws
}

// Matches returns the match repository
func (r *Repositories) Matches() storage.MatchRepository {
	return r.matches
}

// BeginTx starts a transaction and returns a new repositories instance
func (r *Repositories) BeginTx(ctx context.Context) (storage.Repositories, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &Repositories{
		db:      r.db,
		tx:      tx,
		venues:  NewTxVenueRepository(tx),
		teams:   NewTxTeamRepository(tx),
		draws:   NewTxDrawRepository(tx),
		matches: NewTxMatchRepository(tx),
	}, nil
}

// Commit commits the transaction
func (r *Repositories) Commit() error {
	if r.tx == nil {
		return nil
	}
	return r.tx.Commit()
}

// Rollback rolls back the transaction
func (r *Repositories) Rollback() error {
	if r.tx == nil {
		return nil
	}
	return r.tx.Rollback()
}

// Transaction repository implementations using sql.Tx

// NewTxVenueRepository creates a venue repository that uses a transaction
func NewTxVenueRepository(tx *sql.Tx) *VenueRepository {
	return NewVenueRepository(tx)
}

// NewTxTeamRepository creates a team repository that uses a transaction
func NewTxTeamRepository(tx *sql.Tx) *TeamRepository {
	return NewTeamRepository(tx)
}

// NewTxDrawRepository creates a draw repository that uses a transaction
func NewTxDrawRepository(tx *sql.Tx) *DrawRepository {
	return NewDrawRepository(tx)
}

// NewTxMatchRepository creates a match repository that uses a transaction
func NewTxMatchRepository(tx *sql.Tx) *MatchRepository {
	return NewMatchRepository(tx)
}