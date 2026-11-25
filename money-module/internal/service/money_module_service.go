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

func (MMS MoneyModuleService) CreateAccount(ctx context.Context, uuid [16]byte, name string, ownerType string) error {
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

func toProtoOwnerType(t db.OwnerType) pb.OwnerType {
	switch t {
	case db.OwnerTypeUSER:
		return pb.OwnerType_USER 
	case db.OwnerTypeORGANISATION:
		return pb.OwnerType_ORG
	default:
		panic("unsupported owner type: " + t)
	}
}

// TODO: Assert that we actually get a budget/account back
func (MMS MoneyModuleService) GetDetails(ctx context.Context, uuid [16]byte, accountType constants.AccountType) (pb.OwnerType, pb.UUID, pb.WalletType, string, bool, error) {
	queries := db.New(MMS.Pg)

	account, err := queries.ReadAccount(ctx, pgtype.UUID{Bytes: uuid, Valid: true})
	if err != nil {
		return pb.OwnerType_USER, pb.UUID{}, pb.WalletType_BUDGET, "", false, err
	}

	//ownerType, ownerId, walletType, walletName, isArchived, err

	return toProtoOwnerType(account.OwnerType), pb.UUID{Value: account.OwnerID.Bytes[:]}, pb.WalletType_BUDGET, account.AccountName, account.ArchivedAt.Valid, nil
}

func (MMS MoneyModuleService) UpdateDetails(ctx context.Context, uuid [16]byte, ownerType string, onwerId [16]byte, walletName string, accountType constants.AccountType) error {
	//TODO: Do not know what to do here
	
}

//transaction_id, status, err := h.Svc.Transaction(req.PayeeId, req.PayerId, req.Amount, false, true)

// TODO: Check return values for this function
func (MMS MoneyModuleService) Transaction(payer_id [16]byte, payee_id [16]byte, amount uint64, PayerType constants.AccountType, PayeeType constants.AccountType) ([16]byte, status, error) {

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
	NewTransfer := tb.Transfer{}

	var transfers = []tb.Account{NewTransfer}
	MMS.Tb.CreateTransfers(transfers)

}

func (MMS MoneyModuleService) GetTransfers(ctx context.Context, uuid [16]byte, accountType constants.AccountType) ([]tb.Transfer, error) {
	//TODO: Finish

}
