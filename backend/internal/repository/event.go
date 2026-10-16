// talks directly to the database, forwarded to db.Queries
// this layer separate data access from business logic

package repository

import (
	"backend/db"
	"context"
)

// what methods do we have for event
type EventRepository interface {
	CreateEvent(ctx context.Context, params db.CreateEventParams) (db.Budget, error)
	ListEvents(ctx context.Context, params db.ListEventsParams) ([]db.Budget, error)
	// add more methods for event...
}

// define eventRepositoryImpl struct
type eventRepositoryImpl struct {
	queries *db.Queries
}

// constructor function to create eventRepositoryImpl
func NewEventRepository(queries *db.Queries) EventRepository {
	// stores the incoming queries into its queries field
	return &eventRepositoryImpl{queries: queries}
}

// write func for eventRepositoryImpl to CreateEvent
func (r *eventRepositoryImpl) CreateEvent(
	ctx context.Context,
	params db.CreateEventParams,
) (db.Budget, error) {
	return r.queries.CreateEvent(ctx, params)
	// simply call sqlc-generated method and return the result
}

func (r *eventRepositoryImpl) ListEvents(
	ctx context.Context,
	params db.ListEventsParams,
) ([]db.Budget, error) {
	return r.queries.ListEvents(ctx, params)
}
