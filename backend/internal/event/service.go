// business logic and validation

package event

import (
	"context"
	"errors"
	"time"
	"backend/db"
	)

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
	// TODO: auth check

	// event_name cannot be empty
	if name == "" {
		return db.Event{}, errors.New("event name is required")
	}
	// budget cannot be negative
	if budget < 0 {
		return db.Event{}, errors.New("budget could not be negative")
	}
	// TODO: convert plain types into pgtype

	// TODO: create the event
	return db.Event{}, errors.New("TODO: create event not implemented")
}

