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

// handler maps these to status codes via errors.Is
var (
	ErrLineItemNotFound          = errors.New("line item not found")       // 404
	ErrInvalidLineItem           = errors.New("invalid line item")         // 400
	ErrInvalidLineItemTransition = errors.New("invalid status transition") // 409
	ErrImmutableLineItemField    = errors.New(                             // 409
		"field is not editable in the current status",
	)
	ErrLineItemConflict = errors.New( // 409
		"line item was modified concurrently",
	)
)

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

// status -> statuses it can move to, empty = terminal
var lineItemTransitions = map[db.LineItemStatus][]db.LineItemStatus{
	// purchase request phase
	db.LineItemStatusPrDraft: {db.LineItemStatusPrPending},
	db.LineItemStatusPrPending: {
		db.LineItemStatusPrDraft, // soft reject, submitter can edit and resubmit
		db.LineItemStatusPrApproved,
		db.LineItemStatusPrRejected,
	},
	db.LineItemStatusPrApproved: {db.LineItemStatusReimbDraft},
	db.LineItemStatusPrRejected: {},

	// reimbursement phase
	db.LineItemStatusReimbDraft: {db.LineItemStatusReimbPending},
	db.LineItemStatusReimbPending: {
		db.LineItemStatusReimbDraft, // soft reject, submitter can edit and resubmit
		db.LineItemStatusReimbApproved,
		db.LineItemStatusReimbRejected,
	},
	db.LineItemStatusReimbApproved: {},
	db.LineItemStatusReimbRejected: {},
}

func CanTransitionLineItem(from, to db.LineItemStatus) bool {
	for _, next := range lineItemTransitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

// returns a copy so callers can't mutate the table
func NextLineItemStates(from db.LineItemStatus) []db.LineItemStatus {
	next := lineItemTransitions[from]
	out := make([]db.LineItemStatus, len(next))
	copy(out, next)
	return out
}

// unknown statuses are invalid, not terminal
func IsTerminalLineItemStatus(s db.LineItemStatus) bool {
	next, known := lineItemTransitions[s]
	return known && len(next) == 0
}

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

// columns are unbounded TEXT
const maxLineItemDescriptionLen = 1000

// keeps quantity * cost, and sums of thousands of those, inside int64
const (
	maxLineItemQuantity = 1_000_000
	maxLineItemCents    = 1_000_000_000 // $10M
)

// no Status: new items always start as pr_draft (column default)
type CreateLineItemInput struct {
	PurchaseRequestID    int64
	EstimatedQuantity    int64
	EstimatedCostPerItem int64
	EstimatedUnitCost    int64
	Description          string
}

type LineItemEstimatesInput struct {
	EstimatedQuantity    int64
	EstimatedCostPerItem int64
	EstimatedUnitCost    int64
	Description          string
}

// all optional, nil = leave unchanged (filled in as receipts arrive)
type LineItemActualsInput struct {
	ActualQuantity    *int64
	ActualCostPerItem *int64
	ActualUnitCost    *int64
	Description       *string
}

// in cents, active items only
type LineItemTotals struct {
	EstimatedTotal int64
	ActualTotal    int64
}

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

func NewLineItemService(repo repository.LineItemRepository) LineItemService {
	return &lineItemServiceImpl{repo: repo}
}

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

// includes archived items
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

// archived items are read-only, archivedErr is the error callers report it as
func (s *lineItemServiceImpl) getActive(
	ctx context.Context,
	id int64,
	archivedErr error,
) (db.LineItem, error) {
	item, err := s.GetByID(ctx, id)
	if err != nil {
		return db.LineItem{}, err
	}
	if !item.IsActive {
		return db.LineItem{}, fmt.Errorf(
			"%w: line item %d is archived", archivedErr, id,
		)
	}
	return item, nil
}

func (s *lineItemServiceImpl) ListByPurchaseRequest(
	ctx context.Context,
	prID int64,
) ([]db.LineItem, error) {
	return s.repo.GetLineItemsByPRID(ctx, prID)
}

func (s *lineItemServiceImpl) ListByReimbursementRequest(
	ctx context.Context,
	reimbID int64,
) ([]db.LineItem, error) {
	return s.repo.GetLineItemsByReimbID(ctx, int8From(&reimbID))
}

// actual total only counts items with both actual quantity and cost set
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
		totals.EstimatedTotal += LineItemEstimatedTotal(item)
		if actual, ok := LineItemActualTotal(item); ok {
			totals.ActualTotal += actual
		}
	}
	return totals, nil
}

// in cents, quantity * cost per item
func LineItemEstimatedTotal(item db.LineItem) int64 {
	return item.EstimatedQuantity * item.EstimatedCostPerItem
}

// false until both actual quantity and cost per item are recorded
func LineItemActualTotal(item db.LineItem) (int64, bool) {
	if !item.ActualQuantity.Valid || !item.ActualCostPerItem.Valid {
		return 0, false
	}
	return item.ActualQuantity.Int64 * item.ActualCostPerItem.Int64, true
}

// only in pr_draft, so approvers never see numbers change under them
func (s *lineItemServiceImpl) UpdateEstimates(
	ctx context.Context,
	id int64,
	in LineItemEstimatesInput,
) (db.LineItem, error) {
	item, err := s.getActive(ctx, id, ErrImmutableLineItemField)
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
	return s.update(ctx, params)
}

