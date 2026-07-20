// business logic and validation

package event

import {
	"context"
	"errors"
	"time"
	"backend/db"
	"github.com/jackc/pgx/v5/pgtype" // UUID, Timestamptz, Numeric, Text - some database specifix types
}

// define EventService, what event can do
// create interface for testing
type EventService interface {
	CreateEvent(
		ctx context.Context,
		name string,
		eventTime time.Time,
		budget float64,
		location string,
		description string,
		societyID string,
		createdBy int64,
	) (db.Event, error)
	// add more... delete/update...
}

// create eventservice struct and add func to it
type eventServiceImpl struct {
	repo EventRepository
}

func NewEventService(repo EventRepository) EventService {
	return &eventServiceImpl{repo: repo}
}

func (s *eventServiceImpl) CreateEvent(
		ctx context.Context,
		name string,
		eventTime time.Time,
		budget float64,
		location string,
		description string,
		societyID string,
		createdBy int64,
) (db.Event, error) {
	// logic and validation

	
}

