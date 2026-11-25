package service

import (
	"context"
	"errors"
	"fmt"
	db "money-module/db/postgres/generated"

	pb "warchest/protos/mm_pb"

	constants "money-module/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	tb "github.com/tigerbeetle/tigerbeetle-go"
)

type MoneyModuleService struct {
	Tb tb.Client
	Pg *pgxpool.Pool
}

// Public methods for modifying specific accounts
type AccountService interface {
	CreateAccount(ctx context.Context, uuid [16]byte) error
	ReadAccount(ctx context.Context, uuid [16]byte, accountType constants.AccountType) (pb.OwnerType, pb.UUID, pb.WalletType, string, bool, error)
	UpdateAccount(ctx context.Context, uuid [16]byte, ownerType string, owerId [16]byte, walletName string, accountType constants.AccountType) error
	ArchiveAccount(ctx context.Context, uuid [16]byte) error
}

// Public methods for making a transaction.
type TransactionService interface {
	AccountToBudget(ctx context.Context, accountUuid [16]byte, budgetUuid [16]byte, amount uint64) error
	BudgetToAccount(ctx context.Context, accountUuid [16]byte, budgetUuid [16]byte, amount uint64) error
	AccountToAccount(ctx context.Context, payer [16]byte, payee [16]byte, amount uint64) (constants.TransactionStatus, error)
	BudgetToBudget(ctx context.Context, payer [16]byte, payee [16]byte, amount uint64) error
}

