package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"backend/internal/domain"
	"backend/internal/repository"
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

var AllLineItemStatuses = []domain.LineItemStatus{
	domain.LineItemStatusPrDraft,
	domain.LineItemStatusPrPending,
	domain.LineItemStatusPrApproved,
	domain.LineItemStatusPrRejected,
	domain.LineItemStatusReimbDraft,
	domain.LineItemStatusReimbPending,
	domain.LineItemStatusReimbApproved,
	domain.LineItemStatusReimbRejected,
}

// status -> statuses it can move to, empty = terminal
var lineItemTransitions = map[domain.LineItemStatus][]domain.LineItemStatus{
	// purchase request phase
	domain.LineItemStatusPrDraft: {domain.LineItemStatusPrPending},
	domain.LineItemStatusPrPending: {
		domain.LineItemStatusPrDraft, // soft reject, submitter can edit and resubmit
		domain.LineItemStatusPrApproved,
		domain.LineItemStatusPrRejected,
	},
	domain.LineItemStatusPrApproved: {domain.LineItemStatusReimbDraft},
	domain.LineItemStatusPrRejected: {},

	// reimbursement phase
	domain.LineItemStatusReimbDraft: {domain.LineItemStatusReimbPending},
	domain.LineItemStatusReimbPending: {
		domain.LineItemStatusReimbDraft, // soft reject, submitter can edit and resubmit
		domain.LineItemStatusReimbApproved,
		domain.LineItemStatusReimbRejected,
	},
	domain.LineItemStatusReimbApproved: {},
	domain.LineItemStatusReimbRejected: {},
}

