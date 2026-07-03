package repository

import (
	"context"

	db "money-module/db/postgres/generated"

	"github.com/jackc/pgx/v5/pgtype"
)

type AccountRepository interface {
	CreateAccount(
		ctx context.Context,
		id pgtype.UUID,
		accountName string,
		ownerType db.AccountTypes,
		ownerID pgtype.UUID) error
	CreateBalance(
		ctx context.Context,
		id pgtype.UUID,
		accountName string,
		ownerType db.AccountTypes,
		ownerID pgtype.UUID) error
	ReadAccount(ctx context.Context, id pgtype.UUID) (db.Account, error)
	ArchiveAccount(ctx context.Context, id pgtype.UUID) error
}

type sqlAccountRepository struct {
	queries *db.Queries
}

func NewSQLAccountRepository(q *db.Queries) AccountRepository {
	return &sqlAccountRepository{
		queries: q,
	}
}

func (r *sqlAccountRepository) CreateAccount(
	ctx context.Context,
	id pgtype.UUID,
	accountName string,
	ownerType db.AccountTypes,
	ownerID pgtype.UUID,
) error {
	return r.queries.CreateAccount(ctx, db.CreateAccountParams{
		ID: id,
		AccountType: db.NullAccountTypes{
			AccountTypes: db.AccountTypesREAL, Valid: true},
		AccountName: accountName,
		OwnerType:   ownerType,
		OwnerID:     ownerID,
	})
}

func (r *sqlAccountRepository) CreateBalance(
	ctx context.Context,
	id pgtype.UUID,
	accountName string,
	ownerType db.AccountTypes,
	ownerID pgtype.UUID,
) error {
	return r.queries.CreateAccount(ctx, db.CreateAccountParams{
		ID: id,
		AccountType: db.NullAccountTypes{
			AccountTypes: db.AccountTypesVIRTUAL, Valid: true},
		AccountName: accountName,
		OwnerType:   ownerType,
		OwnerID:     ownerID,
	})
}

func (r *sqlAccountRepository) ReadAccount(
	ctx context.Context,
	id pgtype.UUID,
) (db.Account, error) {
	return r.queries.ReadAccount(ctx, id)
}

func (r *sqlAccountRepository) ArchiveAccount(
	ctx context.Context,
	id pgtype.UUID,
) error {
	return r.queries.ArchiveAccount(ctx, id)
}
