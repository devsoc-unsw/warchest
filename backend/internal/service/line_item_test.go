package service

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"backend/internal/domain"
	"backend/internal/repository"
)

// in-memory repo, mimics insert defaults and the is_active filter on lists
type fakeLineItemRepo struct {
	items  map[int64]domain.LineItem
	nextID int64
	clock  int64 // bumped on every write, stands in for updated_at = NOW()
	err    error // when set, every method fails with it
}

func (f *fakeLineItemRepo) now() time.Time {
	f.clock++
	return time.Unix(f.clock, 0)
}

var _ repository.LineItemRepository = (*fakeLineItemRepo)(nil)

func newFakeLineItemRepo() *fakeLineItemRepo {
	return &fakeLineItemRepo{items: map[int64]domain.LineItem{}, nextID: 1}
}

func (f *fakeLineItemRepo) seed(items ...domain.LineItem) {
	for _, item := range items {
		f.items[item.ID] = item
		if item.ID >= f.nextID {
			f.nextID = item.ID + 1
		}
	}
}

func (f *fakeLineItemRepo) sorted(
	match func(domain.LineItem) bool,
) []domain.LineItem {
	var out []domain.LineItem
	for _, item := range f.items {
		if match(item) {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (f *fakeLineItemRepo) CreateLineItem(
	_ context.Context,
	arg repository.CreateLineItemParams,
) (domain.LineItem, error) {
	if f.err != nil {
		return domain.LineItem{}, f.err
	}
	item := domain.LineItem{
		ID:                   f.nextID,
		EstimatedQuantity:    arg.EstimatedQuantity,
		EstimatedCostPerItem: arg.EstimatedCostPerItem,
		EstimatedUnitCost:    arg.EstimatedUnitCost,
		PrDescription:        arg.PrDescription,
		PurchaseRequestID:    arg.PurchaseRequestID,
		Status:               domain.LineItemStatusPrDraft, // column default
		IsActive:             true,                         // column default
	}
	f.nextID++
	f.items[item.ID] = item
	return item, nil
}

func (f *fakeLineItemRepo) GetLineItemByID(
	_ context.Context,
	id int64,
) (domain.LineItem, error) {
	if f.err != nil {
		return domain.LineItem{}, f.err
	}
	item, ok := f.items[id]
	if !ok {
		return domain.LineItem{}, repository.ErrLineItemNotFound
	}
	return item, nil
}

func (f *fakeLineItemRepo) GetLineItemsByPRID(
	_ context.Context,
	prID int64,
) ([]domain.LineItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.sorted(func(item domain.LineItem) bool {
		return item.PurchaseRequestID == prID && item.IsActive
	}), nil
}

func (f *fakeLineItemRepo) GetLineItemsByReimbID(
	_ context.Context,
	reimbID int64,
) ([]domain.LineItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.sorted(func(item domain.LineItem) bool {
		return item.ReimbursementRequestID != nil && item.IsActive &&
			*item.ReimbursementRequestID == reimbID
	}), nil
}

// mimics the WHERE updated_at = read version guard, and the SET list:
// purchase_request_ID and created_at are not written
func (f *fakeLineItemRepo) UpdateLineItem(
	_ context.Context,
	item domain.LineItem,
) (domain.LineItem, error) {
	if f.err != nil {
		return domain.LineItem{}, f.err
	}
	stored, ok := f.items[item.ID]
	if !ok || !stored.UpdatedAt.Equal(item.UpdatedAt) {
		return domain.LineItem{}, repository.ErrLineItemConflict
	}
	stored.EstimatedQuantity = item.EstimatedQuantity
	stored.EstimatedCostPerItem = item.EstimatedCostPerItem
	stored.EstimatedUnitCost = item.EstimatedUnitCost
	stored.PrDescription = item.PrDescription
	stored.ActualQuantity = item.ActualQuantity
	stored.ActualCostPerItem = item.ActualCostPerItem
	stored.ActualUnitCost = item.ActualUnitCost
	stored.ReimbDescription = item.ReimbDescription
	stored.ReimbursementRequestID = item.ReimbursementRequestID
	stored.Status = item.Status
	stored.IsActive = item.IsActive
	stored.UpdatedAt = f.now()
	f.items[item.ID] = stored
	return stored, nil
}

// mimics the WHERE status = from AND is_active guard
func (f *fakeLineItemRepo) TransitionLineItem(
	_ context.Context,
	id int64,
	from, to domain.LineItemStatus,
) (domain.LineItem, error) {
	if f.err != nil {
		return domain.LineItem{}, f.err
	}
	item, ok := f.items[id]
	if !ok || item.Status != from || !item.IsActive {
		return domain.LineItem{}, repository.ErrLineItemConflict
	}
	item.Status = to
	item.UpdatedAt = f.now()
	f.items[id] = item
	return item, nil
}

func (f *fakeLineItemRepo) ArchiveLineItem(_ context.Context, id int64) error {
	if f.err != nil {
		return f.err
	}
	if item, ok := f.items[id]; ok {
		item.IsActive = false
		item.UpdatedAt = f.now()
		f.items[id] = item
	}
	return nil
}

func validCreate() CreateLineItemInput {
	return CreateLineItemInput{
		PurchaseRequestID:    7,
		EstimatedQuantity:    2,
		EstimatedCostPerItem: 1000,
		EstimatedUnitCost:    100,
		Description:          "two bags of paper cups",
	}
}

func TestCreateAppliesColumnDefaults(t *testing.T) {
	svc := NewLineItemService(newFakeLineItemRepo())

	item, err := svc.Create(context.Background(), validCreate())
	if err != nil {
		t.Fatalf("Create returned %v, want nil", err)
	}
	if item.Status != domain.LineItemStatusPrDraft {
		t.Errorf("status = %q, want %q", item.Status, domain.LineItemStatusPrDraft)
	}
	if !item.IsActive {
		t.Error("IsActive = false, want true")
	}
}

func TestCreateStoresSuppliedValues(t *testing.T) {
	svc := NewLineItemService(newFakeLineItemRepo())
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
	if item.PrDescription == nil || *item.PrDescription != in.Description {
		t.Errorf("description = %v, want %q", item.PrDescription, in.Description)
	}
}

// empty description is stored as NULL, not an empty string
func TestCreateTreatsEmptyDescriptionAsNull(t *testing.T) {
	svc := NewLineItemService(newFakeLineItemRepo())
	in := validCreate()
	in.Description = ""

	item, err := svc.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("Create returned %v, want nil", err)
	}
	if item.PrDescription != nil {
		t.Errorf("description = %q, want NULL", *item.PrDescription)
	}
}

func TestCreateValidation(t *testing.T) {
	tests := map[string]func(*CreateLineItemInput){
		"zero purchase request": func(c *CreateLineItemInput) {
			c.PurchaseRequestID = 0
		},
		"negative purchase request": func(c *CreateLineItemInput) {
			c.PurchaseRequestID = -1
		},
		"zero quantity": func(c *CreateLineItemInput) {
			c.EstimatedQuantity = 0
		},
		"negative quantity": func(c *CreateLineItemInput) {
			c.EstimatedQuantity = -3
		},
		"negative unit cost": func(c *CreateLineItemInput) {
			c.EstimatedUnitCost = -1
		},
		"negative cost per item": func(c *CreateLineItemInput) {
			c.EstimatedCostPerItem = -1
		},
		"over long description": func(c *CreateLineItemInput) {
			c.Description = strings.Repeat("x", maxLineItemDescriptionLen+1)
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			repo := newFakeLineItemRepo()
			in := validCreate()
			mutate(&in)

			_, err := NewLineItemService(repo).Create(context.Background(), in)
			if err != nil {
				if !errors.Is(err, ErrInvalidLineItem) {
					t.Fatalf("Create returned %v, want ErrInvalidLineItem", err)
				}
			} else {
				t.Fatal("Create returned nil, want ErrInvalidLineItem")
			}
			if len(repo.items) != 0 {
				t.Errorf("wrote %d rows, want 0", len(repo.items))
			}
		})
	}
}

func TestCreateAcceptsZeroCosts(t *testing.T) {
	svc := NewLineItemService(newFakeLineItemRepo())
	in := validCreate()
	in.EstimatedCostPerItem = 0
	in.EstimatedUnitCost = 0

	if _, err := svc.Create(context.Background(), in); err != nil {
		t.Fatalf("Create returned %v, want nil: a free item is legitimate", err)
	}
}

func TestGetByIDMapsMissingRowToErrNotFound(t *testing.T) {
	svc := NewLineItemService(newFakeLineItemRepo())

	_, err := svc.GetByID(context.Background(), 404)
	if !errors.Is(err, ErrLineItemNotFound) {
		t.Fatalf("GetByID returned %v, want ErrLineItemNotFound", err)
	}
	if errors.Is(err, repository.ErrLineItemNotFound) {
		t.Error("GetByID leaked the repository error to the caller")
	}
}

func TestGetByIDReturnsItem(t *testing.T) {
	repo := newFakeLineItemRepo()
	repo.seed(domain.LineItem{ID: 5, PurchaseRequestID: 7, IsActive: true})

	item, err := NewLineItemService(repo).GetByID(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetByID returned %v, want nil", err)
	}
	if item.ID != 5 {
		t.Errorf("id = %d, want 5", item.ID)
	}
}

func TestListByPurchaseRequestExcludesOtherPRsAndArchived(t *testing.T) {
	repo := newFakeLineItemRepo()
	repo.seed(
		domain.LineItem{ID: 1, PurchaseRequestID: 7, IsActive: true},
		domain.LineItem{ID: 2, PurchaseRequestID: 7, IsActive: false},
		domain.LineItem{ID: 3, PurchaseRequestID: 8, IsActive: true},
	)

	items, err := NewLineItemService(repo).
		ListByPurchaseRequest(context.Background(), 7)
	if err != nil {
		t.Fatalf("ListByPurchaseRequest returned %v, want nil", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("got %d items %v, want only item 1", len(items), ids(items))
	}
}

func TestListByReimbursementRequest(t *testing.T) {
	repo := newFakeLineItemRepo()
	repo.seed(
		domain.LineItem{ID: 1, ReimbursementRequestID: p64(42), IsActive: true},
		domain.LineItem{ID: 2, ReimbursementRequestID: p64(99), IsActive: true},
		domain.LineItem{ID: 3, IsActive: true}, // not attached to any reimbursement
	)

	svc := NewLineItemService(repo)
	items, err := svc.ListByReimbursementRequest(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListByReimbursementRequest returned %v, want nil", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("got %d items %v, want only item 1", len(items), ids(items))
	}
}

func TestTotalsByPurchaseRequest(t *testing.T) {
	repo := newFakeLineItemRepo()
	repo.seed(
		// 2 x 1000 estimated, 2 x 1100 actual
		domain.LineItem{
			ID: 1, PurchaseRequestID: 7, IsActive: true,
			EstimatedQuantity: 2, EstimatedCostPerItem: 1000,
			ActualQuantity: p64(2), ActualCostPerItem: p64(1100),
		},
		// 3 x 500 estimated, no actuals recorded yet
		domain.LineItem{
			ID: 2, PurchaseRequestID: 7, IsActive: true,
			EstimatedQuantity: 3, EstimatedCostPerItem: 500,
		},
		// archived, must not count
		domain.LineItem{
			ID: 3, PurchaseRequestID: 7, IsActive: false,
			EstimatedQuantity: 100, EstimatedCostPerItem: 100,
		},
	)

	totals, err := NewLineItemService(repo).
		TotalsByPurchaseRequest(context.Background(), 7)
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

// quantity without cost contributes nothing to the actual total
func TestTotalsIgnoresPartiallyRecordedActuals(t *testing.T) {
	repo := newFakeLineItemRepo()
	repo.seed(domain.LineItem{
		ID: 1, PurchaseRequestID: 7, IsActive: true,
		EstimatedQuantity: 1, EstimatedCostPerItem: 900,
		ActualQuantity: p64(5), // cost per item still unset
	})

	totals, err := NewLineItemService(repo).
		TotalsByPurchaseRequest(context.Background(), 7)
	if err != nil {
		t.Fatalf("TotalsByPurchaseRequest returned %v, want nil", err)
	}
	if totals.ActualTotal != 0 {
		t.Errorf("actual total = %d, want 0", totals.ActualTotal)
	}
}

func TestTotalsOfEmptyPurchaseRequestIsZero(t *testing.T) {
	svc := NewLineItemService(newFakeLineItemRepo())

	totals, err := svc.TotalsByPurchaseRequest(context.Background(), 7)
	if err != nil {
		t.Fatalf("TotalsByPurchaseRequest returned %v, want nil", err)
	}
	if totals != (LineItemTotals{}) {
		t.Errorf("totals = %+v, want zero", totals)
	}
}

func TestReadsPropagateRepositoryErrors(t *testing.T) {
	repo := newFakeLineItemRepo()
	repo.err = errors.New("connection refused")
	svc := NewLineItemService(repo)
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

// the update writes the whole row, a field it dropped would get zeroed
func TestUpdateLineItemWritesEveryField(t *testing.T) {
	original := domain.LineItem{
		ID:                     1,
		PurchaseRequestID:      7,
		ReimbursementRequestID: p64(42),
		Status:                 domain.LineItemStatusReimbDraft,
		IsActive:               true,
		EstimatedQuantity:      2,
		EstimatedCostPerItem:   1000,
		EstimatedUnitCost:      100,
		PrDescription:          pstr("cups"),
		ActualQuantity:         p64(3),
		ActualCostPerItem:      p64(1100),
		ActualUnitCost:         p64(110),
		ReimbDescription:       pstr("receipt"),
	}
	repo := newFakeLineItemRepo()
	repo.seed(original)

	written, err := repo.UpdateLineItem(context.Background(), original)
	if err != nil {
		t.Fatalf("UpdateLineItem returned %v, want nil", err)
	}
	if !written.UpdatedAt.After(original.UpdatedAt) {
		t.Errorf("updated_at = %v, want it bumped past %v",
			written.UpdatedAt, original.UpdatedAt)
	}
	written.UpdatedAt = original.UpdatedAt
	if !reflect.DeepEqual(written, original) {
		t.Errorf("round trip changed the row:\n got %+v\nwant %+v", written, original)
	}
}

func p64(v int64) *int64    { return &v }
func pstr(s string) *string { return &s }

// one line item in the given status
func seedItem(
	repo *fakeLineItemRepo,
	status domain.LineItemStatus,
) domain.LineItem {
	item := domain.LineItem{
		ID: 1, PurchaseRequestID: 7, IsActive: true,
		EstimatedQuantity: 2, EstimatedCostPerItem: 1000, EstimatedUnitCost: 100,
		PrDescription: pstr("cups"),
		Status:        status,
	}
	// reimbursement phase items always have one, AttachReimbursementRequest sets it
	if IsLineItemReimbursementPhase(status) {
		item.ReimbursementRequestID = p64(42)
	}
	repo.seed(item)
	return item
}

func validEstimates() LineItemEstimatesInput {
	return LineItemEstimatesInput{
		EstimatedQuantity:    5,
		EstimatedCostPerItem: 200,
		EstimatedUnitCost:    20,
		Description:          "revised",
	}
}

func TestUpdateEstimatesAllowedOnlyInDraft(t *testing.T) {
	for _, status := range AllLineItemStatuses {
		t.Run(string(status), func(t *testing.T) {
			repo := newFakeLineItemRepo()
			seedItem(repo, status)
			svc := NewLineItemService(repo)

			_, err := svc.UpdateEstimates(context.Background(), 1, validEstimates())
			if status == domain.LineItemStatusPrDraft {
				if err != nil {
					t.Fatalf("UpdateEstimates returned %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, ErrImmutableLineItemField) {
				t.Fatalf(
					"UpdateEstimates returned %v, want ErrImmutableLineItemField",
					err,
				)
			}
		})
	}
}

func TestUpdateEstimatesAppliesValuesAndLeavesTheRestAlone(t *testing.T) {
	repo := newFakeLineItemRepo()
	seedItem(repo, domain.LineItemStatusPrDraft)
	repo.items[1] = withActuals(repo.items[1])

	in := validEstimates()
	item, err := NewLineItemService(repo).UpdateEstimates(context.Background(), 1, in)
	if err != nil {
		t.Fatalf("UpdateEstimates returned %v, want nil", err)
	}
	if item.EstimatedQuantity != in.EstimatedQuantity {
		t.Errorf("quantity = %d, want %d",
			item.EstimatedQuantity, in.EstimatedQuantity)
	}
	if item.PrDescription == nil || *item.PrDescription != in.Description {
		t.Errorf("description = %v, want %q", item.PrDescription, in.Description)
	}
	if item.Status != domain.LineItemStatusPrDraft {
		t.Errorf("status = %q, want it untouched", item.Status)
	}
	if item.ActualQuantity == nil || *item.ActualQuantity != 9 {
		t.Errorf("actual quantity = %v, want it untouched", item.ActualQuantity)
	}
}

func TestUpdateEstimatesValidation(t *testing.T) {
	repo := newFakeLineItemRepo()
	seedItem(repo, domain.LineItemStatusPrDraft)
	in := validEstimates()
	in.EstimatedQuantity = 0

	_, err := NewLineItemService(repo).UpdateEstimates(context.Background(), 1, in)
	if !errors.Is(err, ErrInvalidLineItem) {
		t.Fatalf("UpdateEstimates returned %v, want ErrInvalidLineItem", err)
	}
	if repo.items[1].EstimatedQuantity != 2 {
		t.Error("a rejected update still wrote to the row")
	}
}

func TestUpdateActualsAllowedOnlyInReimbDraft(t *testing.T) {
	in := LineItemActualsInput{ActualQuantity: p64(3)}

	for _, status := range AllLineItemStatuses {
		t.Run(string(status), func(t *testing.T) {
			repo := newFakeLineItemRepo()
			seedItem(repo, status)

			_, err := NewLineItemService(repo).
				UpdateActuals(context.Background(), 1, in)
			if status == domain.LineItemStatusReimbDraft {
				if err != nil {
					t.Fatalf("UpdateActuals returned %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, ErrImmutableLineItemField) {
				t.Fatalf(
					"UpdateActuals returned %v, want ErrImmutableLineItemField",
					err,
				)
			}
		})
	}
}

// updating one actual must not blank the others
func TestUpdateActualsAppliesOnlySuppliedFields(t *testing.T) {
	repo := newFakeLineItemRepo()
	seedItem(repo, domain.LineItemStatusReimbDraft)
	repo.items[1] = withActuals(repo.items[1])

	svc := NewLineItemService(repo)
	in := LineItemActualsInput{ActualCostPerItem: p64(1234)}
	item, err := svc.UpdateActuals(context.Background(), 1, in)
	if err != nil {
		t.Fatalf("UpdateActuals returned %v, want nil", err)
	}
	if item.ActualCostPerItem == nil || *item.ActualCostPerItem != 1234 {
		t.Errorf("cost per item = %v, want 1234", item.ActualCostPerItem)
	}
	if item.ActualQuantity == nil || *item.ActualQuantity != 9 {
		t.Errorf("quantity = %v, want 9 untouched", item.ActualQuantity)
	}
	if item.ReimbDescription == nil || *item.ReimbDescription != "receipt" {
		t.Errorf("description = %v, want it untouched", item.ReimbDescription)
	}
}

func TestUpdateActualsValidation(t *testing.T) {
	tests := map[string]LineItemActualsInput{
		"zero quantity":          {ActualQuantity: p64(0)},
		"negative quantity":      {ActualQuantity: p64(-1)},
		"negative cost per item": {ActualCostPerItem: p64(-1)},
		"negative unit cost":     {ActualUnitCost: p64(-1)},
		"over long description": {
			Description: pstr(strings.Repeat("x", maxLineItemDescriptionLen+1)),
		},
	}

	for name, in := range tests {
		t.Run(name, func(t *testing.T) {
			repo := newFakeLineItemRepo()
			seedItem(repo, domain.LineItemStatusReimbDraft)

			_, err := NewLineItemService(repo).
				UpdateActuals(context.Background(), 1, in)
			if !errors.Is(err, ErrInvalidLineItem) {
				t.Fatalf("UpdateActuals returned %v, want ErrInvalidLineItem", err)
			}
		})
	}
}

func TestAttachReimbursementRequestOnlyFromApproved(t *testing.T) {
	for _, status := range AllLineItemStatuses {
		t.Run(string(status), func(t *testing.T) {
			repo := newFakeLineItemRepo()
			seedItem(repo, status)
			svc := NewLineItemService(repo)

			item, err := svc.AttachReimbursementRequest(context.Background(), 1, 42)
			if status == domain.LineItemStatusPrApproved {
				if err != nil {
					t.Fatalf("AttachReimbursementRequest returned %v, want nil", err)
				}
				if item.Status != domain.LineItemStatusReimbDraft {
					t.Errorf("status = %q, want reimb_draft", item.Status)
				}
				if item.ReimbursementRequestID == nil ||
					*item.ReimbursementRequestID != 42 {
					t.Errorf("reimbursement id = %v, want 42",
						item.ReimbursementRequestID)
				}
				return
			}
			if !errors.Is(err, ErrInvalidLineItemTransition) {
				t.Fatalf("returned %v, want ErrInvalidLineItemTransition", err)
			}
		})
	}
}

func TestAttachReimbursementRequestValidatesID(t *testing.T) {
	repo := newFakeLineItemRepo()
	seedItem(repo, domain.LineItemStatusPrApproved)

	svc := NewLineItemService(repo)
	_, err := svc.AttachReimbursementRequest(context.Background(), 1, 0)
	if !errors.Is(err, ErrInvalidLineItem) {
		t.Fatalf("returned %v, want ErrInvalidLineItem", err)
	}
}

func TestArchiveClearsIsActiveAndKeepsStatus(t *testing.T) {
	repo := newFakeLineItemRepo()
	seedItem(repo, domain.LineItemStatusPrPending)

	if err := NewLineItemService(repo).Archive(context.Background(), 1); err != nil {
		t.Fatalf("Archive returned %v, want nil", err)
	}
	if repo.items[1].IsActive {
		t.Error("IsActive = true, want false")
	}
	if repo.items[1].Status != domain.LineItemStatusPrPending {
		t.Errorf("status = %q, want it untouched", repo.items[1].Status)
	}
}

func TestArchiveIsIdempotent(t *testing.T) {
	repo := newFakeLineItemRepo()
	seedItem(repo, domain.LineItemStatusPrDraft)
	svc := NewLineItemService(repo)
	ctx := context.Background()

	if err := svc.Archive(ctx, 1); err != nil {
		t.Fatalf("first Archive returned %v, want nil", err)
	}
	if err := svc.Archive(ctx, 1); err != nil {
		t.Fatalf("second Archive returned %v, want nil", err)
	}
}

func TestMutationsOnMissingItemReturnNotFound(t *testing.T) {
	svc := NewLineItemService(newFakeLineItemRepo())
	ctx := context.Background()

	_, err := svc.UpdateEstimates(ctx, 404, validEstimates())
	if !errors.Is(err, ErrLineItemNotFound) {
		t.Errorf("UpdateEstimates returned %v, want ErrLineItemNotFound", err)
	}
	_, err = svc.UpdateActuals(ctx, 404, LineItemActualsInput{})
	if !errors.Is(err, ErrLineItemNotFound) {
		t.Errorf("UpdateActuals returned %v, want ErrLineItemNotFound", err)
	}
	_, err = svc.AttachReimbursementRequest(ctx, 404, 42)
	if !errors.Is(err, ErrLineItemNotFound) {
		t.Errorf(
			"AttachReimbursementRequest returned %v, want ErrLineItemNotFound",
			err,
		)
	}
	if err = svc.Archive(ctx, 404); !errors.Is(err, ErrLineItemNotFound) {
		t.Errorf("Archive returned %v, want ErrLineItemNotFound", err)
	}
}

// every status pair through Transition, not just the table
func TestTransitionAllowsExactlyTheLegalEdges(t *testing.T) {
	for _, from := range AllLineItemStatuses {
		for _, to := range AllLineItemStatuses {
			t.Run(string(from)+"_to_"+string(to), func(t *testing.T) {
				repo := newFakeLineItemRepo()
				// actuals set so the reimb_pending guard doesn't interfere
				repo.seed(withActuals(seedItem(repo, from)))

				_, err := NewLineItemService(repo).
					Transition(context.Background(), 1, to)
				// legal edge, but only via AttachReimbursementRequest
				enteringReimb := from == domain.LineItemStatusPrApproved &&
					to == domain.LineItemStatusReimbDraft
				if !CanTransitionLineItem(from, to) || enteringReimb {
					if !errors.Is(err, ErrInvalidLineItemTransition) {
						t.Fatalf("returned %v, want ErrInvalidLineItemTransition", err)
					}
					return
				}
				if err != nil {
					t.Fatalf("Transition returned %v, want nil", err)
				}
				if repo.items[1].Status != to {
					t.Errorf("status = %q, want %q", repo.items[1].Status, to)
				}
			})
		}
	}
}

// empty claims can't be submitted
func TestTransitionToReimbPendingRequiresActuals(t *testing.T) {
	partial := map[string]func(domain.LineItem) domain.LineItem{
		"nothing recorded": func(item domain.LineItem) domain.LineItem {
			return item
		},
		"quantity only": func(item domain.LineItem) domain.LineItem {
			item.ActualQuantity = p64(3)
			return item
		},
		"missing unit cost": func(item domain.LineItem) domain.LineItem {
			item.ActualQuantity = p64(3)
			item.ActualCostPerItem = p64(900)
			return item
		},
	}

	for name, fill := range partial {
		t.Run(name, func(t *testing.T) {
			repo := newFakeLineItemRepo()
			repo.seed(fill(seedItem(repo, domain.LineItemStatusReimbDraft)))

			svc := NewLineItemService(repo)
			_, err := svc.Transition(
				context.Background(), 1, domain.LineItemStatusReimbPending,
			)
			if !errors.Is(err, ErrInvalidLineItem) {
				t.Fatalf("Transition returned %v, want ErrInvalidLineItem", err)
			}
			if repo.items[1].Status != domain.LineItemStatusReimbDraft {
				t.Error("a rejected transition still moved the status")
			}
		})
	}
}

func TestTransitionToReimbPendingWithCompleteActuals(t *testing.T) {
	repo := newFakeLineItemRepo()
	repo.seed(withActuals(seedItem(repo, domain.LineItemStatusReimbDraft)))

	svc := NewLineItemService(repo)
	item, err := svc.Transition(
		context.Background(), 1, domain.LineItemStatusReimbPending,
	)
	if err != nil {
		t.Fatalf("Transition returned %v, want nil", err)
	}
	if item.Status != domain.LineItemStatusReimbPending {
		t.Errorf("status = %q, want reimb_pending", item.Status)
	}
}

// archived items can't move through the lifecycle
func TestTransitionRejectsArchivedItem(t *testing.T) {
	repo := newFakeLineItemRepo()
	item := seedItem(repo, domain.LineItemStatusPrDraft)
	item.IsActive = false
	repo.seed(item)

	svc := NewLineItemService(repo)
	_, err := svc.Transition(
		context.Background(), 1, domain.LineItemStatusPrPending,
	)
	if !errors.Is(err, ErrInvalidLineItemTransition) {
		t.Fatalf("Transition returned %v, want ErrInvalidLineItemTransition", err)
	}
}

func TestTransitionLeavesOtherFieldsAlone(t *testing.T) {
	repo := newFakeLineItemRepo()
	before := withActuals(seedItem(repo, domain.LineItemStatusPrDraft))
	repo.seed(before)

	svc := NewLineItemService(repo)
	after, err := svc.Transition(
		context.Background(), 1, domain.LineItemStatusPrPending,
	)
	if err != nil {
		t.Fatalf("Transition returned %v, want nil", err)
	}

	before.Status = domain.LineItemStatusPrPending
	before.UpdatedAt = after.UpdatedAt
	if !reflect.DeepEqual(after, before) {
		t.Errorf("transition changed more than status:\n got %+v\nwant %+v",
			after, before)
	}
}

func TestTransitionOnMissingItemReturnsNotFound(t *testing.T) {
	svc := NewLineItemService(newFakeLineItemRepo())

	_, err := svc.Transition(
		context.Background(), 404, domain.LineItemStatusPrPending,
	)
	if !errors.Is(err, ErrLineItemNotFound) {
		t.Fatalf("Transition returned %v, want ErrLineItemNotFound", err)
	}
}

// fills in the reimbursement half
func withActuals(item domain.LineItem) domain.LineItem {
	item.ActualQuantity = p64(9)
	item.ActualCostPerItem = p64(900)
	item.ActualUnitCost = p64(90)
	item.ReimbDescription = pstr("receipt")
	return item
}

func ids(items []domain.LineItem) []int64 {
	out := make([]int64, len(items))
	for i, item := range items {
		out[i] = item.ID
	}
	return out
}

// spec written separately from lineItemTransitions, anything not here must be rejected
var legalEdges = map[[2]domain.LineItemStatus]bool{
	{domain.LineItemStatusPrDraft, domain.LineItemStatusPrPending}:          true,
	{domain.LineItemStatusPrPending, domain.LineItemStatusPrDraft}:          true,
	{domain.LineItemStatusPrPending, domain.LineItemStatusPrApproved}:       true,
	{domain.LineItemStatusPrPending, domain.LineItemStatusPrRejected}:       true,
	{domain.LineItemStatusPrApproved, domain.LineItemStatusReimbDraft}:      true,
	{domain.LineItemStatusReimbDraft, domain.LineItemStatusReimbPending}:    true,
	{domain.LineItemStatusReimbPending, domain.LineItemStatusReimbDraft}:    true,
	{domain.LineItemStatusReimbPending, domain.LineItemStatusReimbApproved}: true,
	{domain.LineItemStatusReimbPending, domain.LineItemStatusReimbRejected}: true,
}

var terminalStatuses = []domain.LineItemStatus{
	domain.LineItemStatusPrRejected,
	domain.LineItemStatusReimbApproved,
	domain.LineItemStatusReimbRejected,
}

// adding an edge to the table also needs adding to legalEdges
func TestCanTransitionExhaustive(t *testing.T) {
	for _, from := range AllLineItemStatuses {
		for _, to := range AllLineItemStatuses {
			want := legalEdges[[2]domain.LineItemStatus{from, to}]
			if got := CanTransitionLineItem(from, to); got != want {
				t.Errorf(
					"CanTransitionLineItem(%q, %q) = %v, want %v",
					from, to, got, want,
				)
			}
		}
	}
}

func TestCanTransitionRejectsSelfTransitions(t *testing.T) {
	for _, s := range AllLineItemStatuses {
		if CanTransitionLineItem(s, s) {
			t.Errorf("CanTransitionLineItem(%q, %q) = true, want false", s, s)
		}
	}
}

func TestCanTransitionRejectsUnknownStatus(t *testing.T) {
	unknown := domain.LineItemStatus("not_a_status")

	if CanTransitionLineItem(unknown, domain.LineItemStatusPrPending) {
		t.Error("CanTransitionLineItem from an unknown status = true, want false")
	}
	if CanTransitionLineItem(domain.LineItemStatusPrDraft, unknown) {
		t.Error("CanTransitionLineItem to an unknown status = true, want false")
	}
}

func TestTerminalStatusesHaveNoOutgoingEdges(t *testing.T) {
	for _, s := range terminalStatuses {
		if !IsTerminalLineItemStatus(s) {
			t.Errorf("IsTerminalLineItemStatus(%q) = false, want true", s)
		}
		if next := NextLineItemStates(s); len(next) != 0 {
			t.Errorf("NextLineItemStates(%q) = %v, want empty", s, next)
		}
	}
}

func TestNonTerminalStatusesHaveOutgoingEdges(t *testing.T) {
	terminal := make(map[domain.LineItemStatus]bool, len(terminalStatuses))
	for _, s := range terminalStatuses {
		terminal[s] = true
	}

	for _, s := range AllLineItemStatuses {
		if terminal[s] {
			continue
		}
		if IsTerminalLineItemStatus(s) {
			t.Errorf("IsTerminalLineItemStatus(%q) = true, want false", s)
		}
		if len(NextLineItemStates(s)) == 0 {
			t.Errorf("NextLineItemStates(%q) is empty, want at least one edge", s)
		}
	}
}

func TestIsTerminalRejectsUnknownStatus(t *testing.T) {
	if IsTerminalLineItemStatus(domain.LineItemStatus("not_a_status")) {
		t.Error("IsTerminalLineItemStatus(unknown) = true, want false")
	}
}

// NextLineItemStates and CanTransitionLineItem must agree
func TestNextStatesAgreesWithCanTransition(t *testing.T) {
	for _, from := range AllLineItemStatuses {
		for _, to := range NextLineItemStates(from) {
			if !CanTransitionLineItem(from, to) {
				t.Errorf(
					"NextLineItemStates(%q) offers %q, CanTransitionLineItem rejects",
					from, to,
				)
			}
		}
	}
}

// mutating the result must not change the table
func TestNextStatesReturnsCopy(t *testing.T) {
	const from = domain.LineItemStatusPrPending

	first := NextLineItemStates(from)
	if len(first) == 0 {
		t.Fatalf("NextLineItemStates(%q) is empty, cannot test copy semantics", from)
	}
	first[0] = domain.LineItemStatus("mutated")

	second := NextLineItemStates(from)
	if second[0] == domain.LineItemStatus("mutated") {
		t.Error(
			"NextLineItemStates returned a view of the transition table, " +
				"want a copy",
		)
	}
}

func TestIsReimbursementPhase(t *testing.T) {
	tests := map[domain.LineItemStatus]bool{
		domain.LineItemStatusPrDraft:        false,
		domain.LineItemStatusPrPending:      false,
		domain.LineItemStatusPrApproved:     false,
		domain.LineItemStatusPrRejected:     false,
		domain.LineItemStatusReimbDraft:     true,
		domain.LineItemStatusReimbPending:   true,
		domain.LineItemStatusReimbApproved:  true,
		domain.LineItemStatusReimbRejected:  true,
		domain.LineItemStatus("not_status"): false,
	}

	for status, want := range tests {
		if got := IsLineItemReimbursementPhase(status); got != want {
			t.Errorf(
				"IsLineItemReimbursementPhase(%q) = %v, want %v",
				status, got, want,
			)
		}
	}
}

// a status missing from AllLineItemStatuses would skip the exhaustive tests
func TestAllStatusesCoversTransitionTable(t *testing.T) {
	got, want := len(AllLineItemStatuses), len(lineItemTransitions)
	if got != want {
		t.Errorf(
			"AllLineItemStatuses has %d entries, transition table has %d",
			got, want,
		)
	}

	for _, s := range AllLineItemStatuses {
		if _, ok := lineItemTransitions[s]; !ok {
			t.Errorf("status %q in AllLineItemStatuses but not in transition table", s)
		}
	}
}

// another approver acts between our read and our write
type racingLineItemRepo struct {
	*fakeLineItemRepo
	interleave domain.LineItemStatus
}

func (r *racingLineItemRepo) TransitionLineItem(
	ctx context.Context,
	id int64,
	from, to domain.LineItemStatus,
) (domain.LineItem, error) {
	item := r.items[id]
	item.Status = r.interleave
	r.items[id] = item
	return r.fakeLineItemRepo.TransitionLineItem(ctx, id, from, to)
}

func TestTransitionLosesRaceReturnsConflict(t *testing.T) {
	fake := newFakeLineItemRepo()
	seedItem(fake, domain.LineItemStatusPrPending)
	repo := &racingLineItemRepo{fake, domain.LineItemStatusPrRejected}

	svc := NewLineItemService(repo)
	_, err := svc.Transition(
		context.Background(), 1, domain.LineItemStatusPrApproved,
	)
	if !errors.Is(err, ErrInvalidLineItemTransition) {
		t.Fatalf("Transition returned %v, want ErrInvalidLineItemTransition", err)
	}
	if got := fake.items[1].Status; got != domain.LineItemStatusPrRejected {
		t.Errorf("status = %q, want the other approver's %q to stand", got,
			domain.LineItemStatusPrRejected)
	}
}

// someone submits the item between our read and our full-row write
type racingUpdateRepo struct {
	*fakeLineItemRepo
	interleave domain.LineItemStatus
}

func (r *racingUpdateRepo) UpdateLineItem(
	ctx context.Context,
	item domain.LineItem,
) (domain.LineItem, error) {
	_, err := r.TransitionLineItem(
		ctx, item.ID, r.items[item.ID].Status, r.interleave,
	)
	if err != nil {
		return domain.LineItem{}, err
	}
	return r.fakeLineItemRepo.UpdateLineItem(ctx, item)
}

func TestFullRowWritesLoseRaceReturnConflict(t *testing.T) {
	tests := []struct {
		name       string
		start      domain.LineItemStatus
		interleave domain.LineItemStatus
		mutate     func(LineItemService) error
	}{
		{"UpdateEstimates", domain.LineItemStatusPrDraft,
			domain.LineItemStatusPrPending,
			func(s LineItemService) error {
				_, err := s.UpdateEstimates(context.Background(), 1, validEstimates())
				return err
			}},
		{"UpdateActuals", domain.LineItemStatusReimbDraft,
			domain.LineItemStatusReimbPending,
			func(s LineItemService) error {
				_, err := s.UpdateActuals(context.Background(), 1,
					LineItemActualsInput{ActualQuantity: p64(3)})
				return err
			}},
		{"AttachReimbursementRequest", domain.LineItemStatusPrApproved,
			domain.LineItemStatusReimbDraft,
			func(s LineItemService) error {
				_, err := s.AttachReimbursementRequest(context.Background(), 1, 42)
				return err
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := newFakeLineItemRepo()
			seedItem(fake, tt.start)
			svc := NewLineItemService(&racingUpdateRepo{fake, tt.interleave})

			if err := tt.mutate(svc); !errors.Is(err, ErrLineItemConflict) {
				t.Fatalf("returned %v, want ErrLineItemConflict", err)
			}
			got := fake.items[1]
			if got.Status != tt.interleave || !got.IsActive {
				t.Errorf("status = %q active = %v, want the concurrent %q to stand",
					got.Status, got.IsActive, tt.interleave)
			}
		})
	}
}

func TestTransitionCannotEnterReimbursementWithoutRequest(t *testing.T) {
	repo := newFakeLineItemRepo()
	seedItem(repo, domain.LineItemStatusPrApproved)

	_, err := NewLineItemService(repo).
		Transition(context.Background(), 1, domain.LineItemStatusReimbDraft)
	if !errors.Is(err, ErrInvalidLineItemTransition) {
		t.Fatalf("Transition returned %v, want ErrInvalidLineItemTransition", err)
	}
	if got := repo.items[1]; got.Status != domain.LineItemStatusPrApproved {
		t.Errorf("status = %q, want it left at %q", got.Status,
			domain.LineItemStatusPrApproved)
	}
}

func TestTransitionSoftRejectKeepsReimbursementRequest(t *testing.T) {
	repo := newFakeLineItemRepo()
	seedItem(repo, domain.LineItemStatusReimbPending)

	got, err := NewLineItemService(repo).
		Transition(context.Background(), 1, domain.LineItemStatusReimbDraft)
	if err != nil {
		t.Fatalf("Transition returned %v, want nil", err)
	}
	if got.ReimbursementRequestID == nil || *got.ReimbursementRequestID != 42 {
		t.Errorf("reimbursement request = %v, want it kept",
			got.ReimbursementRequestID)
	}
}

func TestMutationsOnArchivedItemAreRejected(t *testing.T) {
	tests := []struct {
		name   string
		status domain.LineItemStatus
		want   error
		mutate func(LineItemService) error
	}{
		{"UpdateEstimates", domain.LineItemStatusPrDraft, ErrImmutableLineItemField,
			func(s LineItemService) error {
				_, err := s.UpdateEstimates(context.Background(), 1, validEstimates())
				return err
			}},
		{"UpdateActuals", domain.LineItemStatusReimbDraft, ErrImmutableLineItemField,
			func(s LineItemService) error {
				_, err := s.UpdateActuals(context.Background(), 1,
					LineItemActualsInput{ActualQuantity: p64(3)})
				return err
			}},
		{"AttachReimbursementRequest", domain.LineItemStatusPrApproved,
			ErrInvalidLineItemTransition,
			func(s LineItemService) error {
				_, err := s.AttachReimbursementRequest(context.Background(), 1, 42)
				return err
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeLineItemRepo()
			item := seedItem(repo, tt.status)
			item.IsActive = false
			repo.seed(item)

			if err := tt.mutate(NewLineItemService(repo)); !errors.Is(err, tt.want) {
				t.Fatalf("returned %v, want %v", err, tt.want)
			}
			if !reflect.DeepEqual(repo.items[1], item) {
				t.Errorf("archived item changed:\n got %+v\nwant %+v",
					repo.items[1], item)
			}
		})
	}
}

func TestTransitionToUnknownStatusIsInvalidInput(t *testing.T) {
	repo := newFakeLineItemRepo()
	seedItem(repo, domain.LineItemStatusPrPending)

	_, err := NewLineItemService(repo).
		Transition(context.Background(), 1, "pr_aproved")
	if !errors.Is(err, ErrInvalidLineItem) {
		t.Fatalf("Transition returned %v, want ErrInvalidLineItem", err)
	}
}

func TestAmountsAreBounded(t *testing.T) {
	svc := NewLineItemService(newFakeLineItemRepo())
	ctx := context.Background()

	tooMany := validCreate()
	tooMany.EstimatedQuantity = maxLineItemQuantity + 1
	tooDear := validCreate()
	tooDear.EstimatedCostPerItem = maxLineItemCents + 1
	tooDearUnit := validCreate()
	tooDearUnit.EstimatedUnitCost = maxLineItemCents + 1
	atLimit := validCreate()
	atLimit.EstimatedQuantity = maxLineItemQuantity
	atLimit.EstimatedCostPerItem = maxLineItemCents

	for name, in := range map[string]CreateLineItemInput{
		"quantity": tooMany, "cost per item": tooDear, "unit cost": tooDearUnit,
	} {
		if _, err := svc.Create(ctx, in); !errors.Is(err, ErrInvalidLineItem) {
			t.Errorf("%s over limit: Create returned %v, want ErrInvalidLineItem",
				name, err)
		}
	}
	if _, err := svc.Create(ctx, atLimit); err != nil {
		t.Errorf("at limit: Create returned %v, want nil", err)
	}
}

func TestActualsAreBounded(t *testing.T) {
	repo := newFakeLineItemRepo()
	seedItem(repo, domain.LineItemStatusReimbDraft)
	svc := NewLineItemService(repo)

	for name, in := range map[string]LineItemActualsInput{
		"quantity":      {ActualQuantity: p64(maxLineItemQuantity + 1)},
		"cost per item": {ActualCostPerItem: p64(maxLineItemCents + 1)},
		"unit cost":     {ActualUnitCost: p64(maxLineItemCents + 1)},
	} {
		_, err := svc.UpdateActuals(context.Background(), 1, in)
		if !errors.Is(err, ErrInvalidLineItem) {
			t.Errorf(
				"%s over limit: UpdateActuals returned %v, want ErrInvalidLineItem",
				name, err,
			)
		}
	}
}

// what the response offers must be exactly what Transition accepts
func TestAllowedTransitionsAgreesWithTransition(t *testing.T) {
	variants := map[string]func(domain.LineItem) domain.LineItem{
		"active": func(item domain.LineItem) domain.LineItem { return item },
		"archived": func(item domain.LineItem) domain.LineItem {
			item.IsActive = false
			return item
		},
		"no reimbursement request": func(item domain.LineItem) domain.LineItem {
			item.ReimbursementRequestID = nil
			return item
		},
	}
	for _, from := range AllLineItemStatuses {
		for name, variant := range variants {
			for _, to := range AllLineItemStatuses {
				t.Run(string(from)+"_"+name+"_to_"+string(to), func(t *testing.T) {
					repo := newFakeLineItemRepo()
					// actuals set so the reimb_pending input check doesn't interfere
					item := variant(withActuals(seedItem(repo, from)))
					repo.seed(item)

					offered := false
					for _, next := range AllowedLineItemTransitions(item) {
						offered = offered || next == to
					}
					_, err := NewLineItemService(repo).
						Transition(context.Background(), 1, to)
					if offered && err != nil {
						t.Errorf("offered but Transition returned %v", err)
					}
					if !offered && err == nil {
						t.Error("not offered but Transition succeeded")
					}
				})
			}
		}
	}
}

func TestAllowedTransitionsOfApprovedItemNeedsAttach(t *testing.T) {
	repo := newFakeLineItemRepo()
	item := seedItem(repo, domain.LineItemStatusPrApproved)
	if got := AllowedLineItemTransitions(item); len(got) != 0 {
		t.Errorf("AllowedLineItemTransitions = %v, want none until attached", got)
	}
}
