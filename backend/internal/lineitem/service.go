package lineitem

import (
	"context"
	"errors"
	"fmt"

	"backend/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// maxDescriptionLen bounds the free text fields. The columns are unbounded
// TEXT, so this is the service's own limit, there to stop a single request
// carrying an unreasonable payload.
const maxDescriptionLen = 1000

// errNotImplemented is returned by the mutating methods until phases 4 and 5
// fill them in. It is unexported so it never becomes part of the error
// taxonomy that callers switch on.
var errNotImplemented = errors.New("not implemented")

// CreateInput is the data needed to open a new line item.
//
// Status is deliberately absent: a new line item takes the column default of
// pr_draft, so no caller can create one that is already approved.
type CreateInput struct {
	PurchaseRequestID    int64
	EstimatedQuantity    int64
	EstimatedCostPerItem int64
	EstimatedUnitCost    int64
	Description          string
}

// EstimatesInput updates the purchase request half of a line item: what the
// requester expects to spend.
type EstimatesInput struct {
	EstimatedQuantity    int64
	EstimatedCostPerItem int64
	EstimatedUnitCost    int64
	Description          string
}

// ActualsInput records what was really spent, submitted with the
// reimbursement request.
//
// Every field is optional, because a reimbursement draft is filled in as
// receipts come in rather than in one go. A nil field leaves the stored value
// untouched; that is deliberately different from EstimatesInput, where the
// caller is editing a draft it already holds in full and sends every field.
// Making the description optional too means an update of one number cannot
// silently wipe the note attached to it.
type ActualsInput struct {
	ActualQuantity    *int64
	ActualCostPerItem *int64
	ActualUnitCost    *int64
	Description       *string
}

// Totals is the summed cost of the active line items on a purchase request,
// in cents. ActualTotal counts only items whose actuals have been recorded.
type Totals struct {
	EstimatedTotal int64
	ActualTotal    int64
}

// Service holds the line item business rules: the status lifecycle, the field
// mutability rules that follow from it, and input validation.
//
// It takes and returns plain Go types. Converting to and from the pgtype
// values the database layer needs happens here, so nothing above the service
// has to know about pgx.
type Service interface {
	Create(ctx context.Context, in CreateInput) (db.LineItem, error)
	GetByID(ctx context.Context, id int64) (db.LineItem, error)
	ListByPurchaseRequest(ctx context.Context, prID int64) ([]db.LineItem, error)
	ListByReimbursementRequest(
		ctx context.Context,
		reimbID int64,
	) ([]db.LineItem, error)
	TotalsByPurchaseRequest(ctx context.Context, prID int64) (Totals, error)

	UpdateEstimates(
		ctx context.Context,
		id int64,
		in EstimatesInput,
	) (db.LineItem, error)
	UpdateActuals(
		ctx context.Context,
		id int64,
		in ActualsInput,
	) (db.LineItem, error)
	AttachReimbursementRequest(
		ctx context.Context,
		id, reimbID int64,
	) (db.LineItem, error)
	Transition(
		ctx context.Context,
		id int64,
		to db.LineItemStatus,
	) (db.LineItem, error)
	Archive(ctx context.Context, id int64) error
}

type service struct {
	repo Repository
}

var _ Service = (*service)(nil)

// NewService returns the line item service backed by the given repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Create opens a line item against a purchase request, in pr_draft.
func (s *service) Create(
	ctx context.Context,
	in CreateInput,
) (db.LineItem, error) {
	if in.PurchaseRequestID <= 0 {
		return db.LineItem{}, fmt.Errorf(
			"%w: purchase request id must be positive, got %d",
			ErrValidation, in.PurchaseRequestID,
		)
	}
	err := validateAmounts(
		"estimated",
		in.EstimatedQuantity,
		in.EstimatedCostPerItem,
		in.EstimatedUnitCost,
	)
	if err != nil {
		return db.LineItem{}, err
	}
	if err = validateDescription(in.Description); err != nil {
		return db.LineItem{}, err
	}

	return s.repo.CreateLineItem(ctx, db.CreateLineItemParams{
		EstimatedQuantity:    in.EstimatedQuantity,
		EstimatedCostPerItem: in.EstimatedCostPerItem,
		EstimatedUnitCost:    in.EstimatedUnitCost,
		PrDescription:        textFrom(in.Description),
		PurchaseRequestID:    in.PurchaseRequestID,
	})
}

// GetByID returns one line item, including archived ones, translating a
// missing row into ErrNotFound.
func (s *service) GetByID(
	ctx context.Context,
	id int64,
) (db.LineItem, error) {
	item, err := s.repo.GetLineItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.LineItem{}, fmt.Errorf("%w: line item %d", ErrNotFound, id)
		}
		return db.LineItem{}, err
	}
	return item, nil
}

