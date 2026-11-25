package service

import (
	"context"
	"errors"
	"testing"
	"time"

	db "money-module/db/postgres/generated"
	constants "money-module/internal/config"
	pb "warchest/protos/mm_pb"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	tb "github.com/tigerbeetle/tigerbeetle-go"
)

// to test our service layers, we provide mocks for our repo layers, based on their interfaces
// We then used the mocked repo + our actual service logic

type fakeAccountStore struct {
	accounts   map[[16]byte]db.Account
	readErrors map[[16]byte]error
	createErr  error
	archiveErr error

	createdParams []db.CreateAccountParams
	archivedIDs   [][16]byte
}

func newFakeAccountStore() *fakeAccountStore {
	return &fakeAccountStore{
		accounts:   make(map[[16]byte]db.Account),
		readErrors: make(map[[16]byte]error),
	}
}

func (f *fakeAccountStore) CreateAccount(_ context.Context, arg db.CreateAccountParams) error {
	f.createdParams = append(f.createdParams, arg)
	if f.createErr != nil {
		return f.createErr
	}
	f.accounts[arg.ID.Bytes] = db.Account{
		ID:          arg.ID,
		AccountType: arg.AccountType,
		AccountName: arg.AccountName,
		OwnerType:   arg.OwnerType,
		OwnerID:     arg.OwnerID,
	}
	return nil
}

func (f *fakeAccountStore) ReadAccount(_ context.Context, id pgtype.UUID) (db.Account, error) {
	if err, ok := f.readErrors[id.Bytes]; ok {
		return db.Account{}, err
	}
	acc, ok := f.accounts[id.Bytes]
	if !ok {
		return db.Account{}, pgx.ErrNoRows
	}
	return acc, nil
}

func (f *fakeAccountStore) ArchiveAccount(_ context.Context, id pgtype.UUID) error {
	f.archivedIDs = append(f.archivedIDs, id.Bytes)
	if f.archiveErr != nil {
		return f.archiveErr
	}
	if acc, ok := f.accounts[id.Bytes]; ok {
		acc.ArchivedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		f.accounts[id.Bytes] = acc
	}
	return nil
}

type fakeTBClient struct {
	tb.Client // embed interface so only used methods need to be implemented

	createAccountsRes []tb.CreateAccountResult
	createAccountsErr error
	createdAccounts   []tb.Account

	createTransfersRes []tb.CreateTransferResult
	createTransfersErr error
	createdTransfers   []tb.Transfer

	accountTransfersRes []tb.Transfer
	accountTransfersErr error
	lastAccountFilter   tb.AccountFilter
}

func (f *fakeTBClient) CreateAccounts(accounts []tb.Account) ([]tb.CreateAccountResult, error) {
	f.createdAccounts = append(f.createdAccounts, accounts...)
	if f.createAccountsErr != nil {
		return nil, f.createAccountsErr
	}
	return f.createAccountsRes, nil
}

func (f *fakeTBClient) CreateTransfers(transfers []tb.Transfer) ([]tb.CreateTransferResult, error) {
	f.createdTransfers = append(f.createdTransfers, transfers...)
	if f.createTransfersErr != nil {
		return nil, f.createTransfersErr
	}
	return f.createTransfersRes, nil
}

func (f *fakeTBClient) GetAccountTransfers(filter tb.AccountFilter) ([]tb.Transfer, error) {
	f.lastAccountFilter = filter
	if f.accountTransfersErr != nil {
		return nil, f.accountTransfersErr
	}
	return f.accountTransfersRes, nil
}

// --- Helpers ---

func testUUID(lastByte byte) [16]byte {
	return [16]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x47, 0x88, 0x89, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, lastByte}
}

func makeAccount(id [16]byte, accType db.AccountType, archived bool) db.Account {
	acc := db.Account{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		AccountType: db.NullAccountType{AccountType: accType, Valid: true},
		AccountName: "test-account",
		OwnerType:   db.OwnerTypeUSER,
		OwnerID:     pgtype.UUID{Bytes: testUUID(0x99), Valid: true},
	}
	if archived {
		acc.ArchivedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	}
	return acc
}

