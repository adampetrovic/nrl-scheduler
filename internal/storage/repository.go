package storage

import (
	"context"
	"errors"

	"github.com/adampetrovic/nrl-scheduler/internal/core/models"
)

// Common errors
var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

// VenueRepository defines methods for venue storage
type VenueRepository interface {
	Create(ctx context.Context, venue *models.Venue) error
	Get(ctx context.Context, id int) (*models.Venue, error)
	List(ctx context.Context) ([]*models.Venue, error)
	Update(ctx context.Context, venue *models.Venue) error
	Delete(ctx context.Context, id int) error
}

// TeamRepository defines methods for team storage
type TeamRepository interface {
	Create(ctx context.Context, team *models.Team) error
	Get(ctx context.Context, id int) (*models.Team, error)
	GetWithVenue(ctx context.Context, id int) (*models.Team, error)
	List(ctx context.Context) ([]*models.Team, error)
	ListWithVenues(ctx context.Context) ([]*models.Team, error)
	Update(ctx context.Context, team *models.Team) error
	Delete(ctx context.Context, id int) error
}

// DrawRepository defines methods for draw storage
type DrawRepository interface {
	Create(ctx context.Context, draw *models.Draw) error
	Get(ctx context.Context, id int) (*models.Draw, error)
	GetWithMatches(ctx context.Context, id int) (*models.Draw, error)
	List(ctx context.Context) ([]*models.Draw, error)
	Update(ctx context.Context, draw *models.Draw) error
	Delete(ctx context.Context, id int) error
}

// MatchRepository defines methods for match storage
type MatchRepository interface {
	Create(ctx context.Context, match *models.Match) error
	CreateBatch(ctx context.Context, matches []*models.Match) error
	Get(ctx context.Context, id int) (*models.Match, error)
	GetWithRelations(ctx context.Context, id int) (*models.Match, error)
	ListByDraw(ctx context.Context, drawID int) ([]*models.Match, error)
	ListByDrawWithRelations(ctx context.Context, drawID int) ([]*models.Match, error)
	ListByRound(ctx context.Context, drawID, round int) ([]*models.Match, error)
	ListByTeam(ctx context.Context, drawID, teamID int) ([]*models.Match, error)
	Update(ctx context.Context, match *models.Match) error
	UpdateBatch(ctx context.Context, matches []*models.Match) error
	Delete(ctx context.Context, id int) error
	DeleteByDraw(ctx context.Context, drawID int) error
}

// Repositories aggregates all repository interfaces
type Repositories interface {
	Venues() VenueRepository
	Teams() TeamRepository
	Draws() DrawRepository
	Matches() MatchRepository
	
	// Transaction support
	BeginTx(ctx context.Context) (Repositories, error)
	Commit() error
	Rollback() error
}