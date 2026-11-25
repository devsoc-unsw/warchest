package service

import (
	"context"
	db "money-module/db/postgres/generated"

	pb "warchest/protos/mm_pb"

	constants "money-module/internal/config"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	tb "github.com/tigerbeetle/tigerbeetle-go"
)

type MoneyModuleService struct {
	Tb tb.Client
	Pg *pgxpool.Pool
}

type AccountService interface {
	CreateAccount(ctx context.Context, uuid tb.Uint128) error
}

func (MMS MoneyModuleService) CreateAccount(ctx context.Context, uuid [16]byte, name string, ownerType string, accountType constants.AccountType) error {
	//TODO: accountType not used
	var pgOwnerType db.OwnerType
	err := pgOwnerType.Scan(ownerType)
	if err != nil {
		return err
	}

	// create in metadata db
	queries := db.New(MMS.Pg)
	queryInfo := db.CreateAccountParams{
		ID:          pgtype.UUID{Bytes: uuid, Valid: true},
		AccountType: db.NullAccountType{AccountType: db.AccountTypeREAL, Valid: true},
		AccountName: name,
		OwnerType:   pgOwnerType,
	}
	err = queries.CreateAccount(ctx, queryInfo)
	if err != nil {
		return err
	}

	// create in tigerbeetle
	NewAccount := tb.Account{
		ID:     tb.BytesToUint128(uuid),
		Ledger: constants.RealAccountLedger,
		Code:   constants.RealAccountCode,
	}
	var accounts = []tb.Account{NewAccount}
	_, err = MMS.Tb.CreateAccounts(accounts)
	if err != nil {
		return err
	}

	return nil
}

// TODO: Assert that we actually get a budget/account back
func (MMS MoneyModuleService) GetDetails(ctx context.Context, uuid [16]byte, accountType constants.AccountType) (pb.OwnerType, pb.UUID, pb.WalletType, string, bool, error) {
	//TODO: accountType not used
	queries := db.New(MMS.Pg)

	account, err := queries.ReadAccount(ctx, pgtype.UUID{Bytes: uuid, Valid: true})
	if err != nil {
		return pb.OwnerType_USER, pb.UUID{}, pb.WalletType_BUDGET, "", false, err
	}

	//ownerType, ownerId, walletType, walletName, isArchived, err
	//TODO: Determine how to check if archived
	return account.OwnerType, account.OwnerID.Bytes, pb.WalletType_BUDGET, account.AccountName, false, nil

}

func (MMS MoneyModuleService) UpdateDetails(ctx context.Context, uuid [16]byte, ownerType string, onwerId [16]byte, walletName string, accountType constants.AccountType) error {
	//TODO: Do not know what to do here
	//TODO: accountType not used
	queries := db.New(MMS.Pg)
}

//transaction_id, status, err := h.Svc.Transaction(req.PayeeId, req.PayerId, req.Amount, false, true)

// TODO: Check return values for this function
func (MMS MoneyModuleService) Transaction(payer_id *[16]byte, payee_id *[16]byte, amount uint64, PayerType constants.AccountType, PayeeType constants.AccountType) ([16]byte, pb.Status, error) {
	//TODO: PayerType, PayeeType not used
	/*type Transfer struct {
		ID              Uint128
		DebitAccountID  Uint128
		CreditAccountID Uint128
		Amount          Uint128
		PendingID       Uint128
		UserData128     Uint128
		UserData64      uint64
		UserData32      uint32
		Timeout         uint32
		Ledger          uint32
		Code            uint16
		Flags           uint16
		Timestamp       uint64
	}*/
	if payer_id == nil {

	} else {

	}
	if payee_id == nil {

	} else {

	}
	NewTransfer := tb.Transfer{ID: uuid, DebitAccountID: debit_uuid, CreditAccountID: credit_uuid, Amount: amount}

	var transfers = []tb.Account{NewTransfer}
	MMS.Tb.CreateTransfers(transfers)

}

func (MMS MoneyModuleService) GetTransfers(ctx context.Context, uuid [16]byte, accountType constants.AccountType) ([]tb.Transfer, error) {
	//TODO: accountType not used
	//TODO: Finish

}

func (MMS MoneyModuleService) ArchiveAccount(ctx context.Context, uuid [16]byte, accountType constants.AccountType) error {
	//TODO: accountType not used
	//TODO: Finish

}
