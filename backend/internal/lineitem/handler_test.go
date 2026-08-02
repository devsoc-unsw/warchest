package lineitem

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/db"

	"github.com/gin-gonic/gin"
)

func newTestServer(repo *fakeRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router, NewHandler(NewService(repo)))
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

const createBody = `{
	"estimated_quantity": 2,
	"estimated_cost_per_item": 1000,
	"estimated_unit_cost": 100,
	"description": "two bags of paper cups"
}`

func TestCreateEndpoint(t *testing.T) {
	router := newTestServer(newFakeRepo())

	rec := do(t, router, http.MethodPost, "/purchase-requests/7/line-items",
		createBody)
	assertStatus(t, rec, http.StatusCreated)

	item := decodeItem(t, rec)
	if item.Status != string(db.LineItemStatusPrDraft) {
		t.Errorf("status = %q, want pr_draft", item.Status)
	}
	if item.PurchaseRequestID != 7 {
		t.Errorf("purchase request id = %d, want 7", item.PurchaseRequestID)
	}
	if item.EstimatedTotal != 2000 {
		t.Errorf("estimated total = %d, want 2000", item.EstimatedTotal)
	}
}

func TestCreateEndpointRejectsInvalidBody(t *testing.T) {
	tests := map[string]struct {
		path, body string
		want       int
	}{
		"validation failure": {
			path: "/purchase-requests/7/line-items",
			body: `{"estimated_quantity": 0}`,
			want: http.StatusBadRequest,
		},
		"malformed json": {
			path: "/purchase-requests/7/line-items",
			body: `{"estimated_quantity":`,
			want: http.StatusBadRequest,
		},
		"non numeric path id": {
			path: "/purchase-requests/abc/line-items",
			body: createBody,
			want: http.StatusBadRequest,
		},
		"negative path id": {
			path: "/purchase-requests/-1/line-items",
			body: createBody,
			want: http.StatusBadRequest,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			router := newTestServer(newFakeRepo())
			rec := do(t, router, http.MethodPost, tc.path, tc.body)
			assertStatus(t, rec, tc.want)
		})
	}
}

func TestGetEndpointNotFound(t *testing.T) {
	router := newTestServer(newFakeRepo())

	rec := do(t, router, http.MethodGet, "/line-items/404", "")
	assertStatus(t, rec, http.StatusNotFound)
}

// The response must not expose pgtype's wire shape. An unset figure is null,
// not {"Int64":0,"Valid":false}.
func TestResponseRendersUnsetValuesAsNull(t *testing.T) {
	repo := newFakeRepo()
	seedItem(repo, db.LineItemStatusPrDraft)
	router := newTestServer(repo)

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

// An empty list marshals as [] rather than null, so clients can iterate it
// without a nil check.
func TestListEndpointReturnsArrayWhenEmpty(t *testing.T) {
	router := newTestServer(newFakeRepo())

	rec := do(t, router, http.MethodGet, "/purchase-requests/7/line-items", "")
	assertStatus(t, rec, http.StatusOK)

	if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
		t.Errorf("body = %s, want []", got)
	}
}

