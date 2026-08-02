package lineitem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/db"
	service "backend/internal/service/lineitem"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// fakeService stands in for the real service. These tests are about HTTP
// concerns only: what is parsed, the shape of the response, and which status
// code a given service error becomes. Whether the rules themselves are right
// is settled in the service package, against its own fake repository.
type fakeService struct {
	item   db.LineItem
	items  []db.LineItem
	totals service.Totals
	err    error

	gotCreate     service.CreateInput
	gotEstimates  service.EstimatesInput
	gotActuals    service.ActualsInput
	gotStatus     db.LineItemStatus
	gotReimbID    int64
	gotID         int64
	archiveCalled bool
}

var _ service.Service = (*fakeService)(nil)

func (f *fakeService) Create(
	_ context.Context,
	in service.CreateInput,
) (db.LineItem, error) {
	f.gotCreate = in
	return f.item, f.err
}

func (f *fakeService) GetByID(_ context.Context, id int64) (db.LineItem, error) {
	f.gotID = id
	return f.item, f.err
}

func (f *fakeService) ListByPurchaseRequest(
	_ context.Context,
	prID int64,
) ([]db.LineItem, error) {
	f.gotID = prID
	return f.items, f.err
}

func (f *fakeService) ListByReimbursementRequest(
	_ context.Context,
	reimbID int64,
) ([]db.LineItem, error) {
	f.gotReimbID = reimbID
	return f.items, f.err
}

func (f *fakeService) TotalsByPurchaseRequest(
	_ context.Context,
	prID int64,
) (service.Totals, error) {
	f.gotID = prID
	return f.totals, f.err
}

func (f *fakeService) UpdateEstimates(
	_ context.Context,
	id int64,
	in service.EstimatesInput,
) (db.LineItem, error) {
	f.gotID, f.gotEstimates = id, in
	return f.item, f.err
}

func (f *fakeService) UpdateActuals(
	_ context.Context,
	id int64,
	in service.ActualsInput,
) (db.LineItem, error) {
	f.gotID, f.gotActuals = id, in
	return f.item, f.err
}

func (f *fakeService) AttachReimbursementRequest(
	_ context.Context,
	id, reimbID int64,
) (db.LineItem, error) {
	f.gotID, f.gotReimbID = id, reimbID
	return f.item, f.err
}

func (f *fakeService) Transition(
	_ context.Context,
	id int64,
	to db.LineItemStatus,
) (db.LineItem, error) {
	f.gotID, f.gotStatus = id, to
	return f.item, f.err
}

func (f *fakeService) Archive(_ context.Context, id int64) error {
	f.gotID, f.archiveCalled = id, true
	return f.err
}

func newTestServer(svc service.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router, NewHandler(svc))
	return router
}

func do(
	t *testing.T,
	router *gin.Engine,
	method, path, body string,
) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeItem(t *testing.T, rec *httptest.ResponseRecorder) lineItemResponse {
	t.Helper()
	var res lineItemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("decoding %s: %v", rec.Body.String(), err)
	}
	return res
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, want, rec.Body)
	}
}

func i8(v int64) pgtype.Int8 { return pgtype.Int8{Int64: v, Valid: true} }

func sampleItem() db.LineItem {
	return db.LineItem{
		ID: 1, PurchaseRequestID: 7, IsActive: true,
		EstimatedQuantity: 2, EstimatedCostPerItem: 1000,
		EstimatedUnitCost: 100,
		PrDescription:     pgtype.Text{String: "cups", Valid: true},
		Status:            db.LineItemStatusPrDraft,
	}
}

const createBody = `{
	"estimated_quantity": 2,
	"estimated_cost_per_item": 1000,
	"estimated_unit_cost": 100,
	"description": "two bags of paper cups"
}`

func TestCreateEndpoint(t *testing.T) {
	svc := &fakeService{item: sampleItem()}
	router := newTestServer(svc)

	rec := do(t, router, http.MethodPost, "/purchase-requests/7/line-items",
		createBody)
	assertStatus(t, rec, http.StatusCreated)

	if svc.gotCreate.PurchaseRequestID != 7 {
		t.Errorf("purchase request id = %d, want 7 taken from the path",
			svc.gotCreate.PurchaseRequestID)
	}
	if svc.gotCreate.EstimatedQuantity != 2 {
		t.Errorf("quantity = %d, want 2", svc.gotCreate.EstimatedQuantity)
	}
	if item := decodeItem(t, rec); item.EstimatedTotal != 2000 {
		t.Errorf("estimated total = %d, want 2000", item.EstimatedTotal)
	}
}

