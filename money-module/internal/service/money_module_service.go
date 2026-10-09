package service

import (
	"context"
	db "money-module/db/postgres/generated"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	tb "github.com/tigerbeetle/tigerbeetle-go"
	"money-module/internal/config"
)

type MoneyModuleService struct {
	Tb tb.Client
	Pg *pgxpool.Pool
}

type AccountService interface {
	CreateAccount(ctx context.Context, uuid tb.Uint128) error
}


func (MMS MoneyModuleService) CreateAccount(ctx context.Context, uuid [16]byte, name string, ownerType string) error {
	var pgOwnerType db.OwnerType;
	err := pgOwnerType.Scan(ownerType);
	if err != nil {
		return err
	}

	// create in metadata db
	queries := db.New(MMS.Pg)
	queryInfo := db.CreateAccountParams{
		ID: pgtype.UUID{Bytes: uuid, Valid: true},
		AccountType: db.NullAccountType{AccountType: db.AccountTypeREAL, Valid: true},
		AccountName: name,
		OwnerType: pgOwnerType,
	}
	err = queries.CreateAccount(ctx, queryInfo)
	if err != nil {
		return err
	}
	
	// create in tigerbeetle
	NewAccount := tb.Account {
		ID: tb.BytesToUint128(uuid),
		Ledger: constants.RealAccountLedger,
		Code: constants.RealAccountCode,
	}
	var accounts = []tb.Account{NewAccount}
	_, err = MMS.Tb.CreateAccounts(accounts)
	if (err != nil) {
		return err
	}

	return nil 
}