func CanTransitionLineItem(from, to domain.LineItemStatus) bool {
	for _, next := range lineItemTransitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

// returns a copy so callers can't mutate the table
func NextLineItemStates(from domain.LineItemStatus) []domain.LineItemStatus {
	next := lineItemTransitions[from]
	out := make([]domain.LineItemStatus, len(next))
	copy(out, next)
	return out
}

// what Transition accepts for this item, unlike NextLineItemStates which
// only knows the status. reimb_pending is still offered without actuals,
// that's a 400 for missing input, not a move the item can't make
func AllowedLineItemTransitions(item domain.LineItem) []domain.LineItemStatus {
	out := []domain.LineItemStatus{}
	if !item.IsActive {
		return out
	}
	for _, next := range lineItemTransitions[item.Status] {
		if !missingReimbursementRequest(item, next) {
			out = append(out, next)
		}
	}
	return out
}

// pr_approved -> reimb_draft goes through AttachReimbursementRequest,
// which sets the reimbursement id in the same write
func missingReimbursementRequest(
	item domain.LineItem,
	to domain.LineItemStatus,
) bool {
	return IsLineItemReimbursementPhase(to) && item.ReimbursementRequestID == nil
}

// unknown statuses are invalid, not terminal
func IsTerminalLineItemStatus(s domain.LineItemStatus) bool {
	next, known := lineItemTransitions[s]
	return known && len(next) == 0
}

func IsLineItemReimbursementPhase(s domain.LineItemStatus) bool {
	switch s {
	case domain.LineItemStatusReimbDraft,
		domain.LineItemStatusReimbPending,
		domain.LineItemStatusReimbApproved,
		domain.LineItemStatusReimbRejected:
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
	Create(ctx context.Context, in CreateLineItemInput) (domain.LineItem, error)
	GetByID(ctx context.Context, id int64) (domain.LineItem, error)
	ListByPurchaseRequest(
		ctx context.Context,
		prID int64,
	) ([]domain.LineItem, error)
	ListByReimbursementRequest(
		ctx context.Context,
		reimbID int64,
	) ([]domain.LineItem, error)
	TotalsByPurchaseRequest(ctx context.Context, prID int64) (LineItemTotals, error)

	UpdateEstimates(
		ctx context.Context,
		id int64,
		in LineItemEstimatesInput,
	) (domain.LineItem, error)
	UpdateActuals(
		ctx context.Context,
		id int64,
		in LineItemActualsInput,
	) (domain.LineItem, error)
	AttachReimbursementRequest(
		ctx context.Context,
		id, reimbID int64,
	) (domain.LineItem, error)
	Transition(
		ctx context.Context,
		id int64,
		to domain.LineItemStatus,
	) (domain.LineItem, error)
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
) (domain.LineItem, error) {
	if in.PurchaseRequestID <= 0 {
		return domain.LineItem{}, fmt.Errorf(
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
		return domain.LineItem{}, err
	}
	if err = validateLineItemDescription(in.Description); err != nil {
		return domain.LineItem{}, err
	}

	return s.repo.CreateLineItem(ctx, repository.CreateLineItemParams{
		EstimatedQuantity:    in.EstimatedQuantity,
		EstimatedCostPerItem: in.EstimatedCostPerItem,
		EstimatedUnitCost:    in.EstimatedUnitCost,
		PrDescription:        descriptionPtr(in.Description),
		PurchaseRequestID:    in.PurchaseRequestID,
	})
}

// includes archived items
func (s *lineItemServiceImpl) GetByID(
	ctx context.Context,
	id int64,
) (domain.LineItem, error) {
	item, err := s.repo.GetLineItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrLineItemNotFound) {
			return domain.LineItem{}, fmt.Errorf(
				"%w: line item %d", ErrLineItemNotFound, id,
			)
		}
		return domain.LineItem{}, err
	}
	return item, nil
}

// archived items are read-only, archivedErr is the error callers report it as
func (s *lineItemServiceImpl) getActive(
	ctx context.Context,
	id int64,
	archivedErr error,
) (domain.LineItem, error) {
	item, err := s.GetByID(ctx, id)
	if err != nil {
		return domain.LineItem{}, err
	}
	if !item.IsActive {
		return domain.LineItem{}, fmt.Errorf(
			"%w: line item %d is archived", archivedErr, id,
		)
	}
	return item, nil
}

func (s *lineItemServiceImpl) ListByPurchaseRequest(
	ctx context.Context,
	prID int64,
) ([]domain.LineItem, error) {
	return s.repo.GetLineItemsByPRID(ctx, prID)
}

func (s *lineItemServiceImpl) ListByReimbursementRequest(
	ctx context.Context,
	reimbID int64,
) ([]domain.LineItem, error) {
	return s.repo.GetLineItemsByReimbID(ctx, reimbID)
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
func LineItemEstimatedTotal(item domain.LineItem) int64 {
	return item.EstimatedQuantity * item.EstimatedCostPerItem
}

// false until both actual quantity and cost per item are recorded
func LineItemActualTotal(item domain.LineItem) (int64, bool) {
	if item.ActualQuantity == nil || item.ActualCostPerItem == nil {
		return 0, false
	}
	return *item.ActualQuantity * *item.ActualCostPerItem, true
}

// only in pr_draft, so approvers never see numbers change under them
func (s *lineItemServiceImpl) UpdateEstimates(
	ctx context.Context,
	id int64,
	in LineItemEstimatesInput,
) (domain.LineItem, error) {
	item, err := s.getActive(ctx, id, ErrImmutableLineItemField)
	if err != nil {
		return domain.LineItem{}, err
	}
	if item.Status != domain.LineItemStatusPrDraft {
		return domain.LineItem{}, fmt.Errorf(
			"%w: estimates are editable only in %q, line item %d is %q",
			ErrImmutableLineItemField, domain.LineItemStatusPrDraft, id, item.Status,
		)
	}
	err = validateLineItemAmounts(
		"estimated",
		in.EstimatedQuantity,
		in.EstimatedCostPerItem,
		in.EstimatedUnitCost,
	)
	if err != nil {
		return domain.LineItem{}, err
	}
	if err = validateLineItemDescription(in.Description); err != nil {
		return domain.LineItem{}, err
	}

	item.EstimatedQuantity = in.EstimatedQuantity
	item.EstimatedCostPerItem = in.EstimatedCostPerItem
	item.EstimatedUnitCost = in.EstimatedUnitCost
	item.PrDescription = descriptionPtr(in.Description)
	return s.update(ctx, item)
}

// only in reimb_draft, frozen once submitted
func (s *lineItemServiceImpl) UpdateActuals(
	ctx context.Context,
	id int64,
	in LineItemActualsInput,
) (domain.LineItem, error) {
	item, err := s.getActive(ctx, id, ErrImmutableLineItemField)
	if err != nil {
		return domain.LineItem{}, err
	}
	if item.Status != domain.LineItemStatusReimbDraft {
		return domain.LineItem{}, fmt.Errorf(
			"%w: actuals are writable only in %q, line item %d is %q",
			ErrImmutableLineItemField, domain.LineItemStatusReimbDraft, id, item.Status,
		)
	}
	if err = validateLineItemActuals(in); err != nil {
		return domain.LineItem{}, err
	}

	if in.ActualQuantity != nil {
		item.ActualQuantity = in.ActualQuantity
	}
	if in.ActualCostPerItem != nil {
		item.ActualCostPerItem = in.ActualCostPerItem
	}
	if in.ActualUnitCost != nil {
		item.ActualUnitCost = in.ActualUnitCost
	}
	if in.Description != nil {
		item.ReimbDescription = descriptionPtr(*in.Description)
	}
	return s.update(ctx, item)
}

// links the reimbursement and moves to reimb_draft in one write
func (s *lineItemServiceImpl) AttachReimbursementRequest(
	ctx context.Context,
	id, reimbID int64,
) (domain.LineItem, error) {
	if reimbID <= 0 {
		return domain.LineItem{}, fmt.Errorf(
			"%w: reimbursement request id must be positive, got %d",
			ErrInvalidLineItem, reimbID,
		)
	}
	item, err := s.getActive(ctx, id, ErrInvalidLineItemTransition)
	if err != nil {
		return domain.LineItem{}, err
	}
	// only pr_approved. reimb_pending -> reimb_draft is a soft reject, not a reattach
	if item.Status != domain.LineItemStatusPrApproved {
		return domain.LineItem{}, fmt.Errorf(
			"%w: line item %d is %q, cannot enter the reimbursement phase",
			ErrInvalidLineItemTransition, id, item.Status,
		)
	}

	item.ReimbursementRequestID = &reimbID
	item.Status = domain.LineItemStatusReimbDraft
	return s.update(ctx, item)
}

// checks on the read for clear errors, the guarded write catches concurrent changes
func (s *lineItemServiceImpl) Transition(
	ctx context.Context,
	id int64,
	to domain.LineItemStatus,
) (domain.LineItem, error) {
	if _, known := lineItemTransitions[to]; !known {
		return domain.LineItem{}, fmt.Errorf(
			"%w: unknown status %q", ErrInvalidLineItem, to,
		)
	}
	item, err := s.getActive(ctx, id, ErrInvalidLineItemTransition)
	if err != nil {
		return domain.LineItem{}, err
	}

	if !CanTransitionLineItem(item.Status, to) {
		return domain.LineItem{}, fmt.Errorf(
			"%w: line item %d cannot move from %q to %q",
			ErrInvalidLineItemTransition, id, item.Status, to,
		)
	}
	if missingReimbursementRequest(item, to) {
		return domain.LineItem{}, fmt.Errorf(
			"%w: line item %d has no reimbursement request, attach one to move to %q",
			ErrInvalidLineItemTransition, id, to,
		)
	}
	if to == domain.LineItemStatusReimbPending {
		if err = requireLineItemActuals(item); err != nil {
			return domain.LineItem{}, err
		}
	}

	updated, err := s.repo.TransitionLineItem(ctx, id, item.Status, to)
	// no row = someone changed or archived it after our read
	if errors.Is(err, repository.ErrLineItemConflict) {
		return domain.LineItem{}, fmt.Errorf(
			"%w: line item %d changed while moving from %q to %q, reload and retry",
			ErrInvalidLineItemTransition, id, item.Status, to,
		)
	}
	return updated, err
}

// can't submit a reimbursement with empty actuals
func requireLineItemActuals(item domain.LineItem) error {
	var missing []string
	if item.ActualQuantity == nil {
		missing = append(missing, "quantity")
	}
	if item.ActualCostPerItem == nil {
		missing = append(missing, "cost per item")
	}
	if item.ActualUnitCost == nil {
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

// UpdateLineItem writes every column from the item it is given, and is guarded
// on UpdatedAt, so the item must be the one that was just read
func (s *lineItemServiceImpl) update(
	ctx context.Context,
	item domain.LineItem,
) (domain.LineItem, error) {
	updated, err := s.repo.UpdateLineItem(ctx, item)
	if errors.Is(err, repository.ErrLineItemConflict) {
		return domain.LineItem{}, fmt.Errorf(
			"%w: line item %d changed since it was read, reload and retry",
			ErrLineItemConflict, item.ID,
		)
	}
	return updated, err
}

// empty string -> NULL
func descriptionPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
