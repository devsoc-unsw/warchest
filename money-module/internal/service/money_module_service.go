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

// Bring service layer expectations of repo so we can override for tests
type AccountStore interface {
	CreateAccount(ctx context.Context, arg db.CreateAccountParams) error
	ReadAccount(ctx context.Context, id pgtype.UUID) (db.Account, error)
	ArchiveAccount(ctx context.Context, id pgtype.UUID) error
}

type MoneyModuleService struct {
	Tb      tb.Client
	Pg      *pgxpool.Pool
	Queries AccountStore
}

func (MMS MoneyModuleService) queries() AccountStore {
	if MMS.Queries != nil {
		return MMS.Queries
	}
	return db.New(MMS.Pg)
}

// Public methods for CRUA on specific accounts
type AccountService interface {
	CreateAccount(ctx context.Context, uuid [16]byte) error
	ReadAccount(
		ctx context.Context,
		uuid [16]byte,
		accountType constants.AccountType,
	) (pb.OwnerType, *pb.UUID, pb.WalletType, string, bool, error)
	UpdateAccount(
		ctx context.Context,
		uuid [16]byte,
		ownerType string,
		owerId [16]byte,
		walletName string,
		accountType constants.AccountType,
	) error
	ArchiveAccount(ctx context.Context, uuid [16]byte) error
}

// Public methods for various types of transactions.
type TransactionService interface {
	AccountToBudget(
		ctx context.Context,
		accountUuid [16]byte,
		budgetUuid [16]byte,
		amount uint64,
	) ([16]byte, constants.TransactionStatus, error)
	BudgetToAccount(
		ctx context.Context,
		accountUuid [16]byte,
		budgetUuid [16]byte,
		amount uint64,
	) ([16]byte, constants.TransactionStatus, error)
	AccountToAccount(
		ctx context.Context,
		payer [16]byte,
		payee [16]byte,
		amount uint64,
	) ([16]byte, constants.TransactionStatus, error)
	BudgetToBudget(
		ctx context.Context,
		payer [16]byte,
		payee [16]byte,
		amount uint64,
	) ([16]byte, constants.TransactionStatus, error)
	RootAccountToAccount(
		ctx context.Context,
		accountUuid [16]byte,
		amount uint64,
	) ([16]byte, constants.TransactionStatus, error)
	AccountToRootAccount(
		ctx context.Context,
		accountUuid [16]byte,
		amount uint64,
	) ([16]byte, constants.TransactionStatus, error)
	GetAllAccountTransfers(
		ctx context.Context,
		accountUuid [16]byte,
	) ([]tb.Transfer, constants.TransactionStatus, error)
	GetAllBudgetTransfers(
		ctx context.Context,
		budgetUuid [16]byte,
	) ([]tb.Transfer, constants.TransactionStatus, error)
}

func (MMS MoneyModuleService) CreateAccount(
	ctx context.Context,
	uuid [16]byte,
	name string,
	ownerType string,
	accountType constants.AccountType,
) error {
	var pgOwnerType db.OwnerType
	if err := pgOwnerType.Scan(ownerType); err != nil {
		return err
	}

	// create in metadata db
	queries := MMS.queries()
	queryInfo := db.CreateAccountParams{
		ID:          pgtype.UUID{Bytes: uuid, Valid: true},
		AccountType: db.NullAccountType{AccountType: db.AccountTypeREAL, Valid: true},
		AccountName: name,
		OwnerType:   pgOwnerType,
	}
	if err := queries.CreateAccount(ctx, queryInfo); err != nil {
		return err
	}

	// User accounts and budgets are credit-normal (balance = credits - debits).
	// Setting DebitsMustNotExceedCredits guarantees TigerBeetle rejects any transfer
	// that would push the balance below 0 (world/liquidity pool accounts use Flags: 0).
	newAccount := tb.Account{
		ID:     tb.BytesToUint128(uuid),
		Ledger: constants.RealAccountLedger,
		Code:   constants.RealAccountCode,
		Flags:  tb.AccountFlags{DebitsMustNotExceedCredits: true}.ToUint16(),
	}
	_, err := MMS.Tb.CreateAccounts([]tb.Account{newAccount})
	return err
}

