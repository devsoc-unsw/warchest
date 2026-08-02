package lineitem

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"

	"backend/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// fakeRepo is an in-memory Repository. It reproduces the two behaviours of the
// real schema the service depends on: the column defaults applied on insert
// (pr_draft, is_active true), and the is_active filter on the list queries.
type fakeRepo struct {
	items  map[int64]db.LineItem
	nextID int64
	err    error // when set, every method fails with it
}

var _ Repository = (*fakeRepo)(nil)

func newFakeRepo() *fakeRepo {
	return &fakeRepo{items: map[int64]db.LineItem{}, nextID: 1}
}

func (f *fakeRepo) seed(items ...db.LineItem) {
	for _, item := range items {
		f.items[item.ID] = item
		if item.ID >= f.nextID {
			f.nextID = item.ID + 1
		}
	}
}

func (f *fakeRepo) sorted(match func(db.LineItem) bool) []db.LineItem {
	var out []db.LineItem
	for _, item := range f.items {
		if match(item) {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (f *fakeRepo) CreateLineItem(
	_ context.Context,
	arg db.CreateLineItemParams,
) (db.LineItem, error) {
	if f.err != nil {
		return db.LineItem{}, f.err
	}
	item := db.LineItem{
		ID:                   f.nextID,
		EstimatedQuantity:    arg.EstimatedQuantity,
		EstimatedCostPerItem: arg.EstimatedCostPerItem,
		EstimatedUnitCost:    arg.EstimatedUnitCost,
		PrDescription:        arg.PrDescription,
		PurchaseRequestID:    arg.PurchaseRequestID,
		Status:               db.LineItemStatusPrDraft, // column default
		IsActive:             true,                     // column default
	}
	f.nextID++
	f.items[item.ID] = item
	return item, nil
}

func (f *fakeRepo) GetLineItemByID(
	_ context.Context,
	id int64,
) (db.LineItem, error) {
	if f.err != nil {
		return db.LineItem{}, f.err
	}
	item, ok := f.items[id]
	if !ok {
		return db.LineItem{}, pgx.ErrNoRows
	}
	return item, nil
}

func (f *fakeRepo) GetLineItemsByPRID(
	_ context.Context,
	prID int64,
) ([]db.LineItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.sorted(func(item db.LineItem) bool {
		return item.PurchaseRequestID == prID && item.IsActive
	}), nil
}

func (f *fakeRepo) GetLineItemsByReimbID(
	_ context.Context,
	reimbID pgtype.Int8,
) ([]db.LineItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.sorted(func(item db.LineItem) bool {
		return item.ReimbursementRequestID.Valid && item.IsActive &&
			item.ReimbursementRequestID.Int64 == reimbID.Int64
	}), nil
}

func (f *fakeRepo) ListLineItems(
	_ context.Context,
	limit int32,
) ([]db.LineItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	all := f.sorted(func(db.LineItem) bool { return true })
	if int(limit) < len(all) {
		all = all[:limit]
	}
	return all, nil
}

func (f *fakeRepo) UpdateLineItem(
	_ context.Context,
	arg db.UpdateLineItemParams,
) (db.LineItem, error) {
	if f.err != nil {
		return db.LineItem{}, f.err
	}
	item, ok := f.items[arg.ID]
	if !ok {
		return db.LineItem{}, pgx.ErrNoRows
	}
	item.EstimatedQuantity = arg.EstimatedQuantity
	item.EstimatedCostPerItem = arg.EstimatedCostPerItem
	item.EstimatedUnitCost = arg.EstimatedUnitCost
	item.PrDescription = arg.PrDescription
	item.ActualQuantity = arg.ActualQuantity
	item.ActualCostPerItem = arg.ActualCostPerItem
	item.ActualUnitCost = arg.ActualUnitCost
	item.ReimbDescription = arg.ReimbDescription
	item.ReimbursementRequestID = arg.ReimbursementRequestID
	item.Status = arg.Status
	item.IsActive = arg.IsActive
	f.items[arg.ID] = item
	return item, nil
}

func (f *fakeRepo) DeleteLineItem(_ context.Context, id int64) error {
	if f.err != nil {
		return f.err
	}
	delete(f.items, id)
	return nil
}

func i8(v int64) pgtype.Int8 { return pgtype.Int8{Int64: v, Valid: true} }

func validCreate() CreateInput {
	return CreateInput{
		PurchaseRequestID:    7,
		EstimatedQuantity:    2,
		EstimatedCostPerItem: 1000,
		EstimatedUnitCost:    100,
		Description:          "two bags of paper cups",
	}
}

func TestCreateAppliesColumnDefaults(t *testing.T) {
	svc := NewService(newFakeRepo())

	item, err := svc.Create(context.Background(), validCreate())
	if err != nil {
		t.Fatalf("Create returned %v, want nil", err)
	}
	if item.Status != db.LineItemStatusPrDraft {
		t.Errorf("status = %q, want %q", item.Status, db.LineItemStatusPrDraft)
	}
	if !item.IsActive {
		t.Error("IsActive = false, want true")
	}
}

func TestCreateStoresSuppliedValues(t *testing.T) {
	svc := NewService(newFakeRepo())
	in := validCreate()

	item, err := svc.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("Create returned %v, want nil", err)
	}
	if item.EstimatedQuantity != in.EstimatedQuantity {
		t.Errorf("quantity = %d, want %d",
			item.EstimatedQuantity, in.EstimatedQuantity)
	}
	if item.PurchaseRequestID != in.PurchaseRequestID {
		t.Errorf("purchase request id = %d, want %d",
			item.PurchaseRequestID, in.PurchaseRequestID)
	}
	if !item.PrDescription.Valid || item.PrDescription.String != in.Description {
		t.Errorf("description = %+v, want %q", item.PrDescription, in.Description)
	}
}

// An absent description is stored as NULL rather than as an empty string, so
// "no description" has one representation instead of two.
func TestCreateTreatsEmptyDescriptionAsNull(t *testing.T) {
	svc := NewService(newFakeRepo())
	in := validCreate()
	in.Description = ""

	item, err := svc.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("Create returned %v, want nil", err)
	}
	if item.PrDescription.Valid {
		t.Errorf("description = %+v, want NULL", item.PrDescription)
	}
}

func TestCreateValidation(t *testing.T) {
	tests := map[string]func(*CreateInput){
		"zero purchase request":     func(c *CreateInput) { c.PurchaseRequestID = 0 },
		"negative purchase request": func(c *CreateInput) { c.PurchaseRequestID = -1 },
		"zero quantity":             func(c *CreateInput) { c.EstimatedQuantity = 0 },
		"negative quantity":         func(c *CreateInput) { c.EstimatedQuantity = -3 },
		"negative unit cost":        func(c *CreateInput) { c.EstimatedUnitCost = -1 },
		"negative cost per item": func(c *CreateInput) {
			c.EstimatedCostPerItem = -1
		},
		"over long description": func(c *CreateInput) {
			c.Description = strings.Repeat("x", maxDescriptionLen+1)
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			repo := newFakeRepo()
			in := validCreate()
			mutate(&in)

			if _, err := NewService(repo).Create(context.Background(), in); err != nil {
				if !errors.Is(err, ErrValidation) {
					t.Fatalf("Create returned %v, want ErrValidation", err)
				}
			} else {
				t.Fatal("Create returned nil, want ErrValidation")
			}
			if len(repo.items) != 0 {
				t.Errorf("wrote %d rows, want 0", len(repo.items))
			}
		})
	}
}