// only in reimb_draft, frozen once submitted
func (s *lineItemServiceImpl) UpdateActuals(
	ctx context.Context,
	id int64,
	in LineItemActualsInput,
) (db.LineItem, error) {
	item, err := s.getActive(ctx, id, ErrImmutableLineItemField)
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
	return s.update(ctx, params)
}

// links the reimbursement and moves to reimb_draft in one write
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
	item, err := s.getActive(ctx, id, ErrInvalidLineItemTransition)
	if err != nil {
		return db.LineItem{}, err
	}
	// only pr_approved. reimb_pending -> reimb_draft is a soft reject, not a reattach
	if item.Status != db.LineItemStatusPrApproved {
		return db.LineItem{}, fmt.Errorf(
			"%w: line item %d is %q, cannot enter the reimbursement phase",
			ErrInvalidLineItemTransition, id, item.Status,
		)
	}

	params := lineItemParamsFrom(item)
	params.ReimbursementRequestID = int8From(&reimbID)
	params.Status = db.LineItemStatusReimbDraft
	return s.update(ctx, params)
}

// checks on the read for clear errors, the guarded write catches concurrent changes
func (s *lineItemServiceImpl) Transition(
	ctx context.Context,
	id int64,
	to db.LineItemStatus,
) (db.LineItem, error) {
	if _, known := lineItemTransitions[to]; !known {
		return db.LineItem{}, fmt.Errorf(
			"%w: unknown status %q", ErrInvalidLineItem, to,
		)
	}
	item, err := s.getActive(ctx, id, ErrInvalidLineItemTransition)
	if err != nil {
		return db.LineItem{}, err
	}

	if !CanTransitionLineItem(item.Status, to) {
		return db.LineItem{}, fmt.Errorf(
			"%w: line item %d cannot move from %q to %q",
			ErrInvalidLineItemTransition, id, item.Status, to,
		)
	}
	// pr_approved -> reimb_draft goes through AttachReimbursementRequest,
	// which sets the reimbursement id in the same write
	if IsLineItemReimbursementPhase(to) && !item.ReimbursementRequestID.Valid {
		return db.LineItem{}, fmt.Errorf(
			"%w: line item %d has no reimbursement request, attach one to move to %q",
			ErrInvalidLineItemTransition, id, to,
		)
	}
	if to == db.LineItemStatusReimbPending {
		if err = requireLineItemActuals(item); err != nil {
			return db.LineItem{}, err
		}
	}

	updated, err := s.repo.TransitionLineItem(ctx, db.TransitionLineItemParams{
		ID:         id,
		FromStatus: item.Status,
		ToStatus:   to,
	})
	// no row = someone changed or archived it after our read
	if errors.Is(err, pgx.ErrNoRows) {
		return db.LineItem{}, fmt.Errorf(
			"%w: line item %d changed while moving from %q to %q, reload and retry",
			ErrInvalidLineItemTransition, id, item.Status, to,
		)
	}
	return updated, err
}

// can't submit a reimbursement with empty actuals
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

// soft delete, keeps status, archiving twice is a no-op
func (s *lineItemServiceImpl) Archive(ctx context.Context, id int64) error {
	item, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !item.IsActive {
		return nil
	}
	return s.repo.ArchiveLineItem(ctx, id)
}

// kind is "estimated" or "actual"
func validateLineItemAmounts(kind string, quantity, costPerItem, unitCost int64) error {
	return errors.Join(
		checkLineItemAmount(kind+" quantity", quantity, 1, maxLineItemQuantity),
		checkLineItemAmount(kind+" cost per item", costPerItem, 0, maxLineItemCents),
		checkLineItemAmount(kind+" unit cost", unitCost, 0, maxLineItemCents),
	)
}

// only checks fields that were sent
func validateLineItemActuals(in LineItemActualsInput) error {
	var errs []error
	if in.ActualQuantity != nil {
		errs = append(errs, checkLineItemAmount(
			"actual quantity", *in.ActualQuantity, 1, maxLineItemQuantity,
		))
	}
	if in.ActualCostPerItem != nil {
		errs = append(errs, checkLineItemAmount(
			"actual cost per item", *in.ActualCostPerItem, 0, maxLineItemCents,
		))
	}
	if in.ActualUnitCost != nil {
		errs = append(errs, checkLineItemAmount(
			"actual unit cost", *in.ActualUnitCost, 0, maxLineItemCents,
		))
	}
	if in.Description != nil {
		errs = append(errs, validateLineItemDescription(*in.Description))
	}
	return errors.Join(errs...)
}

func checkLineItemAmount(name string, v, lo, hi int64) error {
	if v < lo || v > hi {
		return fmt.Errorf(
			"%w: %s must be between %d and %d, got %d",
			ErrInvalidLineItem, name, lo, hi, v,
		)
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

// UpdateLineItem writes every column, so start from the current row.
// UpdatedAt is the read version the write is guarded on
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
		UpdatedAt:              item.UpdatedAt,
	}
}

// guarded on updated_at, no row = someone wrote it after our read
func (s *lineItemServiceImpl) update(
	ctx context.Context,
	params db.UpdateLineItemParams,
) (db.LineItem, error) {
	updated, err := s.repo.UpdateLineItem(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return db.LineItem{}, fmt.Errorf(
			"%w: line item %d changed since it was read, reload and retry",
			ErrLineItemConflict, params.ID,
		)
	}
	return updated, err
}

// empty string -> NULL
func textFrom(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func int8From(v *int64) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *v, Valid: true}
}
