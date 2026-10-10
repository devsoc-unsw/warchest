package service

import (
	"backend/db"
	"backend/internal/repository"
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgtype"
)

// define purchaseRequestService, what event can do
// create interface for testing
type PurchaseRequestService interface {
	CreatePR(
		ctx context.Context,
		createdByUserID int64,
		eventID *int64,
		portfolioID *int64,
		title string,
		description string,
		expectedBudget *int64,
		actualBudget *int64,
	) (db.PurchaseRequest, error)
	//updatePR
	//deletePR
}

// create purchaseRequest Service struct and add function to it
type purchaseRequestImpl struct {
	repo repository.PurchaseRequestRepository
}

func NewPurchaseRequestService(
	repo repository.PurchaseRequestRepository,
) PurchaseRequestService {
	return &purchaseRequestServiceImpl{repo: repo}
}
func (s *purchaseRequestServiceImpl) CreatePR(
	ctx context.Context,
	createdByUserID int64,
	eventID *int64,
	portfolioID *int64,
	title string,
	description string,
	expectedBudget *int64,
	actualBudget *int64,
) (db.PurchaseRequest, error) {
	//validation
	if createdByUserID <= 0 {
		return db.PurchaseRequest{}, errors.New("created by user ID is required")
	}
	if title == "" {
		return db.PurchaseRequest{}, errors.New("title is required")
	}
	// exactly one (event or profolio) must be provided
	if (eventID == nil && portfolioID == nil) ||
		(eventID != nil && portfolioID != nil) {
		return db.PurchaseRequest{}, errors.New(
			"exactly one of event ID or portfolio ID must be provided",
		)
	}
	if expectedBudget != nil && *expectedBudget < 0 {
		return db.PurchaseRequest{}, errors.New(
			"expected budget cannot be negative",
		)
	}
	if actualBudget != nil && *actualBudget < 0 {
		return db.PurchaseRequest{}, errors.New(
			"actual budget cannot be negative",
		)
	}
	// convert nullable values to pgtype

	var eventIDPG pgtype.Int8
	if eventID != nil {
		eventIDPG = pgtype.Int8{
			Int64: *eventID,
			Valid: true,
		}
	}

	var portfolioIDPG pgtype.Int8
	if portfolioID != nil {
		portfolioIDPG = pgtype.Int8{
			Int64: *portfolioID,
			Valid: true,
		}
	}

	descriptionPG := pgtype.Text{
		String: description,
		Valid:  description != "",
	}

	var expectedBudgetPG pgtype.Int8
	if expectedBudget != nil {
		expectedBudgetPG = pgtype.Int8{
			Int64: *expectedBudget,
			Valid: true,
		}
	}

	var actualBudgetPG pgtype.Int8
	if actualBudget != nil {
		actualBudgetPG = pgtype.Int8{
			Int64: *actualBudget,
			Valid: true,
		}
	}

	params := db.CreatePRParams{
		CreatedByUserID: createdByUserID,
		EventID:         eventIDPG,
		PortfolioID:     portfolioIDPG,
		Title:           title,
		Description:     descriptionPG,
		ExpectedBudget:  expectedBudgetPG,
		ActualBudget:    actualBudgetPG,
	}

	return s.repo.CreatePR(ctx, params)
}
