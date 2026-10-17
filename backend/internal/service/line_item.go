// Line item business rules: the status lifecycle, the field mutability rules
// that fall out of it, and input validation.
//
// This is the middle of three layers. It depends only on the
// repository.LineItemRepository interface, so the rules here can be exercised
// against a fake without a database or an HTTP server. repository/line_item.go
// supplies the Postgres implementation and handler/line_item.go exposes these
// methods over HTTP.
//
// All money values handled here are minor units (cents). Nothing
// here converts between minor and major units; callers pass and receive cents.
//
// The three cost fields describe a purchase at different granularities. For
// two bags of ten paper cups at $10 a bag:
//
//	quantity      = 2   (bags bought)
//	cost per item = $10 (price of one bag)
//	unit cost     = $1  (price of one cup)
//
// Quantity times cost per item gives the total. Unit cost is entered by the
// user for reporting and cannot be derived from the other two, since the
// number of cups in a bag is not recorded; it is therefore bounds checked but
// never cross checked against them.
//
// Each field exists twice: the estimated_* values are what the requester
// expects to spend, submitted with the purchase request, and the actual_*
// values are what was really spent, submitted with the reimbursement request
// once the goods have been bought.

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"backend/db"
	"backend/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Sentinel errors returned by the service. They exist so that the HTTP layer
// can map failures to status codes with errors.Is, without inspecting strings
// or knowing anything about the rules that produced them.
//
// Wrap them with fmt.Errorf("%w: detail", ErrX) to add context.
var (
	// ErrLineItemNotFound is returned when no line item exists for the given id.
	// Maps to 404.
	ErrLineItemNotFound = errors.New("line item not found")

	// ErrInvalidLineItem is returned when input fails a bounds or presence check.
	// Maps to 400.
	ErrInvalidLineItem = errors.New("invalid line item")

	// ErrInvalidLineItemTransition is returned when a status change is not a legal
	// edge of the lifecycle. Maps to 409.
	ErrInvalidLineItemTransition = errors.New("invalid status transition")

	// ErrImmutableLineItemField is returned when a field is written in a status that
	// does not permit editing it. Maps to 409.
	ErrImmutableLineItemField = errors.New(
		"field is not editable in the current status",
	)
)

// AllLineItemStatuses lists every status in lifecycle order. It is the authoritative
// set for iterating the lifecycle (tests, and later the API surface that tells
// the frontend which actions to render).
var AllLineItemStatuses = []db.LineItemStatus{
	db.LineItemStatusPrDraft,
	db.LineItemStatusPrPending,
	db.LineItemStatusPrApproved,
	db.LineItemStatusPrRejected,
	db.LineItemStatusReimbDraft,
	db.LineItemStatusReimbPending,
	db.LineItemStatusReimbApproved,
	db.LineItemStatusReimbRejected,
}

// lineItemTransitions maps each status to the statuses it may legally move to.
// It is held as data rather than as conditionals so that the legal next states
// can be served to callers, not just enforced.
//
// A line item starts in pr_draft and is submitted for approval. Approval moves
// it into the reimbursement phase, where the buyer records what was actually
// spent and submits again. An entry with no outgoing edges is terminal.
var lineItemTransitions = map[db.LineItemStatus][]db.LineItemStatus{
	// Purchase request phase.
	db.LineItemStatusPrDraft: {db.LineItemStatusPrPending},
	db.LineItemStatusPrPending: {
		db.LineItemStatusPrApproved,
		db.LineItemStatusPrRejected,
	},
	db.LineItemStatusPrApproved: {db.LineItemStatusReimbDraft},
	db.LineItemStatusPrRejected: {},

	// Reimbursement phase.
	db.LineItemStatusReimbDraft: {db.LineItemStatusReimbPending},
	db.LineItemStatusReimbPending: {
		db.LineItemStatusReimbApproved,
		db.LineItemStatusReimbRejected,
	},
	db.LineItemStatusReimbApproved: {},
	db.LineItemStatusReimbRejected: {},
}

