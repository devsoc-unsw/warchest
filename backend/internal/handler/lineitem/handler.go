// Package lineitem exposes the line item service over HTTP.
//
// It is the outermost of three layers and holds no rules: it parses input,
// calls one service method, and maps the resulting error to a status code. The
// rules it enforces on a client's behalf live in
// backend/internal/service/lineitem.
package lineitem

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"backend/db"
	service "backend/internal/service/lineitem"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// Handler adapts HTTP to the service. It holds no rules of its own: it parses
// input, calls one service method, and maps the resulting error to a status
// code.
type Handler struct {
	svc service.Service
}

// NewHandler returns a handler backed by the given service.
func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

// lineItemResponse is the wire shape of a line item.
//
// It exists so pgtype values never reach the client: a pgtype.Int8 marshals as
// {"Int64":0,"Valid":false}, which would force every consumer to understand
// pgx. Absent values are null here instead.
//
// All money fields are in cents. The totals are derived rather than stored,
// so a client never has to decide how to combine quantity and cost itself.
type lineItemResponse struct {
	ID                     int64    `json:"id"`
	PurchaseRequestID      int64    `json:"purchase_request_id"`
	ReimbursementRequestID *int64   `json:"reimbursement_request_id"`
	Status                 string   `json:"status"`
	AllowedTransitions     []string `json:"allowed_transitions"`
	IsActive               bool     `json:"is_active"`

	EstimatedQuantity    int64   `json:"estimated_quantity"`
	EstimatedCostPerItem int64   `json:"estimated_cost_per_item"`
	EstimatedUnitCost    int64   `json:"estimated_unit_cost"`
	EstimatedTotal       int64   `json:"estimated_total"`
	PrDescription        *string `json:"pr_description"`

	ActualQuantity    *int64  `json:"actual_quantity"`
	ActualCostPerItem *int64  `json:"actual_cost_per_item"`
	ActualUnitCost    *int64  `json:"actual_unit_cost"`
	ActualTotal       *int64  `json:"actual_total"`
	ReimbDescription  *string `json:"reimb_description"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type totalsResponse struct {
	EstimatedTotal int64 `json:"estimated_total"`
	ActualTotal    int64 `json:"actual_total"`
}

type createRequest struct {
	EstimatedQuantity    int64  `json:"estimated_quantity"`
	EstimatedCostPerItem int64  `json:"estimated_cost_per_item"`
	EstimatedUnitCost    int64  `json:"estimated_unit_cost"`
	Description          string `json:"description"`
}

// estimatesRequest and actualsRequest mirror their service inputs field for
// field, differing only in carrying json tags, and are converted rather than
// copied across. Go permits the conversion because only the tags differ; if
// either side gains a field the conversion stops compiling, which is the
// intended warning rather than a silently dropped value.
type estimatesRequest struct {
	EstimatedQuantity    int64  `json:"estimated_quantity"`
	EstimatedCostPerItem int64  `json:"estimated_cost_per_item"`
	EstimatedUnitCost    int64  `json:"estimated_unit_cost"`
	Description          string `json:"description"`
}

// actualsRequest uses pointers throughout so an omitted field is
// distinguishable from a zero one: omitting a figure leaves it untouched,
// which is how a reimbursement is filled in as receipts arrive.
type actualsRequest struct {
	ActualQuantity    *int64  `json:"actual_quantity"`
	ActualCostPerItem *int64  `json:"actual_cost_per_item"`
	ActualUnitCost    *int64  `json:"actual_unit_cost"`
	Description       *string `json:"description"`
}

type transitionRequest struct {
	Status string `json:"status"`
}

type attachRequest struct {
	ReimbursementRequestID int64 `json:"reimbursement_request_id"`
}

// Create opens a line item on a purchase request.
func (h *Handler) Create(c *gin.Context) {
	prID, ok := h.pathID(c, "prID")
	if !ok {
		return
	}
	var req createRequest
	if !bindJSON(c, &req) {
		return
	}

	item, err := h.svc.Create(c.Request.Context(), service.CreateInput{
		PurchaseRequestID:    prID,
		EstimatedQuantity:    req.EstimatedQuantity,
		EstimatedCostPerItem: req.EstimatedCostPerItem,
		EstimatedUnitCost:    req.EstimatedUnitCost,
		Description:          req.Description,
	})
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, toResponse(item))
}

// GetByID returns one line item.
func (h *Handler) GetByID(c *gin.Context) {
	id, ok := h.pathID(c, "id")
	if !ok {
		return
	}
	item, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(item))
}

// ListByPurchaseRequest returns the active line items on a purchase request.
func (h *Handler) ListByPurchaseRequest(c *gin.Context) {
	prID, ok := h.pathID(c, "prID")
	if !ok {
		return
	}
	items, err := h.svc.ListByPurchaseRequest(c.Request.Context(), prID)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponses(items))
}

// ListByReimbursementRequest returns the line items on a reimbursement.
func (h *Handler) ListByReimbursementRequest(c *gin.Context) {
	reimbID, ok := h.pathID(c, "reimbID")
	if !ok {
		return
	}
	items, err := h.svc.ListByReimbursementRequest(c.Request.Context(), reimbID)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponses(items))
}

// Totals returns the summed estimated and actual cost of a purchase request.
func (h *Handler) Totals(c *gin.Context) {
	prID, ok := h.pathID(c, "prID")
	if !ok {
		return
	}
	totals, err := h.svc.TotalsByPurchaseRequest(c.Request.Context(), prID)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, totalsResponse(totals))
}

// UpdateEstimates revises what the requester expects to spend.
func (h *Handler) UpdateEstimates(c *gin.Context) {
	id, ok := h.pathID(c, "id")
	if !ok {
		return
	}
	var req estimatesRequest
	if !bindJSON(c, &req) {
		return
	}

	item, err := h.svc.UpdateEstimates(
		c.Request.Context(), id, service.EstimatesInput(req),
	)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(item))
}

// UpdateActuals records what was really spent.
func (h *Handler) UpdateActuals(c *gin.Context) {
	id, ok := h.pathID(c, "id")
	if !ok {
		return
	}
	var req actualsRequest
	if !bindJSON(c, &req) {
		return
	}

	item, err := h.svc.UpdateActuals(
		c.Request.Context(), id, service.ActualsInput(req),
	)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(item))
}

// AttachReimbursementRequest moves an approved line item into the
// reimbursement phase.
func (h *Handler) AttachReimbursementRequest(c *gin.Context) {
	id, ok := h.pathID(c, "id")
	if !ok {
		return
	}
	var req attachRequest
	if !bindJSON(c, &req) {
		return
	}

	item, err := h.svc.AttachReimbursementRequest(
		c.Request.Context(), id, req.ReimbursementRequestID,
	)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(item))
}

// Transition moves a line item along the lifecycle. The statuses a given item
// will accept are published as allowed_transitions on every response, so a
// client does not have to encode the lifecycle itself.
func (h *Handler) Transition(c *gin.Context) {
	id, ok := h.pathID(c, "id")
	if !ok {
		return
	}
	var req transitionRequest
	if !bindJSON(c, &req) {
		return
	}

	item, err := h.svc.Transition(
		c.Request.Context(), id, db.LineItemStatus(req.Status),
	)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(item))
}

// Archive soft deletes a line item.
func (h *Handler) Archive(c *gin.Context) {
	id, ok := h.pathID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Archive(c.Request.Context(), id); err != nil {
		fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// pathID reads a positive integer path parameter, answering 400 itself if it
// is missing or malformed. The bool reports whether the caller should carry on.
func (h *Handler) pathID(c *gin.Context, name string) (int64, bool) {
	raw := c.Param(name)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": name + " must be a positive integer, got " + raw,
		})
		return 0, false
	}
	return id, true
}

// bindJSON decodes the request body, answering 400 itself on malformed input.
func bindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return false
	}
	return true
}

// fail maps a service error to a status code. It is the only place that
// translation happens, so the handlers stay free of policy.
//
// Anything unrecognised is a 500 and its detail is withheld from the client,
// but attached to the context so logging middleware can record it.
func fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidTransition),
		errors.Is(err, service.ErrImmutableField):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

func toResponse(item db.LineItem) lineItemResponse {
	next := service.NextStates(item.Status)
	allowed := make([]string, len(next))
	for i, status := range next {
		allowed[i] = string(status)
	}

	res := lineItemResponse{
		ID:                     item.ID,
		PurchaseRequestID:      item.PurchaseRequestID,
		ReimbursementRequestID: nullInt(item.ReimbursementRequestID),
		Status:                 string(item.Status),
		AllowedTransitions:     allowed,
		IsActive:               item.IsActive,

		EstimatedQuantity:    item.EstimatedQuantity,
		EstimatedCostPerItem: item.EstimatedCostPerItem,
		EstimatedUnitCost:    item.EstimatedUnitCost,
		EstimatedTotal:       item.EstimatedQuantity * item.EstimatedCostPerItem,
		PrDescription:        nullText(item.PrDescription),

		ActualQuantity:    nullInt(item.ActualQuantity),
		ActualCostPerItem: nullInt(item.ActualCostPerItem),
		ActualUnitCost:    nullInt(item.ActualUnitCost),
		ReimbDescription:  nullText(item.ReimbDescription),

		CreatedAt: item.CreatedAt.Time,
		UpdatedAt: item.UpdatedAt.Time,
	}
	if item.ActualQuantity.Valid && item.ActualCostPerItem.Valid {
		total := item.ActualQuantity.Int64 * item.ActualCostPerItem.Int64
		res.ActualTotal = &total
	}
	return res
}

// toResponses always returns a non-nil slice, so an empty result marshals as
// [] rather than null and clients can iterate it unconditionally.
func toResponses(items []db.LineItem) []lineItemResponse {
	out := make([]lineItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toResponse(item))
	}
	return out
}

func nullInt(v pgtype.Int8) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

func nullText(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
}