func seedLiquidityPools(store *fakeAccountStore) {
	realLP := constants.GetRealLedgerLiquidityBookId().Bytes()
	virtualLP := constants.GetVirtualLedgerLiquidityBookId().Bytes()
	store.accounts[realLP] = makeAccount(realLP, db.AccountTypeREAL, false)
	store.accounts[virtualLP] = makeAccount(virtualLP, db.AccountTypeVIRTUAL, false)
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic, but did not panic")
		}
	}()
	fn()
}

// --- Account CRUD Tests ---

func TestCreateAccount(t *testing.T) {
	ctx := context.Background()
	id := testUUID(1)

	t.Run("success", func(t *testing.T) {
		store := newFakeAccountStore()
		tbClient := &fakeTBClient{
			createAccountsRes: []tb.CreateAccountResult{{Status: tb.AccountCreated}},
		}
		svc := MoneyModuleService{Tb: tbClient, Queries: store}

		err := svc.CreateAccount(ctx, id, "Main Wallet", "USER", constants.ACCOUNT)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(store.createdParams) != 1 {
			t.Fatalf("expected 1 postgres insert, got %d", len(store.createdParams))
		}
		if store.createdParams[0].AccountName != "Main Wallet" {
			t.Errorf("expected AccountName 'Main Wallet', got %q", store.createdParams[0].AccountName)
		}
		if len(tbClient.createdAccounts) != 1 {
			t.Fatalf("expected 1 tigerbeetle account created, got %d", len(tbClient.createdAccounts))
		}
		if tbClient.createdAccounts[0].ID != tb.BytesToUint128(id) {
			t.Errorf("unexpected tigerbeetle account ID")
		}
		if tbClient.createdAccounts[0].Ledger != constants.RealAccountLedger {
			t.Errorf("expected ledger %d, got %d", constants.RealAccountLedger, tbClient.createdAccounts[0].Ledger)
		}
		wantFlags := tb.AccountFlags{DebitsMustNotExceedCredits: true}.ToUint16()
		if tbClient.createdAccounts[0].Flags != wantFlags {
			t.Errorf("expected DebitsMustNotExceedCredits flag (%d), got %d", wantFlags, tbClient.createdAccounts[0].Flags)
		}
	})

	t.Run("postgres error skips tigerbeetle", func(t *testing.T) {
		store := newFakeAccountStore()
		store.createErr = errors.New("db down")
		tbClient := &fakeTBClient{}
		svc := MoneyModuleService{Tb: tbClient, Queries: store}

		if err := svc.CreateAccount(ctx, id, "Main", "USER", constants.ACCOUNT); err == nil {
			t.Fatalf("expected error, got nil")
		}
		if len(tbClient.createdAccounts) != 0 {
			t.Fatalf("expected tigerbeetle not to be called on postgres error")
		}
	})

	t.Run("tigerbeetle error propagated", func(t *testing.T) {
		store := newFakeAccountStore()
		tbClient := &fakeTBClient{createAccountsErr: errors.New("tb down")}
		svc := MoneyModuleService{Tb: tbClient, Queries: store}

		if err := svc.CreateAccount(ctx, id, "Main", "USER", constants.ACCOUNT); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestReadAccount(t *testing.T) {
	ctx := context.Background()
	id := testUUID(1)

	t.Run("success active user account", func(t *testing.T) {
		store := newFakeAccountStore()
		store.accounts[id] = makeAccount(id, db.AccountTypeREAL, false)
		svc := MoneyModuleService{Queries: store}

		ownerType, ownerID, _, name, isArchived, err := svc.ReadAccount(ctx, id, constants.ACCOUNT)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ownerType != pb.OwnerType_USER {
			t.Errorf("expected USER, got %v", ownerType)
		}
		if name != "test-account" {
			t.Errorf("expected 'test-account', got %q", name)
		}
		if isArchived {
			t.Errorf("expected isArchived=false")
		}
		if len(ownerID.Value) != 16 {
			t.Errorf("expected 16-byte ownerID, got %d", len(ownerID.Value))
		}
	})

	t.Run("success archived org account", func(t *testing.T) {
		store := newFakeAccountStore()
		acc := makeAccount(id, db.AccountTypeREAL, true)
		acc.OwnerType = db.OwnerTypeORGANISATION
		store.accounts[id] = acc
		svc := MoneyModuleService{Queries: store}

		ownerType, _, _, _, isArchived, err := svc.ReadAccount(ctx, id, constants.ACCOUNT)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ownerType != pb.OwnerType_ORG {
			t.Errorf("expected ORG, got %v", ownerType)
		}
		if !isArchived {
			t.Errorf("expected isArchived=true")
		}
	})

	t.Run("not found returns error", func(t *testing.T) {
		store := newFakeAccountStore()
		svc := MoneyModuleService{Queries: store}

		_, _, _, _, _, err := svc.ReadAccount(ctx, id, constants.ACCOUNT)
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("expected pgx.ErrNoRows, got %v", err)
		}
	})

	t.Run("unsupported owner type panics", func(t *testing.T) {
		store := newFakeAccountStore()
		acc := makeAccount(id, db.AccountTypeREAL, false)
		acc.OwnerType = db.OwnerTypePLATFORM
		store.accounts[id] = acc
		svc := MoneyModuleService{Queries: store}

		assertPanics(t, func() {
			_, _, _, _, _, _ = svc.ReadAccount(ctx, id, constants.ACCOUNT)
		})
	})
}

func TestArchiveAccount(t *testing.T) {
	ctx := context.Background()
	id := testUUID(1)

	t.Run("success", func(t *testing.T) {
		store := newFakeAccountStore()
		store.accounts[id] = makeAccount(id, db.AccountTypeREAL, false)
		svc := MoneyModuleService{Queries: store}

		if err := svc.ArchiveAccount(ctx, id); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !store.accounts[id].ArchivedAt.Valid {
			t.Errorf("expected account to be marked archived")
		}
	})

	t.Run("error propagated", func(t *testing.T) {
		store := newFakeAccountStore()
		store.archiveErr = errors.New("archive failed")
		svc := MoneyModuleService{Queries: store}

		if err := svc.ArchiveAccount(ctx, id); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

// --- AccountToAccount Tests ---

func TestAccountToAccount_Success(t *testing.T) {
	ctx := context.Background()
	payer, payee := testUUID(1), testUUID(2)

	store := newFakeAccountStore()
	store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
	store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, false)

	tbClient := &fakeTBClient{
		createTransfersRes: []tb.CreateTransferResult{{Status: tb.TransferCreated}},
	}
	svc := MoneyModuleService{Tb: tbClient, Queries: store}

	txID, status, err := svc.AccountToAccount(ctx, payer, payee, 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != constants.TransactionSuccess {
		t.Fatalf("expected %s, got %s", constants.TransactionSuccess, status)
	}
	if len(tbClient.createdTransfers) != 1 {
		t.Fatalf("expected 1 transfer, got %d", len(tbClient.createdTransfers))
	}
	tr := tbClient.createdTransfers[0]
	if txID == ([16]byte{}) || txID != tr.ID.Bytes() || txID != tr.UserData128.Bytes() {
		t.Errorf("expected txID to match transfer ID and UserData128")
	}
	if tr.DebitAccountID != tb.BytesToUint128(payer) || tr.CreditAccountID != tb.BytesToUint128(payee) {
		t.Errorf("unexpected debit/credit account IDs")
	}
	if tr.Amount != tb.ToUint128(500) {
		t.Errorf("expected amount 500, got %v", tr.Amount)
	}
	if tr.Ledger != constants.RealAccountLedger || tr.Code != constants.RealAccountCode {
		t.Errorf("expected RealAccountLedger/Code, got %d/%d", tr.Ledger, tr.Code)
	}
	if tr.Flags != 0 {
		t.Errorf("expected Flags=0, got %d", tr.Flags)
	}
}

func TestAccountToAccount_Errors(t *testing.T) {
	ctx := context.Background()
	payer, payee := testUUID(1), testUUID(2)

	tests := []struct {
		name       string
		payer      [16]byte
		payee      [16]byte
		amount     uint64
		setup      func(store *fakeAccountStore, tbClient *fakeTBClient)
		wantStatus constants.TransactionStatus
	}{
		{
			name:       "zero amount",
			payer:      payer,
			payee:      payee,
			amount:     0,
			setup:      func(_ *fakeAccountStore, _ *fakeTBClient) {},
			wantStatus: constants.TransactionInvalidAmount,
		},
		{
			name:       "same account",
			payer:      payer,
			payee:      payer,
			amount:     100,
			setup:      func(_ *fakeAccountStore, _ *fakeTBClient) {},
			wantStatus: constants.TransactionSameAccount,
		},
		{
			name:   "payer not found",
			payer:  payer,
			payee:  payee,
			amount: 100,
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, false)
			},
			wantStatus: constants.TransactionPayerNotFound,
		},
		{
			name:   "payer db error",
			payer:  payer,
			payee:  payee,
			amount: 100,
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				store.readErrors[payer] = errors.New("connection reset")
			},
			wantStatus: constants.TransactionInternalError,
		},
		{
			name:   "payer archived",
			payer:  payer,
			payee:  payee,
			amount: 100,
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, true)
				store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, false)
			},
			wantStatus: constants.TransactionPayerArchived,
		},
		{
			name:   "payer is virtual budget",
			payer:  payer,
			payee:  payee,
			amount: 100,
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeVIRTUAL, false)
				store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, false)
			},
			wantStatus: constants.TransactionInvalidAccountType,
		},
		{
			name:   "payee not found",
			payer:  payer,
			payee:  payee,
			amount: 100,
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
			},
			wantStatus: constants.TransactionPayeeNotFound,
		},
		{
			name:   "payee db error",
			payer:  payer,
			payee:  payee,
			amount: 100,
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
				store.readErrors[payee] = errors.New("connection reset")
			},
			wantStatus: constants.TransactionInternalError,
		},
		{
			name:   "payee archived",
			payer:  payer,
			payee:  payee,
			amount: 100,
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
				store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, true)
			},
			wantStatus: constants.TransactionPayeeArchived,
		},
		{
			name:   "payee is virtual budget",
			payer:  payer,
			payee:  payee,
			amount: 100,
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
				store.accounts[payee] = makeAccount(payee, db.AccountTypeVIRTUAL, false)
			},
			wantStatus: constants.TransactionInvalidAccountType,
		},
		{
			name:   "tigerbeetle transport error",
			payer:  payer,
			payee:  payee,
			amount: 100,
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
				store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, false)
				tbClient.createTransfersErr = errors.New("tb timeout")
			},
			wantStatus: constants.TransactionInternalError,
		},
		{
			name:   "insufficient funds exceeds credits",
			payer:  payer,
			payee:  payee,
			amount: 100,
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
				store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, false)
				tbClient.createTransfersRes = []tb.CreateTransferResult{{Status: tb.TransferExceedsCredits}}
			},
			wantStatus: constants.TransactionInsufficientFunds,
		},
		{
			name:   "insufficient funds exceeds debits",
			payer:  payer,
			payee:  payee,
			amount: 100,
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
				store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, false)
				tbClient.createTransfersRes = []tb.CreateTransferResult{{Status: tb.TransferExceedsDebits}}
			},
			wantStatus: constants.TransactionInsufficientFunds,
		},
		{
			name:   "unmapped tigerbeetle failure",
			payer:  payer,
			payee:  payee,
			amount: 100,
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
				store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, false)
				tbClient.createTransfersRes = []tb.CreateTransferResult{{Status: tb.TransferOverflowsDebits}}
			},
			wantStatus: constants.TransactionTransferFailed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := newFakeAccountStore()
			tbClient := &fakeTBClient{}
			tc.setup(store, tbClient)
			svc := MoneyModuleService{Tb: tbClient, Queries: store}

			txID, status, err := svc.AccountToAccount(ctx, tc.payer, tc.payee, tc.amount)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if txID != ([16]byte{}) {
				t.Fatalf("expected zero txID on error, got %x", txID)
			}
			if status != tc.wantStatus {
				t.Fatalf("expected status %s, got %s", tc.wantStatus, status)
			}
		})
	}
}