// ListByPurchaseRequest returns the active line items on a purchase request.
func (s *service) ListByPurchaseRequest(
	ctx context.Context,
	prID int64,
) ([]db.LineItem, error) {
	return s.repo.GetLineItemsByPRID(ctx, prID)
}

// ListByReimbursementRequest returns the active line items attached to a
// reimbursement request.
func (s *service) ListByReimbursementRequest(
	ctx context.Context,
	reimbID int64,
) ([]db.LineItem, error) {
	return s.repo.GetLineItemsByReimbID(ctx, int8From(&reimbID))
}

// TotalsByPurchaseRequest sums the active line items on a purchase request.
//
// An item contributes to ActualTotal only once both its actual quantity and
// actual cost per item are recorded, so a half filled reimbursement draft does
// not report a misleadingly small figure.
func (s *service) TotalsByPurchaseRequest(
	ctx context.Context,
	prID int64,
) (Totals, error) {
	items, err := s.repo.GetLineItemsByPRID(ctx, prID)
	if err != nil {
		return Totals{}, err
	}

	var totals Totals
	for _, item := range items {
		totals.EstimatedTotal += item.EstimatedQuantity * item.EstimatedCostPerItem
		if item.ActualQuantity.Valid && item.ActualCostPerItem.Valid {
			totals.ActualTotal +=
				item.ActualQuantity.Int64 * item.ActualCostPerItem.Int64
		}
	}
	return totals, nil
}

// UpdateEstimates revises what the requester expects to spend.
//
// Estimates are editable only while the line item is still a draft. Once it
// has been submitted, the numbers an approver is looking at, or has already
// approved, must not move underneath them.
func (s *service) UpdateEstimates(
	ctx context.Context,
	id int64,
	in EstimatesInput,
) (db.LineItem, error) {
	item, err := s.GetByID(ctx, id)
	if err != nil {
		return db.LineItem{}, err
	}
	if item.Status != db.LineItemStatusPrDraft {
		return db.LineItem{}, fmt.Errorf(
			"%w: estimates are editable only in %q, line item %d is %q",
			ErrImmutableField, db.LineItemStatusPrDraft, id, item.Status,
		)
	}
	err = validateAmounts(
		"estimated",
		in.EstimatedQuantity,
		in.EstimatedCostPerItem,
		in.EstimatedUnitCost,
	)
	if err != nil {
		return db.LineItem{}, err
	}
	if err = validateDescription(in.Description); err != nil {
		return db.LineItem{}, err
	}

	params := paramsFrom(item)
	params.EstimatedQuantity = in.EstimatedQuantity
	params.EstimatedCostPerItem = in.EstimatedCostPerItem
	params.EstimatedUnitCost = in.EstimatedUnitCost
	params.PrDescription = textFrom(in.Description)
	return s.repo.UpdateLineItem(ctx, params)
}

// UpdateActuals records what was really spent.
//
// Actuals are writable only while the reimbursement is still a draft. Once
// submitted they are frozen, for the same reason estimates freeze at
// submission: a treasurer must be deciding on fixed numbers.
func (s *service) UpdateActuals(
	ctx context.Context,
	id int64,
	in ActualsInput,
) (db.LineItem, error) {
	item, err := s.GetByID(ctx, id)
	if err != nil {
		return db.LineItem{}, err
	}
	if item.Status != db.LineItemStatusReimbDraft {
		return db.LineItem{}, fmt.Errorf(
			"%w: actuals are writable only in %q, line item %d is %q",
			ErrImmutableField, db.LineItemStatusReimbDraft, id, item.Status,
		)
	}
	if err = validateActuals(in); err != nil {
		return db.LineItem{}, err
	}

	params := paramsFrom(item)
	if in.ActualQuantity != nil {
		params.ActualQuantity = int8From(in.ActualQuantity)
	}
	if in.ActualCostPerItem != nil {
		params.ActualCostPerItem = int8From(in.ActualCostPerItem)
	}
	if in.ActualUnitCost != nil {
		params.ActualUnitCost = int8From(in.ActualUnitCost)
	}
	if in.Description != nil {
		params.ReimbDescription = textFrom(*in.Description)
	}
	return s.repo.UpdateLineItem(ctx, params)
}

// AttachReimbursementRequest moves an approved line item into the
// reimbursement phase, linking it to the reimbursement request that will carry
// it.
//
// The attachment and the move to reimb_draft are one write: a line item that
// pointed at a reimbursement request while still sitting in pr_approved, or
// that reached reimb_draft attached to nothing, would be a state the rest of
// the package does not expect. Whether the move is legal is asked of the
// transition table rather than hardcoded here, so the lifecycle has one
// definition.
func (s *service) AttachReimbursementRequest(
	ctx context.Context,
	id, reimbID int64,
) (db.LineItem, error) {
	if reimbID <= 0 {
		return db.LineItem{}, fmt.Errorf(
			"%w: reimbursement request id must be positive, got %d",
			ErrValidation, reimbID,
		)
	}
	item, err := s.GetByID(ctx, id)
	if err != nil {
		return db.LineItem{}, err
	}
	if !CanTransition(item.Status, db.LineItemStatusReimbDraft) {
		return db.LineItem{}, fmt.Errorf(
			"%w: line item %d is %q, cannot enter the reimbursement phase",
			ErrInvalidTransition, id, item.Status,
		)
	}

	params := paramsFrom(item)
	params.ReimbursementRequestID = int8From(&reimbID)
	params.Status = db.LineItemStatusReimbDraft
	return s.repo.UpdateLineItem(ctx, params)
}