// helper to convert from repo type -> handler type
// probably shouldn't exist but we ball
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

func (MMS MoneyModuleService) ReadAccount(
	ctx context.Context,
	uuid [16]byte,
	accountType constants.AccountType,
) (pb.OwnerType, *pb.UUID, pb.WalletType, string, bool, error) {
	//TODO: accountType not used
	id := pgtype.UUID{Bytes: uuid, Valid: true}
	account, err := MMS.queries().ReadAccount(ctx, id)
	if err != nil {
		return pb.OwnerType_USER, nil, pb.WalletType_BUDGET, "", false, err
	}

	ownerID := &pb.UUID{Value: account.OwnerID.Bytes[:]}
	return toProtoOwnerType(account.OwnerType),
		ownerID,
		pb.WalletType_BUDGET,
		account.AccountName,
		account.ArchivedAt.Valid,
		nil
}

func (MMS MoneyModuleService) UpdateAccount(
	ctx context.Context,
	uuid [16]byte,
	ownerType string,
	owerId [16]byte,
	walletName string,
	accountType constants.AccountType,
) error {
	return errors.New("Unimplemented")
}

func (MMS MoneyModuleService) ArchiveAccount(ctx context.Context, uuid [16]byte) error {
	return MMS.queries().ArchiveAccount(ctx, pgtype.UUID{Bytes: uuid, Valid: true})
}

// --- Public Transaction Methods ---

// Transfer from two accounts (i.e two accounts within the real ledger)
func (MMS MoneyModuleService) AccountToAccount(
	ctx context.Context,
	payer [16]byte,
	payee [16]byte,
	amount uint64,
) ([16]byte, constants.TransactionStatus, error) {
	return MMS.sameLedgerTransfer(
		ctx,
		payer,
		payee,
		amount,
		db.AccountTypeREAL,
		constants.RealAccountLedger,
		constants.RealAccountCode,
	)
}

// Transfer from two budgets (i.e two accounts within the virtual ledger)
func (MMS MoneyModuleService) BudgetToBudget(
	ctx context.Context,
	payer [16]byte,
	payee [16]byte,
	amount uint64,
) ([16]byte, constants.TransactionStatus, error) {
	return MMS.sameLedgerTransfer(
		ctx,
		payer,
		payee,
		amount,
		db.AccountTypeVIRTUAL,
		constants.VirtualAccountLedger,
		constants.VirtualAccountCode,
	)
}

// Transfer from account to budget, note we go over a ledger for this
func (MMS MoneyModuleService) AccountToBudget(
	ctx context.Context,
	accountUuid [16]byte,
	budgetUuid [16]byte,
	amount uint64,
) ([16]byte, constants.TransactionStatus, error) {
	return MMS.crossLedgerTransfer(
		ctx,
		ledgerEndpoint{
			id:     accountUuid,
			typ:    db.AccountTypeREAL,
			ledger: constants.RealAccountLedger,
			code:   constants.RealAccountCode,
			poolID: constants.GetRealLedgerLiquidityBookId(),
		},
		ledgerEndpoint{
			id:     budgetUuid,
			typ:    db.AccountTypeVIRTUAL,
			ledger: constants.VirtualAccountLedger,
			code:   constants.VirtualAccountCode,
			poolID: constants.GetVirtualLedgerLiquidityBookId(),
		},
		amount,
	)
}