func TestAccountToAccount_Panics(t *testing.T) {
	ctx := context.Background()
	payer, payee := testUUID(1), testUUID(2)

	tests := []struct {
		name  string
		setup func(store *fakeAccountStore, tbClient *fakeTBClient)
	}{
		{
			name: "payer NULL account_type",
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				acc := makeAccount(payer, db.AccountTypeREAL, false)
				acc.AccountType.Valid = false
				store.accounts[payer] = acc
				store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, false)
			},
		},
		{
			name: "payee NULL account_type",
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
				acc := makeAccount(payee, db.AccountTypeREAL, false)
				acc.AccountType.Valid = false
				store.accounts[payee] = acc
			},
		},
		{
			name: "unexpected result count",
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
				store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, false)
				tbClient.createTransfersRes = []tb.CreateTransferResult{}
			},
		},
		{
			name: "desync debit account not found",
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
				store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, false)
				tbClient.createTransfersRes = []tb.CreateTransferResult{{Status: tb.TransferDebitAccountNotFound}}
			},
		},
		{
			name: "desync credit account closed",
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
				store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, false)
				tbClient.createTransfersRes = []tb.CreateTransferResult{{Status: tb.TransferCreditAccountAlreadyClosed}}
			},
		},
		{
			name: "desync ledger mismatch",
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
				store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, false)
				tbClient.createTransfersRes = []tb.CreateTransferResult{{Status: tb.TransferAccountsMustHaveTheSameLedger}}
			},
		},
		{
			name: "invariant duplicate transfer ID",
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
				store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, false)
				tbClient.createTransfersRes = []tb.CreateTransferResult{{Status: tb.TransferExists}}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := newFakeAccountStore()
			tbClient := &fakeTBClient{}
			tc.setup(store, tbClient)
			svc := MoneyModuleService{Tb: tbClient, Queries: store}

			assertPanics(t, func() {
				_, _, _ = svc.AccountToAccount(ctx, payer, payee, 100)
			})
		})
	}
}

