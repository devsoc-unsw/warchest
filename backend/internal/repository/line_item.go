// thin wrapper over sqlc queries, no rules here

package repository

import (
	"context"

	"backend/db"

	"github.com/jackc/pgx/v5/pgtype"
)

// interface so service tests can use a fake
type LineItemRepository interface {
	GetLineItemByID(ctx context.Context, id int64) (db.LineItem, error)
	GetLineItemsByPRID(ctx context.Context, prID int64) ([]db.LineItem, error)
	GetLineItemsByReimbID(
		ctx context.Context,
		reimbID pgtype.Int8,
	) ([]db.LineItem, error)
	CreateLineItem(
		ctx context.Context,
		arg db.CreateLineItemParams,
	) (db.LineItem, error)
	UpdateLineItem(
		ctx context.Context,
		arg db.UpdateLineItemParams,
	) (db.LineItem, error)
	TransitionLineItem(
		ctx context.Context,
		arg db.TransitionLineItemParams,
	) (db.LineItem, error)
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

func (r *lineItemRepositoryImpl) ArchiveLineItem(
	ctx context.Context,
	id int64,
) error {
	return r.queries.ArchiveLineItem(ctx, id)
}

func (r *lineItemRepositoryImpl) TransitionLineItem(
	ctx context.Context,
	arg db.TransitionLineItemParams,
) (db.LineItem, error) {
	return r.queries.TransitionLineItem(ctx, arg)
}