// Transition is implemented in phase 5.
func (s *service) Transition(
	_ context.Context,
	_ int64,
	_ db.LineItemStatus,
) (db.LineItem, error) {
	return db.LineItem{}, errNotImplemented
}

// Archive removes a line item from view by clearing is_active.
//
// It is a soft delete: a line item cited by a purchase request must not vanish
// from underneath it, so the row stays and the reads filter it out. The status
// is left alone, since archiving is not a step in the lifecycle and an
// archived item should still say where it had got to. Archiving an already
// archived item is a no-op rather than an error.
func (s *service) Archive(ctx context.Context, id int64) error {
	item, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !item.IsActive {
		return nil
	}

	params := paramsFrom(item)
	params.IsActive = false
	_, err = s.repo.UpdateLineItem(ctx, params)
	return err
}

// validateAmounts bounds the quantity and money fields of one phase. kind
// names that phase ("estimated" or "actual") so the error says which half of
// the line item was rejected.
//
// Unit cost is checked for sign only: it cannot be reconciled against quantity
// and cost per item, because the number of items in a pack is not stored.
func validateAmounts(kind string, quantity, costPerItem, unitCost int64) error {
	if quantity <= 0 {
		return fmt.Errorf(
			"%w: %s quantity must be positive, got %d",
			ErrValidation, kind, quantity,
		)
	}
	if costPerItem < 0 {
		return fmt.Errorf(
			"%w: %s cost per item must not be negative, got %d",
			ErrValidation, kind, costPerItem,
		)
	}
	if unitCost < 0 {
		return fmt.Errorf(
			"%w: %s unit cost must not be negative, got %d",
			ErrValidation, kind, unitCost,
		)
	}
	return nil
}

// validateActuals bounds whichever actual fields the caller supplied. Absent
// fields are not defaulted to zero and then rejected: leaving a value unset is
// how a reimbursement draft is filled in over time.
func validateActuals(in ActualsInput) error {
	if in.ActualQuantity != nil && *in.ActualQuantity <= 0 {
		return fmt.Errorf(
			"%w: actual quantity must be positive, got %d",
			ErrValidation, *in.ActualQuantity,
		)
	}
	if in.ActualCostPerItem != nil && *in.ActualCostPerItem < 0 {
		return fmt.Errorf(
			"%w: actual cost per item must not be negative, got %d",
			ErrValidation, *in.ActualCostPerItem,
		)
	}
	if in.ActualUnitCost != nil && *in.ActualUnitCost < 0 {
		return fmt.Errorf(
			"%w: actual unit cost must not be negative, got %d",
			ErrValidation, *in.ActualUnitCost,
		)
	}
	if in.Description != nil {
		return validateDescription(*in.Description)
	}
	return nil
}

func validateDescription(description string) error {
	if len(description) > maxDescriptionLen {
		return fmt.Errorf(
			"%w: description is %d characters, limit is %d",
			ErrValidation, len(description), maxDescriptionLen,
		)
	}
	return nil
}

// paramsFrom builds a full update parameter set from a loaded row.
//
// UpdateLineItem writes all twelve columns, so every mutating method has to
// read the current row and write it back whole. Building the params in one
// place is what keeps "change only what was asked for" a property of the
// package rather than something each method has to remember.
func paramsFrom(item db.LineItem) db.UpdateLineItemParams {
	return db.UpdateLineItemParams{
		ID:                     item.ID,
		EstimatedQuantity:      item.EstimatedQuantity,
		EstimatedCostPerItem:   item.EstimatedCostPerItem,
		EstimatedUnitCost:      item.EstimatedUnitCost,
		PrDescription:          item.PrDescription,
		ActualQuantity:         item.ActualQuantity,
		ActualCostPerItem:      item.ActualCostPerItem,
		ActualUnitCost:         item.ActualUnitCost,
		ReimbDescription:       item.ReimbDescription,
		ReimbursementRequestID: item.ReimbursementRequestID,
		Status:                 item.Status,
		IsActive:               item.IsActive,
	}
}

// textFrom maps a string to a nullable column, treating the empty string as
// absent rather than as an empty description.
func textFrom(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

// int8From maps an optional integer to a nullable column.
func int8From(v *int64) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *v, Valid: true}
}