func TestRequestsRejectedBeforeReachingTheService(t *testing.T) {
	tests := map[string]struct {
		method, path, body string
	}{
		"malformed json": {
			http.MethodPost, "/purchase-requests/7/line-items",
			`{"estimated_quantity":`,
		},
		"non numeric path id": {
			http.MethodPost, "/purchase-requests/abc/line-items", createBody,
		},
		"negative path id": {
			http.MethodPost, "/purchase-requests/-1/line-items", createBody,
		},
		"zero item id": {
			http.MethodGet, "/line-items/0", "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			svc := &fakeService{item: sampleItem()}
			rec := do(t, newTestServer(svc), tc.method, tc.path, tc.body)

			assertStatus(t, rec, http.StatusBadRequest)
			if svc.gotID != 0 || svc.gotCreate.PurchaseRequestID != 0 {
				t.Error("the service was called despite a bad request")
			}
		})
	}
}

// The status code is decided in one place, from the error alone.
func TestServiceErrorsMapToStatusCodes(t *testing.T) {
	tests := map[string]struct {
		err  error
		want int
	}{
		"not found":          {service.ErrNotFound, http.StatusNotFound},
		"validation":         {service.ErrValidation, http.StatusBadRequest},
		"invalid transition": {service.ErrInvalidTransition, http.StatusConflict},
		"immutable field":    {service.ErrImmutableField, http.StatusConflict},
		"unrecognised":       {io.ErrUnexpectedEOF, http.StatusInternalServerError},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			router := newTestServer(&fakeService{err: tc.err})
			rec := do(t, router, http.MethodGet, "/line-items/1", "")
			assertStatus(t, rec, tc.want)
		})
	}
}

// The real service always wraps its sentinels with detail, so the mapping has
// to look through the wrapping.
func TestWrappedServiceErrorsStillMap(t *testing.T) {
	wrapped := fmt.Errorf(
		"%w: line item 1 is %q", service.ErrImmutableField, "pr_approved",
	)
	svc := &fakeService{err: wrapped}

	rec := do(t, newTestServer(svc), http.MethodGet, "/line-items/1", "")
	assertStatus(t, rec, http.StatusConflict)

	if !strings.Contains(rec.Body.String(), "pr_approved") {
		t.Errorf("body %s dropped the detail the service supplied", rec.Body)
	}
}

// An unrecognised failure must not tell the client what broke.
func TestInternalErrorDetailIsWithheld(t *testing.T) {
	svc := &fakeService{err: errors.New("dial tcp 10.0.0.1:5432: refused")}

	rec := do(t, newTestServer(svc), http.MethodGet, "/line-items/1", "")
	assertStatus(t, rec, http.StatusInternalServerError)

	if strings.Contains(rec.Body.String(), "10.0.0.1") {
		t.Errorf("body %s leaked the underlying error", rec.Body)
	}
}

// The response must not expose pgtype's wire shape. An unset figure is null,
// not {"Int64":0,"Valid":false}.
func TestResponseRendersUnsetValuesAsNull(t *testing.T) {
	router := newTestServer(&fakeService{item: sampleItem()})

	rec := do(t, router, http.MethodGet, "/line-items/1", "")
	assertStatus(t, rec, http.StatusOK)

	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decoding %s: %v", rec.Body.String(), err)
	}
	for _, field := range []string{
		"actual_quantity", "actual_cost_per_item", "actual_unit_cost",
		"actual_total", "reimbursement_request_id", "reimb_description",
	} {
		if value, ok := raw[field]; !ok || value != nil {
			t.Errorf("%s = %v, want null", field, value)
		}
	}
}

func TestResponseDerivesActualTotal(t *testing.T) {
	item := sampleItem()
	item.ActualQuantity = i8(2)
	item.ActualCostPerItem = i8(1100)
	router := newTestServer(&fakeService{item: item})

	rec := do(t, router, http.MethodGet, "/line-items/1", "")
	assertStatus(t, rec, http.StatusOK)

	if got := decodeItem(t, rec); got.ActualTotal == nil ||
		*got.ActualTotal != 2200 {
		t.Errorf("actual total = %v, want 2200", got.ActualTotal)
	}
}

// Every response publishes what the item will accept next, so a client does
// not have to encode the lifecycle itself.
func TestResponsePublishesAllowedTransitions(t *testing.T) {
	item := sampleItem()
	item.Status = db.LineItemStatusPrPending
	router := newTestServer(&fakeService{item: item})

	rec := do(t, router, http.MethodGet, "/line-items/1", "")
	assertStatus(t, rec, http.StatusOK)

	got := decodeItem(t, rec)
	want := []string{
		string(db.LineItemStatusPrApproved),
		string(db.LineItemStatusPrRejected),
	}
	if len(got.AllowedTransitions) != len(want) {
		t.Fatalf("allowed = %v, want %v", got.AllowedTransitions, want)
	}
	for i, status := range want {
		if got.AllowedTransitions[i] != status {
			t.Errorf("allowed[%d] = %q, want %q",
				i, got.AllowedTransitions[i], status)
		}
	}
}

