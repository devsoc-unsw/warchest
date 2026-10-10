// talks directly to the database, forwarded to db.Queries
// this layer separates data access from business logic

package repository

import (
	"backend/db"
	"context"
)

// what methods do we have for purchase requests
type PurchaseRequestRepository interface {
	CreatePR(ctx context.Context, params db.CreatePRParams) (db.PurchaseRequest, error)
	GetPR(ctx context.Context, id int64) (db.PurchaseRequest, error)
	ListPR(ctx context.Context, params db.ListPRParams) ([]db.PurchaseRequest, error)
	UpdatePR(ctx context.Context, params db.UpdatePRParams) (db.PurchaseRequest, error)
	DeletePR(ctx context.Context, id int64) error
}

// define purchaseRequestRepositoryImpl struct
type purchaseRequestRepositoryImpl struct {
	queries *db.Queries
}

// constructor
func NewPurchaseRequestRepository(
	queries *db.Queries,
) PurchaseRequestRepository {
	return &purchaseRequestRepositoryImpl{
		queries: queries,
	}
}

// create
func (r *purchaseRequestRepositoryImpl) CreatePR(
	ctx context.Context,
	params db.CreatePRParams,
) (db.PurchaseRequest, error) {
	return r.queries.CreatePR(ctx, params)
}

// get one
func (r *purchaseRequestRepositoryImpl) GetPR(
	ctx context.Context,
	id int64,
) (db.PurchaseRequest, error) {
	return r.queries.GetPR(ctx, id)
}

// list
func (r *purchaseRequestRepositoryImpl) ListPR(
	ctx context.Context,
	params db.ListPRParams,
) ([]db.PurchaseRequest, error) {
	return r.queries.ListPR(ctx, params)
}

// update
func (r *purchaseRequestRepositoryImpl) UpdatePR(
	ctx context.Context,
	params db.UpdatePRParams,
) (db.PurchaseRequest, error) {
	return r.queries.UpdatePR(ctx, params)
}

// delete
func (r *purchaseRequestRepositoryImpl) DeletePR(
	ctx context.Context,
	id int64,
) error {
	return r.queries.DeletePR(ctx, id)
}