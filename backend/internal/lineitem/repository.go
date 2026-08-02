package lineitem

import (
	"context"

	"backend/db"

	"github.com/jackc/pgx/v5/pgtype"
)

// Repository is the data access seam for line items. It mirrors the
// sqlc-generated queries one for one and holds no logic of its own.
//
// It exists so the service can be exercised against an in-memory fake, which
// is what lets the lifecycle rules be tested without a database.
type Repository interface {
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

	// DeleteLineItem removes a row outright. The service does not call it:
	// removal is a soft delete that clears is_active, so that a line item
	// referenced by a purchase request never vanishes from underneath it.
	// It is kept here to mirror the generated queries.
	DeleteLineItem(ctx context.Context, id int64) error
}

// sqlRepository is the Postgres-backed Repository, delegating to the
// sqlc-generated queries.
type sqlRepository struct {
	queries *db.Queries
}

// Assert at compile time that the real implementation still satisfies the
// interface, so a regenerated query with a changed signature fails the build
// here rather than at the call site.
var _ Repository = (*sqlRepository)(nil)

// NewRepository returns a Repository backed by the given queries, which are
// built with db.New(pool).
func NewRepository(queries *db.Queries) Repository {
	return &sqlRepository{queries: queries}
}

func (r *sqlRepository) GetLineItemByID(
	ctx context.Context,
	id int64,
) (db.LineItem, error) {
	return r.queries.GetLineItemByID(ctx, id)
}

func (r *sqlRepository) GetLineItemsByPRID(
	ctx context.Context,
	prID int64,
) ([]db.LineItem, error) {
	return r.queries.GetLineItemsByPRID(ctx, prID)
}

func (r *sqlRepository) GetLineItemsByReimbID(
	ctx context.Context,
	reimbID pgtype.Int8,
) ([]db.LineItem, error) {
	return r.queries.GetLineItemsByReimbID(ctx, reimbID)
}

func (r *sqlRepository) ListLineItems(
	ctx context.Context,
	limit int32,
) ([]db.LineItem, error) {
	return r.queries.ListLineItems(ctx, limit)
}

func (r *sqlRepository) CreateLineItem(
	ctx context.Context,
	arg db.CreateLineItemParams,
) (db.LineItem, error) {
	return r.queries.CreateLineItem(ctx, arg)
}

func (r *sqlRepository) UpdateLineItem(
	ctx context.Context,
	arg db.UpdateLineItemParams,
) (db.LineItem, error) {
	return r.queries.UpdateLineItem(ctx, arg)
}

func (r *sqlRepository) DeleteLineItem(ctx context.Context, id int64) error {
	return r.queries.DeleteLineItem(ctx, id)
}