// An empty list marshals as [] rather than null, so clients can iterate it
// without a nil check.
func TestListEndpointReturnsArrayWhenEmpty(t *testing.T) {
	router := newTestServer(&fakeService{})

	rec := do(t, router, http.MethodGet, "/purchase-requests/7/line-items", "")
	assertStatus(t, rec, http.StatusOK)

	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Errorf("body = %s, want []", got)
	}
}

func TestListByReimbursementEndpoint(t *testing.T) {
	svc := &fakeService{items: []db.LineItem{sampleItem()}}
	router := newTestServer(svc)

	rec := do(t, router, http.MethodGet,
		"/reimbursement-requests/42/line-items", "")
	assertStatus(t, rec, http.StatusOK)

	if svc.gotReimbID != 42 {
		t.Errorf("reimbursement id = %d, want 42", svc.gotReimbID)
	}
}

func TestTotalsEndpoint(t *testing.T) {
	svc := &fakeService{
		totals: service.Totals{EstimatedTotal: 2000, ActualTotal: 2200},
	}
	router := newTestServer(svc)

	rec := do(t, router, http.MethodGet,
		"/purchase-requests/7/line-items/totals", "")
	assertStatus(t, rec, http.StatusOK)

	var totals totalsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &totals); err != nil {
		t.Fatalf("decoding %s: %v", rec.Body.String(), err)
	}
	if totals.EstimatedTotal != 2000 || totals.ActualTotal != 2200 {
		t.Errorf("totals = %+v, want {2000 2200}", totals)
	}
}

func TestUpdateEstimatesEndpointPassesInputThrough(t *testing.T) {
	svc := &fakeService{item: sampleItem()}
	router := newTestServer(svc)

	rec := do(t, router, http.MethodPatch, "/line-items/1/estimates",
		`{"estimated_quantity": 5, "estimated_cost_per_item": 200,
		  "estimated_unit_cost": 20, "description": "revised"}`)
	assertStatus(t, rec, http.StatusOK)

	if svc.gotEstimates.EstimatedQuantity != 5 {
		t.Errorf("quantity = %d, want 5", svc.gotEstimates.EstimatedQuantity)
	}
	if svc.gotEstimates.Description != "revised" {
		t.Errorf("description = %q, want revised", svc.gotEstimates.Description)
	}
}

// Omitted actuals must arrive at the service as nil, not as zero, since that
// is how it tells "leave this alone" from "set this to nothing".
func TestUpdateActualsOmittedFieldsArriveAsNil(t *testing.T) {
	svc := &fakeService{item: sampleItem()}
	router := newTestServer(svc)

	rec := do(t, router, http.MethodPatch, "/line-items/1/actuals",
		`{"actual_cost_per_item": 1234}`)
	assertStatus(t, rec, http.StatusOK)

	got := svc.gotActuals
	if got.ActualCostPerItem == nil || *got.ActualCostPerItem != 1234 {
		t.Errorf("cost per item = %v, want 1234", got.ActualCostPerItem)
	}
	if got.ActualQuantity != nil {
		t.Errorf("quantity = %d, want nil", *got.ActualQuantity)
	}
	if got.Description != nil {
		t.Errorf("description = %q, want nil", *got.Description)
	}
}

func TestTransitionEndpointPassesStatusThrough(t *testing.T) {
	svc := &fakeService{item: sampleItem()}
	router := newTestServer(svc)

	rec := do(t, router, http.MethodPost, "/line-items/1/transition",
		`{"status": "pr_pending"}`)
	assertStatus(t, rec, http.StatusOK)

	if svc.gotStatus != db.LineItemStatusPrPending {
		t.Errorf("status = %q, want pr_pending", svc.gotStatus)
	}
}

func TestAttachReimbursementEndpoint(t *testing.T) {
	item := sampleItem()
	item.Status = db.LineItemStatusReimbDraft
	item.ReimbursementRequestID = i8(42)
	svc := &fakeService{item: item}
	router := newTestServer(svc)

	rec := do(t, router, http.MethodPost, "/line-items/1/reimbursement",
		`{"reimbursement_request_id": 42}`)
	assertStatus(t, rec, http.StatusOK)

	if svc.gotReimbID != 42 {
		t.Errorf("reimbursement id = %d, want 42", svc.gotReimbID)
	}
	got := decodeItem(t, rec)
	if got.ReimbursementRequestID == nil || *got.ReimbursementRequestID != 42 {
		t.Errorf("response id = %v, want 42", got.ReimbursementRequestID)
	}
}

func TestArchiveEndpointReturnsNoContent(t *testing.T) {
	svc := &fakeService{}
	router := newTestServer(svc)

	rec := do(t, router, http.MethodDelete, "/line-items/1", "")
	assertStatus(t, rec, http.StatusNoContent)

	if !svc.archiveCalled {
		t.Error("Archive was not called")
	}
	if rec.Body.Len() != 0 {
		t.Errorf("body = %s, want empty", rec.Body)
	}
}