func TestTotalsEndpoint(t *testing.T) {
	repo := newFakeRepo()
	repo.seed(db.LineItem{
		ID: 1, PurchaseRequestID: 7, IsActive: true,
		EstimatedQuantity: 2, EstimatedCostPerItem: 1000,
		ActualQuantity: i8(2), ActualCostPerItem: i8(1100),
	})
	router := newTestServer(repo)

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

// Every response publishes what the item will accept next, so a client does
// not have to encode the lifecycle itself.
func TestResponsePublishesAllowedTransitions(t *testing.T) {
	repo := newFakeRepo()
	seedItem(repo, db.LineItemStatusPrPending)
	router := newTestServer(repo)

	rec := do(t, router, http.MethodGet, "/line-items/1", "")
	assertStatus(t, rec, http.StatusOK)

	item := decodeItem(t, rec)
	want := []string{
		string(db.LineItemStatusPrApproved),
		string(db.LineItemStatusPrRejected),
	}
	if len(item.AllowedTransitions) != len(want) {
		t.Fatalf("allowed = %v, want %v", item.AllowedTransitions, want)
	}
	for i, status := range want {
		if item.AllowedTransitions[i] != status {
			t.Errorf("allowed[%d] = %q, want %q",
				i, item.AllowedTransitions[i], status)
		}
	}
}

func TestUpdateEstimatesEndpointConflictsOutsideDraft(t *testing.T) {
	repo := newFakeRepo()
	seedItem(repo, db.LineItemStatusPrPending)
	router := newTestServer(repo)

	rec := do(t, router, http.MethodPatch, "/line-items/1/estimates",
		`{"estimated_quantity": 5, "estimated_cost_per_item": 200}`)
	assertStatus(t, rec, http.StatusConflict)
}

func TestUpdateActualsEndpoint(t *testing.T) {
	repo := newFakeRepo()
	seedItem(repo, db.LineItemStatusReimbDraft)
	router := newTestServer(repo)

	rec := do(t, router, http.MethodPatch, "/line-items/1/actuals",
		`{"actual_quantity": 3, "actual_cost_per_item": 900}`)
	assertStatus(t, rec, http.StatusOK)

	item := decodeItem(t, rec)
	if item.ActualQuantity == nil || *item.ActualQuantity != 3 {
		t.Errorf("actual quantity = %v, want 3", item.ActualQuantity)
	}
	if item.ActualUnitCost != nil {
		t.Errorf("actual unit cost = %v, want it left unset", *item.ActualUnitCost)
	}
}

func TestTransitionEndpoint(t *testing.T) {
	tests := map[string]struct {
		status db.LineItemStatus
		body   string
		want   int
	}{
		"legal edge": {
			status: db.LineItemStatusPrDraft,
			body:   `{"status": "pr_pending"}`,
			want:   http.StatusOK,
		},
		"illegal edge": {
			status: db.LineItemStatusPrDraft,
			body:   `{"status": "reimb_approved"}`,
			want:   http.StatusConflict,
		},
		"unknown status": {
			status: db.LineItemStatusPrDraft,
			body:   `{"status": "nonsense"}`,
			want:   http.StatusConflict,
		},
		"reimbursement without actuals": {
			status: db.LineItemStatusReimbDraft,
			body:   `{"status": "reimb_pending"}`,
			want:   http.StatusBadRequest,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			repo := newFakeRepo()
			seedItem(repo, tc.status)
			router := newTestServer(repo)

			rec := do(t, router, http.MethodPost, "/line-items/1/transition",
				tc.body)
			assertStatus(t, rec, tc.want)
		})
	}
}

func TestAttachReimbursementEndpoint(t *testing.T) {
	repo := newFakeRepo()
	seedItem(repo, db.LineItemStatusPrApproved)
	router := newTestServer(repo)

	rec := do(t, router, http.MethodPost, "/line-items/1/reimbursement",
		`{"reimbursement_request_id": 42}`)
	assertStatus(t, rec, http.StatusOK)

	item := decodeItem(t, rec)
	if item.Status != string(db.LineItemStatusReimbDraft) {
		t.Errorf("status = %q, want reimb_draft", item.Status)
	}
	if item.ReimbursementRequestID == nil || *item.ReimbursementRequestID != 42 {
		t.Errorf("reimbursement id = %v, want 42", item.ReimbursementRequestID)
	}
}

func TestArchiveEndpointReturnsNoContent(t *testing.T) {
	repo := newFakeRepo()
	seedItem(repo, db.LineItemStatusPrDraft)
	router := newTestServer(repo)

	rec := do(t, router, http.MethodDelete, "/line-items/1", "")
	assertStatus(t, rec, http.StatusNoContent)

	if repo.items[1].IsActive {
		t.Error("IsActive = true, want false")
	}
}

func TestListByReimbursementEndpoint(t *testing.T) {
	repo := newFakeRepo()
	repo.seed(db.LineItem{
		ID: 1, IsActive: true, ReimbursementRequestID: i8(42),
	})
	router := newTestServer(repo)

	rec := do(t, router, http.MethodGet,
		"/reimbursement-requests/42/line-items", "")
	assertStatus(t, rec, http.StatusOK)

	var items []lineItemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil {
		t.Fatalf("decoding %s: %v", rec.Body.String(), err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Errorf("got %d items, want line item 1", len(items))
	}
}

// A repository failure must surface as a 500 without leaking its detail.
func TestUnexpectedErrorBecomesInternalServerError(t *testing.T) {
	repo := newFakeRepo()
	repo.err = io.ErrUnexpectedEOF
	router := newTestServer(repo)

	rec := do(t, router, http.MethodGet, "/line-items/1", "")
	assertStatus(t, rec, http.StatusInternalServerError)

	if strings.Contains(rec.Body.String(), io.ErrUnexpectedEOF.Error()) {
		t.Errorf("body %s leaked the underlying error", rec.Body)
	}
}
