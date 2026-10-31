// thin wrapper over sqlc queries, maps rows to the domain model, no rules here

package repository

import (
	"context"
	"errors"
	"time"

	"backend/db"
	"backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// the service maps these onto its own errors, pgx.ErrNoRows stops here
var (
	ErrLineItemNotFound = errors.New("line item row not found")
	// a guarded write matched no row, the item changed since it was read
	ErrLineItemConflict = errors.New("line item changed since it was read")
)

// only the columns the insert writes, the rest come from column defaults
type CreateLineItemParams struct {
	PurchaseRequestID    int64
	EstimatedQuantity    int64
	EstimatedCostPerItem int64
	EstimatedUnitCost    int64
	PrDescription        *string
}

// interface so service tests can use a fake
type LineItemRepository interface {
	GetLineItemByID(ctx context.Context, id int64) (domain.LineItem, error)
	GetLineItemsByPRID(ctx context.Context, prID int64) ([]domain.LineItem, error)
	GetLineItemsByReimbID(
		ctx context.Context,
		reimbID int64,
	) ([]domain.LineItem, error)
	CreateLineItem(
		ctx context.Context,
		arg CreateLineItemParams,
	) (domain.LineItem, error)
	// writes every column, guarded on item.UpdatedAt
	UpdateLineItem(
		ctx context.Context,
		item domain.LineItem,
	) (domain.LineItem, error)
	// guarded on from, archived rows never match
	TransitionLineItem(
		ctx context.Context,
		id int64,
		from, to domain.LineItemStatus,
	) (domain.LineItem, error)
	ArchiveLineItem(ctx context.Context, id int64) error
}

type lineItemRepositoryImpl struct {
	queries *db.Queries
}

func NewLineItemRepository(queries *db.Queries) LineItemRepository {
	return &lineItemRepositoryImpl{queries: queries}
}

func (r *lineItemRepositoryImpl) GetLineItemByID(
	ctx context.Context,
	id int64,
) (domain.LineItem, error) {
	row, err := r.queries.GetLineItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.LineItem{}, ErrLineItemNotFound
		}
		return domain.LineItem{}, err
	}
	return lineItemFromRow(row), nil
}

func (r *lineItemRepositoryImpl) GetLineItemsByPRID(
	ctx context.Context,
	prID int64,
) ([]domain.LineItem, error) {
	rows, err := r.queries.GetLineItemsByPRID(ctx, prID)
	if err != nil {
		return nil, err
	}
	return lineItemsFromRows(rows), nil
}

func (r *lineItemRepositoryImpl) GetLineItemsByReimbID(
	ctx context.Context,
	reimbID int64,
) ([]domain.LineItem, error) {
	// the column is nullable but the lookup never matches NULL
	rows, err := r.queries.GetLineItemsByReimbID(ctx, int8From(&reimbID))
	if err != nil {
		return nil, err
	}
	return lineItemsFromRows(rows), nil
}

func (r *lineItemRepositoryImpl) CreateLineItem(
	ctx context.Context,
	arg CreateLineItemParams,
) (domain.LineItem, error) {
	row, err := r.queries.CreateLineItem(ctx, db.CreateLineItemParams{
		EstimatedQuantity:    arg.EstimatedQuantity,
		EstimatedCostPerItem: arg.EstimatedCostPerItem,
		EstimatedUnitCost:    arg.EstimatedUnitCost,
		PrDescription:        textFrom(arg.PrDescription),
		PurchaseRequestID:    arg.PurchaseRequestID,
	})
	if err != nil {
		return domain.LineItem{}, err
	}
	return lineItemFromRow(row), nil
}

// the query sets every column, so the whole item is the argument.
// purchase_request_ID and created_at are not in the SET list and stay put
func (r *lineItemRepositoryImpl) UpdateLineItem(
	ctx context.Context,
	item domain.LineItem,
) (domain.LineItem, error) {
	row, err := r.queries.UpdateLineItem(ctx, db.UpdateLineItemParams{
		ID:                     item.ID,
		EstimatedQuantity:      item.EstimatedQuantity,
		EstimatedCostPerItem:   item.EstimatedCostPerItem,
		EstimatedUnitCost:      item.EstimatedUnitCost,
		PrDescription:          textFrom(item.PrDescription),
		ActualQuantity:         int8From(item.ActualQuantity),
		ActualCostPerItem:      int8From(item.ActualCostPerItem),
		ActualUnitCost:         int8From(item.ActualUnitCost),
		ReimbDescription:       textFrom(item.ReimbDescription),
		ReimbursementRequestID: int8From(item.ReimbursementRequestID),
		Status:                 db.LineItemStatus(item.Status),
		IsActive:               item.IsActive,
		UpdatedAt:              timestamptzFrom(item.UpdatedAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.LineItem{}, ErrLineItemConflict
		}
		return domain.LineItem{}, err
	}
	return lineItemFromRow(row), nil
}

func (r *lineItemRepositoryImpl) TransitionLineItem(
	ctx context.Context,
	id int64,
	from, to domain.LineItemStatus,
) (domain.LineItem, error) {
	row, err := r.queries.TransitionLineItem(ctx, db.TransitionLineItemParams{
		ID:         id,
		FromStatus: db.LineItemStatus(from),
		ToStatus:   db.LineItemStatus(to),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.LineItem{}, ErrLineItemConflict
		}
		return domain.LineItem{}, err
	}
	return lineItemFromRow(row), nil
}

func (r *lineItemRepositoryImpl) ArchiveLineItem(
	ctx context.Context,
	id int64,
) error {
	return r.queries.ArchiveLineItem(ctx, id)
}

// pgtype goes no further than this file, everything above speaks domain types

func lineItemFromRow(row db.LineItem) domain.LineItem {
	return domain.LineItem{
		ID:                     row.ID,
		PurchaseRequestID:      row.PurchaseRequestID,
		ReimbursementRequestID: int64Ptr(row.ReimbursementRequestID),
		Status:                 domain.LineItemStatus(row.Status),
		IsActive:               row.IsActive,

		EstimatedQuantity:    row.EstimatedQuantity,
		EstimatedCostPerItem: row.EstimatedCostPerItem,
		EstimatedUnitCost:    row.EstimatedUnitCost,
		PrDescription:        stringPtr(row.PrDescription),

		ActualQuantity:    int64Ptr(row.ActualQuantity),
		ActualCostPerItem: int64Ptr(row.ActualCostPerItem),
		ActualUnitCost:    int64Ptr(row.ActualUnitCost),
		ReimbDescription:  stringPtr(row.ReimbDescription),

		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

// empty rather than nil, callers range over it
func lineItemsFromRows(rows []db.LineItem) []domain.LineItem {
	out := make([]domain.LineItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, lineItemFromRow(row))
	}
	return out
}

func int64Ptr(v pgtype.Int8) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

func stringPtr(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
}

// nil -> NULL
func int8From(v *int64) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *v, Valid: true}
}

// nil -> NULL
func textFrom(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *v, Valid: true}
}

// created_at and updated_at are NOT NULL, so every time.Time is a real value
func timestamptzFrom(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