// Transfer from budget to account, note we go over a ledger for this
func (MMS MoneyModuleService) BudgetToAccount(
	ctx context.Context,
	accountUuid [16]byte,
	budgetUuid [16]byte,
	amount uint64,
) ([16]byte, constants.TransactionStatus, error) {
	return MMS.crossLedgerTransfer(
		ctx,
		ledgerEndpoint{
			id:     budgetUuid,
			typ:    db.AccountTypeVIRTUAL,
			ledger: constants.VirtualAccountLedger,
			code:   constants.VirtualAccountCode,
			poolID: constants.GetVirtualLedgerLiquidityBookId(),
		},
		ledgerEndpoint{
			id:     accountUuid,
			typ:    db.AccountTypeREAL,
			ledger: constants.RealAccountLedger,
			code:   constants.RealAccountCode,
			poolID: constants.GetRealLedgerLiquidityBookId(),
		},
		amount,
	)
}

func (MMS MoneyModuleService) RootAccountToAccount(
	ctx context.Context,
	accountUuid [16]byte,
	amount uint64,
) ([16]byte, constants.TransactionStatus, error) {
	return MMS.sameLedgerTransfer(
		ctx,
		constants.GetRealLedgerWorldBookId().Bytes(),
		accountUuid,
		amount,
		db.AccountTypeREAL,
		constants.RealAccountLedger,
		constants.RealAccountCode,
	)
}

func (MMS MoneyModuleService) AccountToRootAccount(
	ctx context.Context,
	accountUuid [16]byte,
	amount uint64,
) ([16]byte, constants.TransactionStatus, error) {
	return MMS.sameLedgerTransfer(
		ctx,
		accountUuid,
		constants.GetRealLedgerWorldBookId().Bytes(),
		amount,
		db.AccountTypeREAL,
		constants.RealAccountLedger,
		constants.RealAccountCode,
	)
}

func (MMS MoneyModuleService) GetAllAccountTransfers(
	ctx context.Context,
	accountUuid [16]byte,
) ([]tb.Transfer, constants.TransactionStatus, error) {
	return MMS.getAccountTransfers(ctx, accountUuid, db.AccountTypeREAL, "account")
}

func (MMS MoneyModuleService) GetAllBudgetTransfers(
	ctx context.Context,
	budgetUuid [16]byte,
) ([]tb.Transfer, constants.TransactionStatus, error) {
	return MMS.getAccountTransfers(ctx, budgetUuid, db.AccountTypeVIRTUAL, "budget")
}

// --- Shared Transaction Helpers ---

type ledgerEndpoint struct {
	id     [16]byte
	typ    db.AccountType
	ledger uint32
	code   uint16
	poolID tb.Uint128
}

func (MMS MoneyModuleService) getAccountTransfers(
	ctx context.Context,
	id [16]byte,
	expectedType db.AccountType,
	role string,
) ([]tb.Transfer, constants.TransactionStatus, error) {
	status, err := MMS.validateAccount(
		ctx,
		id,
		expectedType,
		role,
		constants.TransactionAccountNotFound,
		constants.TransactionAccountArchived,
	)
	if err != nil {
		return nil, status, err
	}

	transfers, err := MMS.Tb.GetAccountTransfers(tb.AccountFilter{
		AccountID: tb.BytesToUint128(id),
		Limit:     8189,
		Flags: tb.AccountFilterFlags{
			Debits:  true,
			Credits: true,
		}.ToUint32(),
	})
	if err != nil {
		return nil,
			constants.TransactionInternalError,
			fmt.Errorf("failed to fetch %s transfers: %w", role, err)
	}
	return transfers, constants.TransactionSuccess, nil
}

// helper to validate that a transfer has valid arguments
func validateTransferArgs(
	payer, payee [16]byte,
	amount uint64,
) (constants.TransactionStatus, error) {
	if amount == 0 {
		return constants.TransactionInvalidAmount,
			errors.New("transfer amount must be greater than zero")
	}
	if payer == payee {
		return constants.TransactionSameAccount,
			errors.New("payer and payee accounts must be different")
	}
	return constants.TransactionSuccess, nil
}