// --- BudgetToBudget Tests ---

func TestBudgetToBudget(t *testing.T) {
	ctx := context.Background()
	payer, payee := testUUID(10), testUUID(11)

	t.Run("success", func(t *testing.T) {
		store := newFakeAccountStore()
		store.accounts[payer] = makeAccount(payer, db.AccountTypeVIRTUAL, false)
		store.accounts[payee] = makeAccount(payee, db.AccountTypeVIRTUAL, false)

		tbClient := &fakeTBClient{
			createTransfersRes: []tb.CreateTransferResult{{Status: tb.TransferCreated}},
		}
		svc := MoneyModuleService{Tb: tbClient, Queries: store}

		txID, status, err := svc.BudgetToBudget(ctx, payer, payee, 250)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != constants.TransactionSuccess {
			t.Fatalf("expected %s, got %s", constants.TransactionSuccess, status)
		}
		tr := tbClient.createdTransfers[0]
		if txID == ([16]byte{}) || txID != tr.ID.Bytes() {
			t.Errorf("expected txID to match transfer ID")
		}
		if tr.Ledger != constants.VirtualAccountLedger || tr.Code != constants.VirtualAccountCode {
			t.Errorf("expected VirtualAccountLedger/Code, got %d/%d", tr.Ledger, tr.Code)
		}
	})

	t.Run("rejects real account as payer", func(t *testing.T) {
		store := newFakeAccountStore()
		store.accounts[payer] = makeAccount(payer, db.AccountTypeREAL, false)
		store.accounts[payee] = makeAccount(payee, db.AccountTypeVIRTUAL, false)
		svc := MoneyModuleService{Tb: &fakeTBClient{}, Queries: store}

		_, status, err := svc.BudgetToBudget(ctx, payer, payee, 250)
		if err == nil || status != constants.TransactionInvalidAccountType {
			t.Fatalf("expected TransactionInvalidAccountType, got (%s, %v)", status, err)
		}
	})

	t.Run("rejects real account as payee", func(t *testing.T) {
		store := newFakeAccountStore()
		store.accounts[payer] = makeAccount(payer, db.AccountTypeVIRTUAL, false)
		store.accounts[payee] = makeAccount(payee, db.AccountTypeREAL, false)
		svc := MoneyModuleService{Tb: &fakeTBClient{}, Queries: store}

		_, status, err := svc.BudgetToBudget(ctx, payer, payee, 250)
		if err == nil || status != constants.TransactionInvalidAccountType {
			t.Fatalf("expected TransactionInvalidAccountType, got (%s, %v)", status, err)
		}
	})
}