// CanTransitionLineItem reports whether moving from one status to another is a legal
// edge of the lifecycle. Unknown statuses have no edges, so they always
// report false.
func CanTransitionLineItem(from, to db.LineItemStatus) bool {
	for _, next := range lineItemTransitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

// NextLineItemStates returns the statuses reachable from the given status. The result
// is a copy, so callers cannot mutate the transition table through it. A
// terminal or unknown status yields an empty slice.
func NextLineItemStates(from db.LineItemStatus) []db.LineItemStatus {
	next := lineItemTransitions[from]
	out := make([]db.LineItemStatus, len(next))
	copy(out, next)
	return out
}

// IsTerminalLineItemStatus reports whether the status is a known end state,
// meaning nothing further can happen to the line item. Unknown statuses report
// false: they are not terminal, they are invalid.
func IsTerminalLineItemStatus(s db.LineItemStatus) bool {
	next, known := lineItemTransitions[s]
	return known && len(next) == 0
}

// IsLineItemReimbursementPhase reports whether the status belongs to the reimbursement
// half of the lifecycle, which is where the actual_* fields become meaningful.
func IsLineItemReimbursementPhase(s db.LineItemStatus) bool {
	switch s {
	case db.LineItemStatusReimbDraft,
		db.LineItemStatusReimbPending,
		db.LineItemStatusReimbApproved,
		db.LineItemStatusReimbRejected:
		return true
	default:
		return false
	}
}

// maxLineItemDescriptionLen bounds the free text fields. The columns are unbounded
// TEXT, so this is the service's own limit, there to stop a single request
// carrying an unreasonable payload.
const maxLineItemDescriptionLen = 1000

// CreateLineItemInput is the data needed to open a new line item.
//
// Status is deliberately absent: a new line item takes the column default of
// pr_draft, so no caller can create one that is already approved.
type CreateLineItemInput struct {
	PurchaseRequestID    int64
	EstimatedQuantity    int64
	EstimatedCostPerItem int64
	EstimatedUnitCost    int64
	Description          string
}

// LineItemEstimatesInput updates the purchase request half of a line item: what the
// requester expects to spend.
type LineItemEstimatesInput struct {
	EstimatedQuantity    int64
	EstimatedCostPerItem int64
	EstimatedUnitCost    int64
	Description          string
}

// LineItemActualsInput records what was really spent, submitted with the
// reimbursement request.
//
// Every field is optional, because a reimbursement draft is filled in as
// receipts come in rather than in one go. A nil field leaves the stored value
// untouched; that is deliberately different from LineItemEstimatesInput, where the
// caller is editing a draft it already holds in full and sends every field.
// Making the description optional too means an update of one number cannot
// silently wipe the note attached to it.
type LineItemActualsInput struct {
	ActualQuantity    *int64
	ActualCostPerItem *int64
	ActualUnitCost    *int64
	Description       *string
}

// LineItemTotals is the summed cost of the active line items on a purchase request,
// in cents. ActualTotal counts only items whose actuals have been recorded.
type LineItemTotals struct {
	EstimatedTotal int64
	ActualTotal    int64
}

// LineItemService holds the line item business rules: the status lifecycle, the field
// mutability rules that follow from it, and input validation.
//
// It takes and returns plain Go types. Converting to and from the pgtype
// values the database layer needs happens here, so nothing above the service
// has to know about pgx.
type LineItemService interface {
	Create(ctx context.Context, in CreateLineItemInput) (db.LineItem, error)
	GetByID(ctx context.Context, id int64) (db.LineItem, error)
	ListByPurchaseRequest(ctx context.Context, prID int64) ([]db.LineItem, error)
	ListByReimbursementRequest(
		ctx context.Context,
		reimbID int64,
	) ([]db.LineItem, error)
	TotalsByPurchaseRequest(ctx context.Context, prID int64) (LineItemTotals, error)

	UpdateEstimates(
		ctx context.Context,
		id int64,
		in LineItemEstimatesInput,
	) (db.LineItem, error)
	UpdateActuals(
		ctx context.Context,
		id int64,
		in LineItemActualsInput,
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

type lineItemServiceImpl struct {
	repo repository.LineItemRepository
}

var _ LineItemService = (*lineItemServiceImpl)(nil)

// NewLineItemService returns the line item service backed by the given repository.
func NewLineItemService(repo repository.LineItemRepository) LineItemService {
	return &lineItemServiceImpl{repo: repo}
}

// Create opens a line item against a purchase request, in pr_draft.
func (s *lineItemServiceImpl) Create(
	ctx context.Context,
	in CreateLineItemInput,
) (db.LineItem, error) {
	if in.PurchaseRequestID <= 0 {
		return db.LineItem{}, fmt.Errorf(
			"%w: purchase request id must be positive, got %d",
			ErrInvalidLineItem, in.PurchaseRequestID,
		)
	}
	err := validateLineItemAmounts(
		"estimated",
		in.EstimatedQuantity,
		in.EstimatedCostPerItem,
		in.EstimatedUnitCost,
	)
	if err != nil {
		return db.LineItem{}, err
	}
	if err = validateLineItemDescription(in.Description); err != nil {
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
// missing row into ErrLineItemNotFound.
func (s *lineItemServiceImpl) GetByID(
	ctx context.Context,
	id int64,
) (db.LineItem, error) {
	item, err := s.repo.GetLineItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.LineItem{}, fmt.Errorf(
				"%w: line item %d", ErrLineItemNotFound, id,
			)
		}
		return db.LineItem{}, err
	}
	return item, nil
}

// ListByPurchaseRequest returns the active line items on a purchase request.
func (s *lineItemServiceImpl) ListByPurchaseRequest(
	ctx context.Context,
	prID int64,
) ([]db.LineItem, error) {
	return s.repo.GetLineItemsByPRID(ctx, prID)
}

// ListByReimbursementRequest returns the active line items attached to a
// reimbursement request.
func (s *lineItemServiceImpl) ListByReimbursementRequest(
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
func (s *lineItemServiceImpl) TotalsByPurchaseRequest(
	ctx context.Context,
	prID int64,
) (LineItemTotals, error) {
	items, err := s.repo.GetLineItemsByPRID(ctx, prID)
	if err != nil {
		return LineItemTotals{}, err
	}

	var totals LineItemTotals
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
func (s *lineItemServiceImpl) UpdateEstimates(
	ctx context.Context,
	id int64,
	in LineItemEstimatesInput,
) (db.LineItem, error) {
	item, err := s.GetByID(ctx, id)
	if err != nil {
		return db.LineItem{}, err
	}
	if item.Status != db.LineItemStatusPrDraft {
		return db.LineItem{}, fmt.Errorf(
			"%w: estimates are editable only in %q, line item %d is %q",
			ErrImmutableLineItemField, db.LineItemStatusPrDraft, id, item.Status,
		)
	}
	err = validateLineItemAmounts(
		"estimated",
		in.EstimatedQuantity,
		in.EstimatedCostPerItem,
		in.EstimatedUnitCost,
	)
	if err != nil {
		return db.LineItem{}, err
	}
	if err = validateLineItemDescription(in.Description); err != nil {
		return db.LineItem{}, err
	}

	params := lineItemParamsFrom(item)
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
func (s *lineItemServiceImpl) UpdateActuals(
	ctx context.Context,
	id int64,
	in LineItemActualsInput,
) (db.LineItem, error) {
	item, err := s.GetByID(ctx, id)
	if err != nil {
		return db.LineItem{}, err
	}
	if item.Status != db.LineItemStatusReimbDraft {
		return db.LineItem{}, fmt.Errorf(
			"%w: actuals are writable only in %q, line item %d is %q",
			ErrImmutableLineItemField, db.LineItemStatusReimbDraft, id, item.Status,
		)
	}
	if err = validateLineItemActuals(in); err != nil {
		return db.LineItem{}, err
	}

	params := lineItemParamsFrom(item)
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
func (s *lineItemServiceImpl) AttachReimbursementRequest(
	ctx context.Context,
	id, reimbID int64,
) (db.LineItem, error) {
	if reimbID <= 0 {
		return db.LineItem{}, fmt.Errorf(
			"%w: reimbursement request id must be positive, got %d",
			ErrInvalidLineItem, reimbID,
		)
	}
	item, err := s.GetByID(ctx, id)
	if err != nil {
		return db.LineItem{}, err
	}
	if !CanTransitionLineItem(item.Status, db.LineItemStatusReimbDraft) {
		return db.LineItem{}, fmt.Errorf(
			"%w: line item %d is %q, cannot enter the reimbursement phase",
			ErrInvalidLineItemTransition, id, item.Status,
		)
	}

	params := lineItemParamsFrom(item)
	params.ReimbursementRequestID = int8From(&reimbID)
	params.Status = db.LineItemStatusReimbDraft
	return s.repo.UpdateLineItem(ctx, params)
}

// Transition moves a line item along the lifecycle.
//
// Only edges in the transition table are permitted, so a line item cannot skip
// approval, return from a terminal state, or jump between the two phases.
//
// TODO: this reads the current status and then writes, with nothing holding
// the row in between. Two callers can both observe pr_draft and both write.
// Closing that needs a status guarded UPDATE (... WHERE id = $1 AND status =
// $2), which is a change to queries/line_items.sql and a regenerate.
func (s *lineItemServiceImpl) Transition(
	ctx context.Context,
	id int64,
	to db.LineItemStatus,
) (db.LineItem, error) {
	item, err := s.GetByID(ctx, id)
	if err != nil {
		return db.LineItem{}, err
	}
	if !item.IsActive {
		return db.LineItem{}, fmt.Errorf(
			"%w: line item %d is archived", ErrInvalidLineItemTransition, id,
		)
	}
	if !CanTransitionLineItem(item.Status, to) {
		return db.LineItem{}, fmt.Errorf(
			"%w: line item %d cannot move from %q to %q",
			ErrInvalidLineItemTransition, id, item.Status, to,
		)
	}
	if to == db.LineItemStatusReimbPending {
		if err = requireLineItemActuals(item); err != nil {
			return db.LineItem{}, err
		}
	}

	params := lineItemParamsFrom(item)
	params.Status = to
	return s.repo.UpdateLineItem(ctx, params)
}

// requireLineItemActuals refuses to submit a reimbursement that has not been filled
// in. Without it an empty claim reaches a treasurer looking like a real one,
// since the actual columns are nullable and default to nothing.
func requireLineItemActuals(item db.LineItem) error {
	var missing []string
	if !item.ActualQuantity.Valid {
		missing = append(missing, "quantity")
	}
	if !item.ActualCostPerItem.Valid {
		missing = append(missing, "cost per item")
	}
	if !item.ActualUnitCost.Valid {
		missing = append(missing, "unit cost")
	}
	if len(missing) > 0 {
		return fmt.Errorf(
			"%w: cannot submit for reimbursement, missing actual %s",
			ErrInvalidLineItem, strings.Join(missing, ", "),
		)
	}
	return nil
}

// Archive removes a line item from view by clearing is_active.
//
// It is a soft delete: a line item cited by a purchase request must not vanish
// from underneath it, so the row stays and the reads filter it out. The status
// is left alone, since archiving is not a step in the lifecycle and an
// archived item should still say where it had got to. Archiving an already
// archived item is a no-op rather than an error.
func (s *lineItemServiceImpl) Archive(ctx context.Context, id int64) error {
	item, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !item.IsActive {
		return nil
	}

	params := lineItemParamsFrom(item)
	params.IsActive = false
	_, err = s.repo.UpdateLineItem(ctx, params)
	return err
}

// validateLineItemAmounts bounds the quantity and money fields of one phase. kind
// names that phase ("estimated" or "actual") so the error says which half of
// the line item was rejected.
//
// Unit cost is checked for sign only: it cannot be reconciled against quantity
// and cost per item, because the number of items in a pack is not stored.
func validateLineItemAmounts(kind string, quantity, costPerItem, unitCost int64) error {
	if quantity <= 0 {
		return fmt.Errorf(
			"%w: %s quantity must be positive, got %d",
			ErrInvalidLineItem, kind, quantity,
		)
	}
	if costPerItem < 0 {
		return fmt.Errorf(
			"%w: %s cost per item must not be negative, got %d",
			ErrInvalidLineItem, kind, costPerItem,
		)
	}
	if unitCost < 0 {
		return fmt.Errorf(
			"%w: %s unit cost must not be negative, got %d",
			ErrInvalidLineItem, kind, unitCost,
		)
	}
	return nil
}

// validateLineItemActuals bounds whichever actual fields the caller supplied. Absent
// fields are not defaulted to zero and then rejected: leaving a value unset is
// how a reimbursement draft is filled in over time.
func validateLineItemActuals(in LineItemActualsInput) error {
	if in.ActualQuantity != nil && *in.ActualQuantity <= 0 {
		return fmt.Errorf(
			"%w: actual quantity must be positive, got %d",
			ErrInvalidLineItem, *in.ActualQuantity,
		)
	}
	if in.ActualCostPerItem != nil && *in.ActualCostPerItem < 0 {
		return fmt.Errorf(
			"%w: actual cost per item must not be negative, got %d",
			ErrInvalidLineItem, *in.ActualCostPerItem,
		)
	}
	if in.ActualUnitCost != nil && *in.ActualUnitCost < 0 {
		return fmt.Errorf(
			"%w: actual unit cost must not be negative, got %d",
			ErrInvalidLineItem, *in.ActualUnitCost,
		)
	}
	if in.Description != nil {
		return validateLineItemDescription(*in.Description)
	}
	return nil
}

func validateLineItemDescription(description string) error {
	if len(description) > maxLineItemDescriptionLen {
		return fmt.Errorf(
			"%w: description is %d characters, limit is %d",
			ErrInvalidLineItem, len(description), maxLineItemDescriptionLen,
		)
	}
	return nil
}

// lineItemParamsFrom builds a full update parameter set from a loaded row.
//
// UpdateLineItem writes all twelve columns, so every mutating method has to
// read the current row and write it back whole. Building the params in one
// place is what keeps "change only what was asked for" a property of the
// package rather than something each method has to remember.
func lineItemParamsFrom(item db.LineItem) db.UpdateLineItemParams {
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
