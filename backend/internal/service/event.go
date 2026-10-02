// business logic and validation

package service

import (
	"backend/db"
	"backend/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// define EventService, what event can do
// create interface for testing
type EventService interface {
	CreateEvent(
		ctx context.Context,
		name string,
		eventTime time.Time,
		budget int64,
		status db.EventStatus,
		location string,
		description string,
		societyID string,
		createdBy string,
	) (db.Event, error)
	// add more... delete/update...
}

// create eventservice struct and add func to it
type eventServiceImpl struct {
	repo repository.EventRepository
}

func NewEventService(repo repository.EventRepository) EventService {
	return &eventServiceImpl{repo: repo}
}

func (s *eventServiceImpl) CreateEvent(
	ctx context.Context,
	name string,
	eventTime time.Time,
	budget int64,
	status db.EventStatus,
	location string,
	description string,
	societyID string,
	createdBy string,
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
	// convert plain types into pgtype
	// convert time.Time to pytype.Timestamptz
	eventTimePG := pgtype.Timestamptz{Time: eventTime, Valid: true}
	// convert string to Text
	locationPG := pgtype.Text{String: location, Valid: true}
	descriptionPG := pgtype.Text{String: description, Valid: description != ""}
	// convert societyID string to UUID
	var societyIDPG pgtype.UUID
	if err := societyIDPG.Scan(societyID); err != nil {
		return db.Event{}, errors.New("invalid society ID")
	}
	// convert createdBy string to UUID
	var createdByPG pgtype.UUID
	if err := createdByPG.Scan(createdBy); err != nil {
		return db.Event{}, errors.New("invalid created by ID")
	}
	params := db.CreateEventParams{
		EventName:   name,
		EventTime:   eventTimePG,
		Budget:      budget,
		Status:      status,
		Location:    locationPG,
		Description: descriptionPG,
		SocietyID:   societyIDPG,
		CreatedBy:   createdByPG,
	}
	// create the event
	return s.repo.CreateEvent(ctx, params)
}