// --- AccountToBudget Tests ---

func TestAccountToBudget_Success(t *testing.T) {
	ctx := context.Background()
	accID, budgetID := testUUID(20), testUUID(21)

	store := newFakeAccountStore()
	seedLiquidityPools(store)
	store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
	store.accounts[budgetID] = makeAccount(budgetID, db.AccountTypeVIRTUAL, false)

	tbClient := &fakeTBClient{
		createTransfersRes: []tb.CreateTransferResult{
			{Status: tb.TransferCreated},
			{Status: tb.TransferCreated},
		},
	}
	svc := MoneyModuleService{Tb: tbClient, Queries: store}

	txID, status, err := svc.AccountToBudget(ctx, accID, budgetID, 1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != constants.TransactionSuccess {
		t.Fatalf("expected %s, got %s", constants.TransactionSuccess, status)
	}
	if len(tbClient.createdTransfers) != 2 {
		t.Fatalf("expected 2 linked transfers, got %d", len(tbClient.createdTransfers))
	}

	leg1, leg2 := tbClient.createdTransfers[0], tbClient.createdTransfers[1]
	if txID == ([16]byte{}) || leg1.UserData128.Bytes() != txID || leg2.UserData128.Bytes() != txID {
		t.Errorf("expected both linked legs to share txID in UserData128")
	}
	if leg1.DebitAccountID != tb.BytesToUint128(accID) || leg1.CreditAccountID != constants.GetRealLedgerLiquidityBookId() {
		t.Errorf("leg 1 debit/credit mismatch")
	}
	expectedLinkedFlag := tb.TransferFlags{Linked: true}.ToUint16()
	if leg1.Ledger != constants.RealAccountLedger || leg1.Flags != expectedLinkedFlag {
		t.Errorf("leg 1 ledger/flags mismatch: ledger=%d flags=%d", leg1.Ledger, leg1.Flags)
	}
	if leg2.DebitAccountID != constants.GetVirtualLedgerLiquidityBookId() || leg2.CreditAccountID != tb.BytesToUint128(budgetID) {
		t.Errorf("leg 2 debit/credit mismatch")
	}
	if leg2.Ledger != constants.VirtualAccountLedger || leg2.Flags != 0 {
		t.Errorf("leg 2 ledger/flags mismatch: ledger=%d flags=%d", leg2.Ledger, leg2.Flags)
	}
}

func TestAccountToBudget_Errors(t *testing.T) {
	ctx := context.Background()
	accID, budgetID := testUUID(20), testUUID(21)

	tests := []struct {
		name       string
		accID      [16]byte
		budgetID   [16]byte
		amount     uint64
		setup      func(store *fakeAccountStore, tbClient *fakeTBClient)
		wantStatus constants.TransactionStatus
	}{
		{
			name:       "zero amount",
			accID:      accID,
			budgetID:   budgetID,
			amount:     0,
			setup:      func(_ *fakeAccountStore, _ *fakeTBClient) {},
			wantStatus: constants.TransactionInvalidAmount,
		},
		{
			name:       "same UUID",
			accID:      accID,
			budgetID:   accID,
			amount:     100,
			setup:      func(_ *fakeAccountStore, _ *fakeTBClient) {},
			wantStatus: constants.TransactionSameAccount,
		},
		{
			name:     "account is virtual instead of real",
			accID:    accID,
			budgetID: budgetID,
			amount:   100,
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				seedLiquidityPools(store)
				store.accounts[accID] = makeAccount(accID, db.AccountTypeVIRTUAL, false)
				store.accounts[budgetID] = makeAccount(budgetID, db.AccountTypeVIRTUAL, false)
			},
			wantStatus: constants.TransactionInvalidAccountType,
		},
		{
			name:     "budget not found",
			accID:    accID,
			budgetID: budgetID,
			amount:   100,
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				seedLiquidityPools(store)
				store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
			},
			wantStatus: constants.TransactionPayeeNotFound,
		},
		{
			name:     "budget archived",
			accID:    accID,
			budgetID: budgetID,
			amount:   100,
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				seedLiquidityPools(store)
				store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
				store.accounts[budgetID] = makeAccount(budgetID, db.AccountTypeVIRTUAL, true)
			},
			wantStatus: constants.TransactionPayeeArchived,
		},
		{
			name:     "budget is real instead of virtual",
			accID:    accID,
			budgetID: budgetID,
			amount:   100,
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				seedLiquidityPools(store)
				store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
				store.accounts[budgetID] = makeAccount(budgetID, db.AccountTypeREAL, false)
			},
			wantStatus: constants.TransactionInvalidAccountType,
		},
		{
			name:     "leg 0 insufficient funds fails linked chain",
			accID:    accID,
			budgetID: budgetID,
			amount:   100,
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				seedLiquidityPools(store)
				store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
				store.accounts[budgetID] = makeAccount(budgetID, db.AccountTypeVIRTUAL, false)
				tbClient.createTransfersRes = []tb.CreateTransferResult{
					{Status: tb.TransferExceedsCredits},
					{Status: tb.TransferLinkedEventFailed},
				}
			},
			wantStatus: constants.TransactionInsufficientFunds,
		},
		{
			name:     "leg 1 insufficient funds fails linked chain",
			accID:    accID,
			budgetID: budgetID,
			amount:   100,
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				seedLiquidityPools(store)
				store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
				store.accounts[budgetID] = makeAccount(budgetID, db.AccountTypeVIRTUAL, false)
				tbClient.createTransfersRes = []tb.CreateTransferResult{
					{Status: tb.TransferLinkedEventFailed},
					{Status: tb.TransferExceedsDebits},
				}
			},
			wantStatus: constants.TransactionInsufficientFunds,
		},
		{
			name:     "tigerbeetle transport error",
			accID:    accID,
			budgetID: budgetID,
			amount:   100,
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				seedLiquidityPools(store)
				store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
				store.accounts[budgetID] = makeAccount(budgetID, db.AccountTypeVIRTUAL, false)
				tbClient.createTransfersErr = errors.New("network error")
			},
			wantStatus: constants.TransactionInternalError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := newFakeAccountStore()
			tbClient := &fakeTBClient{}
			tc.setup(store, tbClient)
			svc := MoneyModuleService{Tb: tbClient, Queries: store}

			txID, status, err := svc.AccountToBudget(ctx, tc.accID, tc.budgetID, tc.amount)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if txID != ([16]byte{}) {
				t.Fatalf("expected zero txID on error, got %x", txID)
			}
			if status != tc.wantStatus {
				t.Fatalf("expected status %s, got %s", tc.wantStatus, status)
			}
		})
	}
}

