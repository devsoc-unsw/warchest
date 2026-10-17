// Postgres data access for line items. It mirrors the sqlc-generated queries
// one for one and holds no rules of its own.

package repository

import (
	"context"

	"backend/db"

	"github.com/jackc/pgx/v5/pgtype"
)

// LineItemRepository is the data access the line item service needs. Keeping
// it an interface is what lets the service rules be tested against an
// in-memory fake with no database.
type LineItemRepository interface {
	GetLineItemByID(ctx context.Context, id int64) (db.LineItem, error)
	GetLineItemsByPRID(ctx context.Context, prID int64) ([]db.LineItem, error)
	GetLineItemsByReimbID(
		ctx context.Context,
		reimbID pgtype.Int8,
	) ([]db.LineItem, error)
	ListLineItems(ctx context.Context, limit int32) ([]db.LineItem, error)
	CreateLineItem(
		ctx context.Context,
		arg db.CreateLineItemParams,
	) (db.LineItem, error)
	UpdateLineItem(
		ctx context.Context,
		arg db.UpdateLineItemParams,
	) (db.LineItem, error)
	ArchiveLineItem(ctx context.Context, id int64) error
}

type lineItemRepositoryImpl struct {
	queries *db.Queries
}

// NewLineItemRepository returns a LineItemRepository backed by the given
// queries, which are built with db.New(pool).
func NewLineItemRepository(queries *db.Queries) LineItemRepository {
	return &lineItemRepositoryImpl{queries: queries}
}

func (r *lineItemRepositoryImpl) GetLineItemByID(
	ctx context.Context,
	id int64,
) (db.LineItem, error) {
	return r.queries.GetLineItemByID(ctx, id)
}

func (r *lineItemRepositoryImpl) GetLineItemsByPRID(
	ctx context.Context,
	prID int64,
) ([]db.LineItem, error) {
	return r.queries.GetLineItemsByPRID(ctx, prID)
}

func (r *lineItemRepositoryImpl) GetLineItemsByReimbID(
	ctx context.Context,
	reimbID pgtype.Int8,
) ([]db.LineItem, error) {
	return r.queries.GetLineItemsByReimbID(ctx, reimbID)
}

func (r *lineItemRepositoryImpl) ListLineItems(
	ctx context.Context,
	limit int32,
) ([]db.LineItem, error) {
	return r.queries.ListLineItems(ctx, limit)
}

func (r *lineItemRepositoryImpl) CreateLineItem(
	ctx context.Context,
	arg db.CreateLineItemParams,
) (db.LineItem, error) {
	return r.queries.CreateLineItem(ctx, arg)
}

func (r *lineItemRepositoryImpl) UpdateLineItem(
	ctx context.Context,
	arg db.UpdateLineItemParams,
) (db.LineItem, error) {
	return r.queries.UpdateLineItem(ctx, arg)
}

// ArchiveLineItem soft deletes a row by clearing is_active, so a line item
// referenced by a purchase request never vanishes from underneath it. The
// service archives through UpdateLineItem instead; this is kept to mirror the
// generated queries.
func (r *lineItemRepositoryImpl) ArchiveLineItem(
	ctx context.Context,
	id int64,
) error {
	return r.queries.ArchiveLineItem(ctx, id)
}