func (MMS MoneyModuleService) CreateAccount(ctx context.Context, uuid [16]byte, name string, ownerType string, accountType constants.AccountType) error {
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
func (MMS MoneyModuleService) ReadAccount(ctx context.Context, uuid [16]byte, accountType constants.AccountType) (pb.OwnerType, pb.UUID, pb.WalletType, string, bool, error) {
	//TODO: accountType not used
	queries := db.New(MMS.Pg)

	account, err := queries.ReadAccount(ctx, pgtype.UUID{Bytes: uuid, Valid: true})
	if err != nil {
		return pb.OwnerType_USER, pb.UUID{}, pb.WalletType_BUDGET, "", false, err
	}

	//ownerType, ownerId, walletType, walletName, isArchived, err

	return toProtoOwnerType(account.OwnerType), pb.UUID{Value: account.OwnerID.Bytes[:]}, pb.WalletType_BUDGET, account.AccountName, account.ArchivedAt.Valid, nil
}

func (MMS MoneyModuleService) UpdateAccount(ctx context.Context, uuid [16]byte, ownerType string, owerId [16]byte, walletName string, accountType constants.AccountType) error {
	// Update the details of an account
	return errors.New("Unimplemented")
}

func (MMS MoneyModuleService) ArchiveAccount(ctx context.Context, uuid [16]byte) error {
	//TODO: accountType not used
	queries := db.New(MMS.Pg)

	err := queries.ArchiveAccount(ctx, pgtype.UUID{Bytes: uuid, Valid: true})

	return err
}

func (MMS MoneyModuleService) AccountToAccount(ctx context.Context, payer [16]byte, payee [16]byte, amount uint64) (constants.TransactionStatus, error) {
	if amount == 0 {
		return constants.TransactionInvalidAmount, errors.New("transfer amount must be greater than zero")
	}
	if payer == payee {
		return constants.TransactionSameAccount, errors.New("payer and payee accounts must be different")
	}

	// query postgres for the tigerbeetle details to use
	queries := db.New(MMS.Pg)

	payerDetails, err := queries.ReadAccount(ctx, pgtype.UUID{Bytes: payer, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return constants.TransactionPayerNotFound, fmt.Errorf("payer account not found: %w", err)
		}
		return constants.TransactionInternalError, fmt.Errorf("failed to read payer account: %w", err)
	}
	if payerDetails.ArchivedAt.Valid {
		return constants.TransactionPayerArchived, errors.New("payer account is archived")
	}
	if !payerDetails.AccountType.Valid {
		panic("invariant violation: payer account_type is NULL in postgres")
	}
	if payerDetails.AccountType.AccountType != db.AccountTypeREAL {
		return constants.TransactionInvalidAccountType, errors.New("payer is not a real account")
	}

	payeeDetails, err := queries.ReadAccount(ctx, pgtype.UUID{Bytes: payee, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return constants.TransactionPayeeNotFound, fmt.Errorf("payee account not found: %w", err)
		}
		return constants.TransactionInternalError, fmt.Errorf("failed to read payee account: %w", err)
	}
	if payeeDetails.ArchivedAt.Valid {
		return constants.TransactionPayeeArchived, errors.New("payee account is archived")
	}
	if !payeeDetails.AccountType.Valid {
		panic("invariant violation: payee account_type is NULL in postgres")
	}
	if payeeDetails.AccountType.AccountType != db.AccountTypeREAL {
		return constants.TransactionInvalidAccountType, errors.New("payee is not a real account")
	}

	transfer := tb.Transfer{
		ID:              tb.ID(),
		DebitAccountID:  tb.BytesToUint128(payer),
		CreditAccountID: tb.BytesToUint128(payee),
		Amount:          tb.ToUint128(amount),
		Ledger:          constants.RealAccountLedger,
		Code:            constants.RealAccountCode,
	}

	transferResults, err := MMS.Tb.CreateTransfers([]tb.Transfer{transfer})
	if err != nil {
		return constants.TransactionInternalError, fmt.Errorf("failed to execute tigerbeetle transfer: %w", err)
	}

	if len(transferResults) != 1 {
		panic(fmt.Sprintf("invariant violation: expected 1 tigerbeetle transfer result, got %d", len(transferResults)))
	}

	switch transferResults[0].Status {
	case tb.TransferCreated:
		return constants.TransactionSuccess, nil
	case tb.TransferExceedsCredits, tb.TransferExceedsDebits:
		return constants.TransactionInsufficientFunds, fmt.Errorf("insufficient funds: %s", transferResults[0].Status.String())
	case tb.TransferDebitAccountNotFound,
		tb.TransferCreditAccountNotFound,
		tb.TransferDebitAccountAlreadyClosed,
		tb.TransferCreditAccountAlreadyClosed,
		tb.TransferAccountsMustHaveTheSameLedger,
		tb.TransferTransferMustHaveTheSameLedgerAsAccounts:
		panic(fmt.Sprintf("postgres/tigerbeetle desync: %s", transferResults[0].Status.String()))
	case tb.TransferAccountsMustBeDifferent, tb.TransferExists:
		panic(fmt.Sprintf("invariant violation: %s", transferResults[0].Status.String()))
	default:
		return constants.TransactionTransferFailed, fmt.Errorf("tigerbeetle transfer failed: %s", transferResults[0].Status.String())
	}
}

func (MMS MoneyModuleService) BudgetToBudget(ctx context.Context, payer [16]byte, payee [16]byte, amount uint64) (constants.TransactionStatus, error) {
	if amount == 0 {
		return constants.TransactionInvalidAmount, errors.New("transfer amount must be greater than zero")
	}
	if payer == payee {
		return constants.TransactionSameAccount, errors.New("payer and payee accounts must be different")
	}

	// query postgres for the tigerbeetle details to use
	queries := db.New(MMS.Pg)

	payerDetails, err := queries.ReadAccount(ctx, pgtype.UUID{Bytes: payer, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return constants.TransactionPayerNotFound, fmt.Errorf("payer account not found: %w", err)
		}
		return constants.TransactionInternalError, fmt.Errorf("failed to read payer account: %w", err)
	}
	if payerDetails.ArchivedAt.Valid {
		return constants.TransactionPayerArchived, errors.New("payer account is archived")
	}
	if !payerDetails.AccountType.Valid {
		panic("invariant violation: payer account_type is NULL in postgres")
	}
	if payerDetails.AccountType.AccountType != db.AccountTypeREAL {
		return constants.TransactionInvalidAccountType, errors.New("payer is not a real account")
	}

	payeeDetails, err := queries.ReadAccount(ctx, pgtype.UUID{Bytes: payee, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return constants.TransactionPayeeNotFound, fmt.Errorf("payee account not found: %w", err)
		}
		return constants.TransactionInternalError, fmt.Errorf("failed to read payee account: %w", err)
	}
	if payeeDetails.ArchivedAt.Valid {
		return constants.TransactionPayeeArchived, errors.New("payee account is archived")
	}
	if !payeeDetails.AccountType.Valid {
		panic("invariant violation: payee account_type is NULL in postgres")
	}
	if payeeDetails.AccountType.AccountType != db.AccountTypeREAL {
		return constants.TransactionInvalidAccountType, errors.New("payee is not a real account")
	}

	transfer := tb.Transfer{
		ID:              tb.ID(),
		DebitAccountID:  tb.BytesToUint128(payer),
		CreditAccountID: tb.BytesToUint128(payee),
		Amount:          tb.ToUint128(amount),
		Ledger:          constants.VirtualAccountLedger,
		Code:            constants.VirtualAccountCode,
	}

	transferResults, err := MMS.Tb.CreateTransfers([]tb.Transfer{transfer})
	if err != nil {
		return constants.TransactionInternalError, fmt.Errorf("failed to execute tigerbeetle transfer: %w", err)
	}

	if len(transferResults) != 1 {
		panic(fmt.Sprintf("invariant violation: expected 1 tigerbeetle transfer result, got %d", len(transferResults)))
	}

	switch transferResults[0].Status {
	case tb.TransferCreated:
		return constants.TransactionSuccess, nil
	case tb.TransferExceedsCredits, tb.TransferExceedsDebits:
		return constants.TransactionInsufficientFunds, fmt.Errorf("insufficient funds: %s", transferResults[0].Status.String())
	case tb.TransferDebitAccountNotFound,
		tb.TransferCreditAccountNotFound,
		tb.TransferDebitAccountAlreadyClosed,
		tb.TransferCreditAccountAlreadyClosed,
		tb.TransferAccountsMustHaveTheSameLedger,
		tb.TransferTransferMustHaveTheSameLedgerAsAccounts:
		panic(fmt.Sprintf("postgres/tigerbeetle desync: %s", transferResults[0].Status.String()))
	case tb.TransferAccountsMustBeDifferent, tb.TransferExists:
		panic(fmt.Sprintf("invariant violation: %s", transferResults[0].Status.String()))
	default:
		return constants.TransactionTransferFailed, fmt.Errorf("tigerbeetle transfer failed: %s", transferResults[0].Status.String())
	}
}