// helper to check that an account/budget exists, is not archived,
// and matches expectedType
func (MMS MoneyModuleService) validateAccount(
	ctx context.Context,
	id [16]byte,
	expectedType db.AccountType,
	role string,
	notFoundStatus constants.TransactionStatus,
	archivedStatus constants.TransactionStatus,
) (constants.TransactionStatus, error) {
	pgID := pgtype.UUID{Bytes: id, Valid: true}
	details, err := MMS.queries().ReadAccount(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return notFoundStatus, fmt.Errorf("%s account not found: %w", role, err)
		}
		return constants.TransactionInternalError,
			fmt.Errorf("failed to read %s account: %w", role, err)
	}
	if details.ArchivedAt.Valid {
		return archivedStatus, fmt.Errorf("%s account is archived", role)
	}
	if !details.AccountType.Valid {
		panic(fmt.Sprintf(
			"invariant violation: %s account_type is NULL in postgres",
			role,
		))
	}
	if details.AccountType.AccountType != expectedType {
		return constants.TransactionInvalidAccountType, fmt.Errorf(
			"%s account has invalid type %s (expected %s)",
			role,
			details.AccountType.AccountType,
			expectedType,
		)
	}
	return constants.TransactionSuccess, nil
}

// helper to check that a liquidity pool is in a valid state
func (MMS MoneyModuleService) mustReadLiquidityPool(
	ctx context.Context,
	id tb.Uint128,
	label string,
) db.Account {
	pgID := pgtype.UUID{Bytes: id.Bytes(), Valid: true}
	pool, err := MMS.queries().ReadAccount(ctx, pgID)
	if err != nil {
		panic(fmt.Sprintf(
			"invariant violation: %s liquidity pool missing in postgres: %v",
			label,
			err,
		))
	}
	return pool
}

func (MMS MoneyModuleService) sameLedgerTransfer(
	ctx context.Context,
	payer, payee [16]byte,
	amount uint64,
	expectedType db.AccountType,
	ledger uint32,
	code uint16,
) ([16]byte, constants.TransactionStatus, error) {
	if status, err := validateTransferArgs(payer, payee, amount); err != nil {
		return [16]byte{}, status, err
	}
	status, err := MMS.validateAccount(
		ctx,
		payer,
		expectedType,
		"payer",
		constants.TransactionPayerNotFound,
		constants.TransactionPayerArchived,
	)
	if err != nil {
		return [16]byte{}, status, err
	}
	status, err = MMS.validateAccount(
		ctx,
		payee,
		expectedType,
		"payee",
		constants.TransactionPayeeNotFound,
		constants.TransactionPayeeArchived,
	)
	if err != nil {
		return [16]byte{}, status, err
	}

	txID := tb.ID()
	status, err = MMS.executeTransfers([]tb.Transfer{
		{
			ID:              txID,
			DebitAccountID:  tb.BytesToUint128(payer),
			CreditAccountID: tb.BytesToUint128(payee),
			Amount:          tb.ToUint128(amount),
			UserData128:     txID,
			Ledger:          ledger,
			Code:            code,
		},
	})
	if err != nil {
		return [16]byte{}, status, err
	}
	return txID.Bytes(), status, nil
}