func TestAccountToBudget_Panics(t *testing.T) {
	ctx := context.Background()
	accID, budgetID := testUUID(20), testUUID(21)

	tests := []struct {
		name  string
		setup func(store *fakeAccountStore, tbClient *fakeTBClient)
	}{
		{
			name: "missing real liquidity pool in postgres",
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
				store.accounts[budgetID] = makeAccount(budgetID, db.AccountTypeVIRTUAL, false)
			},
		},
		{
			name: "missing virtual liquidity pool in postgres",
			setup: func(store *fakeAccountStore, _ *fakeTBClient) {
				realLP := constants.GetRealLedgerLiquidityBookId().Bytes()
				store.accounts[realLP] = makeAccount(realLP, db.AccountTypeREAL, false)
				store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
				store.accounts[budgetID] = makeAccount(budgetID, db.AccountTypeVIRTUAL, false)
			},
		},
		{
			name: "wrong result count",
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				seedLiquidityPools(store)
				store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
				store.accounts[budgetID] = makeAccount(budgetID, db.AccountTypeVIRTUAL, false)
				tbClient.createTransfersRes = []tb.CreateTransferResult{{Status: tb.TransferCreated}}
			},
		},
		{
			name: "linked failure with no root cause",
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				seedLiquidityPools(store)
				store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
				store.accounts[budgetID] = makeAccount(budgetID, db.AccountTypeVIRTUAL, false)
				tbClient.createTransfersRes = []tb.CreateTransferResult{
					{Status: tb.TransferLinkedEventFailed},
					{Status: tb.TransferLinkedEventFailed},
				}
			},
		},
		{
			name: "desync on leg 2",
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				seedLiquidityPools(store)
				store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
				store.accounts[budgetID] = makeAccount(budgetID, db.AccountTypeVIRTUAL, false)
				tbClient.createTransfersRes = []tb.CreateTransferResult{
					{Status: tb.TransferLinkedEventFailed},
					{Status: tb.TransferCreditAccountNotFound},
				}
			},
		},
		{
			name: "linked event chain open",
			setup: func(store *fakeAccountStore, tbClient *fakeTBClient) {
				seedLiquidityPools(store)
				store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
				store.accounts[budgetID] = makeAccount(budgetID, db.AccountTypeVIRTUAL, false)
				tbClient.createTransfersRes = []tb.CreateTransferResult{
					{Status: tb.TransferLinkedEventFailed},
					{Status: tb.TransferLinkedEventChainOpen},
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := newFakeAccountStore()
			tbClient := &fakeTBClient{}
			tc.setup(store, tbClient)
			svc := MoneyModuleService{Tb: tbClient, Queries: store}

			assertPanics(t, func() {
				_, _, _ = svc.AccountToBudget(ctx, accID, budgetID, 100)
			})
		})
	}
}

