// Package lineitem is the Postgres data access for line items. It mirrors the
// sqlc-generated queries one for one and holds no rules of its own.
//
// The interface this satisfies is declared by its consumer, in
// backend/internal/service/lineitem, so the data layer depends on nothing
// above it. Whether the two still agree is settled at compile time where they
// are assembled, in backend/internal/app.
package lineitem

import (
	"context"

	"backend/db"

	"github.com/jackc/pgx/v5/pgtype"
)

// Repository reads and writes line items.
type Repository struct {
	queries *db.Queries
}

// New returns a Repository backed by the given queries, which are built with
// db.New(pool).
func New(queries *db.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) GetLineItemByID(
	ctx context.Context,
	id int64,
) (db.LineItem, error) {
	return r.queries.GetLineItemByID(ctx, id)
}

func (r *Repository) GetLineItemsByPRID(
	ctx context.Context,
	prID int64,
) ([]db.LineItem, error) {
	return r.queries.GetLineItemsByPRID(ctx, prID)
}

func (r *Repository) GetLineItemsByReimbID(
	ctx context.Context,
	reimbID pgtype.Int8,
) ([]db.LineItem, error) {
	return r.queries.GetLineItemsByReimbID(ctx, reimbID)
}

func (r *Repository) ListLineItems(
	ctx context.Context,
	limit int32,
) ([]db.LineItem, error) {
	return r.queries.ListLineItems(ctx, limit)
}

func (r *Repository) CreateLineItem(
	ctx context.Context,
	arg db.CreateLineItemParams,
) (db.LineItem, error) {
	return r.queries.CreateLineItem(ctx, arg)
}

func (r *Repository) UpdateLineItem(
	ctx context.Context,
	arg db.UpdateLineItemParams,
) (db.LineItem, error) {
	return r.queries.UpdateLineItem(ctx, arg)
}

// DeleteLineItem removes a row outright. The service does not call it:
// removal is a soft delete that clears is_active, so a line item referenced by
// a purchase request never vanishes from underneath it. It is kept to mirror
// the generated queries.
func (r *Repository) DeleteLineItem(ctx context.Context, id int64) error {
	return r.queries.DeleteLineItem(ctx, id)
}