func TestCreateAcceptsZeroCosts(t *testing.T) {
	svc := NewService(newFakeRepo())
	in := validCreate()
	in.EstimatedCostPerItem = 0
	in.EstimatedUnitCost = 0

	if _, err := svc.Create(context.Background(), in); err != nil {
		t.Fatalf("Create returned %v, want nil: a free item is legitimate", err)
	}
}

func TestGetByIDMapsMissingRowToErrNotFound(t *testing.T) {
	svc := NewService(newFakeRepo())

	_, err := svc.GetByID(context.Background(), 404)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID returned %v, want ErrNotFound", err)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		t.Error("GetByID leaked pgx.ErrNoRows to the caller")
	}
}

func TestGetByIDReturnsItem(t *testing.T) {
	repo := newFakeRepo()
	repo.seed(db.LineItem{ID: 5, PurchaseRequestID: 7, IsActive: true})

	item, err := NewService(repo).GetByID(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetByID returned %v, want nil", err)
	}
	if item.ID != 5 {
		t.Errorf("id = %d, want 5", item.ID)
	}
}

func TestListByPurchaseRequestExcludesOtherPRsAndArchived(t *testing.T) {
	repo := newFakeRepo()
	repo.seed(
		db.LineItem{ID: 1, PurchaseRequestID: 7, IsActive: true},
		db.LineItem{ID: 2, PurchaseRequestID: 7, IsActive: false},
		db.LineItem{ID: 3, PurchaseRequestID: 8, IsActive: true},
	)

	items, err := NewService(repo).ListByPurchaseRequest(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListByPurchaseRequest returned %v, want nil", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("got %d items %v, want only item 1", len(items), ids(items))
	}
}

func TestListByReimbursementRequest(t *testing.T) {
	repo := newFakeRepo()
	repo.seed(
		db.LineItem{ID: 1, ReimbursementRequestID: i8(42), IsActive: true},
		db.LineItem{ID: 2, ReimbursementRequestID: i8(99), IsActive: true},
		db.LineItem{ID: 3, IsActive: true}, // not attached to any reimbursement
	)

	svc := NewService(repo)
	items, err := svc.ListByReimbursementRequest(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListByReimbursementRequest returned %v, want nil", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("got %d items %v, want only item 1", len(items), ids(items))
	}
}