// --- GetAllAccountTransfers / GetAllBudgetTransfers Tests ---

func TestGetAllTransfers(t *testing.T) {
	ctx := context.Background()
	accID, budgetID := testUUID(30), testUUID(31)

	t.Run("account transfers success", func(t *testing.T) {
		store := newFakeAccountStore()
		store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
		wantTransfers := []tb.Transfer{{ID: tb.ID(), Amount: tb.ToUint128(50)}}
		tbClient := &fakeTBClient{accountTransfersRes: wantTransfers}
		svc := MoneyModuleService{Tb: tbClient, Queries: store}

		got, status, err := svc.GetAllAccountTransfers(ctx, accID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != constants.TransactionSuccess {
			t.Fatalf("expected %s, got %s", constants.TransactionSuccess, status)
		}
		if len(got) != 1 || got[0].ID != wantTransfers[0].ID {
			t.Errorf("unexpected transfers returned: %+v", got)
		}
		if tbClient.lastAccountFilter.AccountID != tb.BytesToUint128(accID) {
			t.Errorf("unexpected AccountFilter.AccountID")
		}
		flags := tbClient.lastAccountFilter.AccountFilterFlags()
		if !flags.Debits || !flags.Credits || tbClient.lastAccountFilter.Limit == 0 {
			t.Errorf("expected Debits=true, Credits=true, Limit>0; got %+v limit=%d", flags, tbClient.lastAccountFilter.Limit)
		}
	})

	t.Run("budget transfers success", func(t *testing.T) {
		store := newFakeAccountStore()
		store.accounts[budgetID] = makeAccount(budgetID, db.AccountTypeVIRTUAL, false)
		tbClient := &fakeTBClient{accountTransfersRes: []tb.Transfer{{ID: tb.ID()}}}
		svc := MoneyModuleService{Tb: tbClient, Queries: store}

		got, status, err := svc.GetAllBudgetTransfers(ctx, budgetID)
		if err != nil || status != constants.TransactionSuccess || len(got) != 1 {
			t.Fatalf("unexpected result: got=%d status=%s err=%v", len(got), status, err)
		}
	})

	t.Run("account not found", func(t *testing.T) {
		svc := MoneyModuleService{Tb: &fakeTBClient{}, Queries: newFakeAccountStore()}
		_, status, err := svc.GetAllAccountTransfers(ctx, accID)
		if err == nil || status != constants.TransactionAccountNotFound {
			t.Fatalf("expected TransactionAccountNotFound, got (%s, %v)", status, err)
		}
	})

	t.Run("account archived", func(t *testing.T) {
		store := newFakeAccountStore()
		store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, true)
		svc := MoneyModuleService{Tb: &fakeTBClient{}, Queries: store}
		_, status, err := svc.GetAllAccountTransfers(ctx, accID)
		if err == nil || status != constants.TransactionAccountArchived {
			t.Fatalf("expected TransactionAccountArchived, got (%s, %v)", status, err)
		}
	})

	t.Run("wrong account type rejected", func(t *testing.T) {
		store := newFakeAccountStore()
		store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
		svc := MoneyModuleService{Tb: &fakeTBClient{}, Queries: store}
		_, status, err := svc.GetAllBudgetTransfers(ctx, accID)
		if err == nil || status != constants.TransactionInvalidAccountType {
			t.Fatalf("expected TransactionInvalidAccountType, got (%s, %v)", status, err)
		}
	})

	t.Run("tigerbeetle error", func(t *testing.T) {
		store := newFakeAccountStore()
		store.accounts[accID] = makeAccount(accID, db.AccountTypeREAL, false)
		tbClient := &fakeTBClient{accountTransfersErr: errors.New("tb down")}
		svc := MoneyModuleService{Tb: tbClient, Queries: store}
		_, status, err := svc.GetAllAccountTransfers(ctx, accID)
		if err == nil || status != constants.TransactionInternalError {
			t.Fatalf("expected TransactionInternalError, got (%s, %v)", status, err)
		}
	})
}
