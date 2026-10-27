// talks directly to the database, forwarded to db.Queries
// this layer separate data access from business logic

package repository

import (
	"backend/db"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Values that are needed to save an Event.
type CreateEventInput struct {
	Name        string
	EventTime   time.Time
	Location    string
	Description string
	SocietyID   string
	CreatedBy   string
	Status      db.EventStatus
}

// what methods do we have for event
type EventRepository interface {
	CreateEvent(ctx context.Context, input CreateEventInput) (db.Budget, error)
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
	input CreateEventInput,
) (db.Budget, error) {
	// Change the values to PostgreSQL types before saving.
	var creator pgtype.UUID
	if err := creator.Scan(input.CreatedBy); err != nil {
		return db.Budget{}, err
	}
	params := db.CreateEventParams{
		Name:      input.Name,
		SocietyID: input.SocietyID,
		CreatedBy: creator,
		Status:    db.NullEventStatus{EventStatus: input.Status, Valid: true},
		EventTime: pgtype.Timestamptz{
			Time:  input.EventTime,
			Valid: !input.EventTime.IsZero(),
		},
		Location: pgtype.Text{String: input.Location, Valid: input.Location != ""},
		Description: pgtype.Text{
			String: input.Description,
			Valid:  input.Description != "",
		},
	}
	return r.queries.CreateEvent(ctx, params)
}

func (r *eventRepositoryImpl) ListEvents(
	ctx context.Context,
	params db.ListEventsParams,
) ([]db.Budget, error) {
	return r.queries.ListEvents(ctx, params)
}