func TestTotalsByPurchaseRequest(t *testing.T) {
	repo := newFakeRepo()
	repo.seed(
		// 2 x 1000 estimated, 2 x 1100 actual
		db.LineItem{
			ID: 1, PurchaseRequestID: 7, IsActive: true,
			EstimatedQuantity: 2, EstimatedCostPerItem: 1000,
			ActualQuantity: i8(2), ActualCostPerItem: i8(1100),
		},
		// 3 x 500 estimated, no actuals recorded yet
		db.LineItem{
			ID: 2, PurchaseRequestID: 7, IsActive: true,
			EstimatedQuantity: 3, EstimatedCostPerItem: 500,
		},
		// archived, must not count
		db.LineItem{
			ID: 3, PurchaseRequestID: 7, IsActive: false,
			EstimatedQuantity: 100, EstimatedCostPerItem: 100,
		},
	)

	totals, err := NewService(repo).TotalsByPurchaseRequest(context.Background(), 7)
	if err != nil {
		t.Fatalf("TotalsByPurchaseRequest returned %v, want nil", err)
	}
	if totals.EstimatedTotal != 3500 {
		t.Errorf("estimated total = %d, want 3500", totals.EstimatedTotal)
	}
	if totals.ActualTotal != 2200 {
		t.Errorf("actual total = %d, want 2200", totals.ActualTotal)
	}
}

// A line item with a quantity but no cost yet must contribute nothing, rather
// than contributing zero and understating the real spend.
func TestTotalsIgnoresPartiallyRecordedActuals(t *testing.T) {
	repo := newFakeRepo()
	repo.seed(db.LineItem{
		ID: 1, PurchaseRequestID: 7, IsActive: true,
		EstimatedQuantity: 1, EstimatedCostPerItem: 900,
		ActualQuantity: i8(5), // cost per item still unset
	})

	totals, err := NewService(repo).TotalsByPurchaseRequest(context.Background(), 7)
	if err != nil {
		t.Fatalf("TotalsByPurchaseRequest returned %v, want nil", err)
	}
	if totals.ActualTotal != 0 {
		t.Errorf("actual total = %d, want 0", totals.ActualTotal)
	}
}

func TestTotalsOfEmptyPurchaseRequestIsZero(t *testing.T) {
	svc := NewService(newFakeRepo())

	totals, err := svc.TotalsByPurchaseRequest(context.Background(), 7)
	if err != nil {
		t.Fatalf("TotalsByPurchaseRequest returned %v, want nil", err)
	}
	if totals != (Totals{}) {
		t.Errorf("totals = %+v, want zero", totals)
	}
}

func TestReadsPropagateRepositoryErrors(t *testing.T) {
	repo := newFakeRepo()
	repo.err = errors.New("connection refused")
	svc := NewService(repo)
	ctx := context.Background()

	if _, err := svc.GetByID(ctx, 1); !errors.Is(err, repo.err) {
		t.Errorf("GetByID returned %v, want the repository error", err)
	}
	if _, err := svc.ListByPurchaseRequest(ctx, 7); !errors.Is(err, repo.err) {
		t.Errorf("ListByPurchaseRequest returned %v, want the repository error", err)
	}
	if _, err := svc.TotalsByPurchaseRequest(ctx, 7); !errors.Is(err, repo.err) {
		t.Errorf("TotalsByPurchaseRequest returned %v, want the repository error", err)
	}
}

// TestParamsFromRoundTripsEveryColumn is the guarantee the whole phase 4
// design rests on: because UpdateLineItem writes all twelve columns, a field
// missing from paramsFrom would be silently zeroed on every update. Writing a
// fully populated row back through it must change nothing.
func TestParamsFromRoundTripsEveryColumn(t *testing.T) {
	original := db.LineItem{
		ID:                     1,
		EstimatedQuantity:      2,
		EstimatedCostPerItem:   1000,
		EstimatedUnitCost:      100,
		PrDescription:          pgtype.Text{String: "cups", Valid: true},
		ActualQuantity:         i8(3),
		ActualCostPerItem:      i8(1100),
		ActualUnitCost:         i8(110),
		ReimbDescription:       pgtype.Text{String: "receipt", Valid: true},
		ReimbursementRequestID: i8(42),
		PurchaseRequestID:      7,
		Status:                 db.LineItemStatusReimbDraft,
		IsActive:               true,
	}
	repo := newFakeRepo()
	repo.seed(original)

	written, err := repo.UpdateLineItem(context.Background(), paramsFrom(original))
	if err != nil {
		t.Fatalf("UpdateLineItem returned %v, want nil", err)
	}
	if written != original {
		t.Errorf("round trip changed the row:\n got %+v\nwant %+v", written, original)
	}
}

func ids(items []db.LineItem) []int64 {
	out := make([]int64, len(items))
	for i, item := range items {
		out[i] = item.ID
	}
	return out
}
