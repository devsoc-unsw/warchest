package service

import (
	"backend/db"
	"backend/internal/repository"
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Used by the handler to distinguish a database failure from invalid input.
var ErrCreateEvent = errors.New("failed to create event")

// Budget is input only. Money Module owns monetary storage.
type CreateEventInput struct {
	Name        string
	EventTime   time.Time
	Location    string
	Description string
	SocietyID   string
	CreatedBy   string
	Status      db.EventStatus
	Budget      uint64
}

type EventService interface {
	CreateEvent(ctx context.Context, input CreateEventInput) (db.Budget, error)
}

type eventServiceImpl struct {
	repo repository.EventRepository
}

func NewEventService(repo repository.EventRepository) EventService {
	return &eventServiceImpl{repo: repo}
}

func (s *eventServiceImpl) CreateEvent(
	ctx context.Context,
	input CreateEventInput,
) (db.Budget, error) {
	// TODO(auth): authenticate the caller and authorise creation within this society.
	// TODO(auth): apply Event/Portfolio RBAC rules in business validation
	// TODO: validate type-specific fields and lifecycle rules in the service.
	// TODO(Money Module): confirm creation and Budget linking requirements
	// with the Money Module contract.
	if input.Status != db.EventStatusDRAFT && input.Status != db.EventStatusINPROGRESS {
		return db.Budget{}, errors.New("status must be DRAFT or IN_PROGRESS")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return db.Budget{}, errors.New("event name is required")
	}
	societyID := strings.TrimSpace(input.SocietyID)
	if societyID == "" {
		return db.Budget{}, errors.New("society ID is required")
	}
	var creator pgtype.UUID
	if err := creator.Scan(input.CreatedBy); err != nil {
		return db.Budget{}, errors.New("invalid created by ID")
	}
	// The frontend sends DRAFT for Create Draft and IN_PROGRESS for Create.
	// TODO(Money Module): save input.Budget after its API and amount unit are agreed.
	params := db.CreateEventParams{
		Name:      name,
		SocietyID: societyID,
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
	event, err := s.repo.CreateEvent(ctx, params)
	if err != nil {
		log.Printf("create event failed: %v", err)
		return db.Budget{}, ErrCreateEvent
	}
	return event, nil
}