// Helper to transfer between two ledgers
func (MMS MoneyModuleService) crossLedgerTransfer(
	ctx context.Context,
	from ledgerEndpoint,
	to ledgerEndpoint,
	amount uint64,
) ([16]byte, constants.TransactionStatus, error) {
	// In TigerBeetle, accounts cannot transfer to accounts on other ledgers.
	// Each ledger has a dedicated "liquidity pool" account acting as a bridge.
	// Say we are transferring between accounts A and B in ledgers 1 and 2:
	// we begin by transferring A into ledger 1's liquidity pool; if valid,
	// we transfer from ledger 2's liquidity pool into B.
	// If either transfer is invalid, TigerBeetle rolls back both atomically.
	if status, err := validateTransferArgs(from.id, to.id, amount); err != nil {
		return [16]byte{}, status, err
	}
	status, err := MMS.validateAccount(
		ctx,
		from.id,
		from.typ,
		"payer",
		constants.TransactionPayerNotFound,
		constants.TransactionPayerArchived,
	)
	if err != nil {
		return [16]byte{}, status, err
	}
	fromPool := MMS.mustReadLiquidityPool(ctx, from.poolID, string(from.typ))
	toPool := MMS.mustReadLiquidityPool(ctx, to.poolID, string(to.typ))
	status, err = MMS.validateAccount(
		ctx,
		to.id,
		to.typ,
		"payee",
		constants.TransactionPayeeNotFound,
		constants.TransactionPayeeArchived,
	)
	if err != nil {
		return [16]byte{}, status, err
	}

	// Shared logical transaction ID stored in UserData128 on both linked legs
	// so both TigerBeetle transfers correlate to a single transaction_id.
	txID := tb.ID()
	status, err = MMS.executeTransfers([]tb.Transfer{
		{
			ID:              tb.ID(),
			DebitAccountID:  tb.BytesToUint128(from.id),
			CreditAccountID: tb.BytesToUint128(fromPool.ID.Bytes),
			Amount:          tb.ToUint128(amount),
			UserData128:     txID,
			Ledger:          from.ledger,
			Code:            from.code,
			// Linked=true links this transfer atomically with the next one.
			Flags: tb.TransferFlags{Linked: true}.ToUint16(),
		},
		{
			ID:              tb.ID(),
			DebitAccountID:  tb.BytesToUint128(toPool.ID.Bytes),
			CreditAccountID: tb.BytesToUint128(to.id),
			Amount:          tb.ToUint128(amount),
			UserData128:     txID,
			Ledger:          to.ledger,
			Code:            to.code,
		},
	})
	if err != nil {
		return [16]byte{}, status, err
	}
	return txID.Bytes(), status, nil
}

func (MMS MoneyModuleService) executeTransfers(
	transfers []tb.Transfer,
) (constants.TransactionStatus, error) {
	results, err := MMS.Tb.CreateTransfers(transfers)
	if err != nil {
		return constants.TransactionInternalError,
			fmt.Errorf("failed to execute tigerbeetle transfer: %w", err)
	}
	if len(results) != len(transfers) {
		panic(fmt.Sprintf(
			"invariant violation: expected %d tigerbeetle transfer results, got %d",
			len(transfers),
			len(results),
		))
	}

	var rootCause tb.CreateTransferStatus
	allCreated := true
	foundCause := false
	for _, res := range results {
		if res.Status != tb.TransferCreated {
			allCreated = false
			if res.Status != tb.TransferLinkedEventFailed && !foundCause {
				rootCause = res.Status
				foundCause = true
			}
		}
	}
	if allCreated {
		return constants.TransactionSuccess, nil
	}
	if !foundCause {
		panic("invariant violation: linked transfer failed without root cause")
	}

	switch rootCause {
	case tb.TransferExceedsCredits, tb.TransferExceedsDebits:
		return constants.TransactionInsufficientFunds,
			fmt.Errorf("insufficient funds: %s", rootCause.String())
	case tb.TransferDebitAccountNotFound,
		tb.TransferCreditAccountNotFound,
		tb.TransferDebitAccountAlreadyClosed,
		tb.TransferCreditAccountAlreadyClosed,
		tb.TransferAccountsMustHaveTheSameLedger,
		tb.TransferTransferMustHaveTheSameLedgerAsAccounts:
		panic(fmt.Sprintf("postgres/tigerbeetle desync: %s", rootCause.String()))
	case tb.TransferAccountsMustBeDifferent,
		tb.TransferExists,
		tb.TransferLinkedEventChainOpen:
		panic(fmt.Sprintf("invariant violation: %s", rootCause.String()))
	default:
		return constants.TransactionTransferFailed,
			fmt.Errorf("tigerbeetle transfer failed: %s", rootCause.String())
	}
}
