package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"backend/db"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type LineItemHandler struct {
	service service.LineItemService
}

func NewLineItemHandler(service service.LineItemService) *LineItemHandler {
	return &LineItemHandler{service: service}
}

// from the db
type LineItemResponse struct {
	ID                     int64     `json:"id"`
	PurchaseRequestID      int64     `json:"purchase_request_id"`
	ReimbursementRequestID *int64    `json:"reimbursement_request_id"`
	Status                 string    `json:"status"`
	AllowedTransitions     []string  `json:"allowed_transitions"`
	IsActive               bool      `json:"is_active"`

	EstimatedQuantity      int64     `json:"estimated_quantity"`
	EstimatedCostPerItem   int64     `json:"estimated_cost_per_item"`
	EstimatedUnitCost      int64     `json:"estimated_unit_cost"`
	EstimatedTotal         int64     `json:"estimated_total"`
	PrDescription          *string   `json:"pr_description"`

	ActualQuantity         *int64    `json:"actual_quantity"`
	ActualCostPerItem      *int64    `json:"actual_cost_per_item"`
	ActualUnitCost         *int64    `json:"actual_unit_cost"`
	ActualTotal            *int64    `json:"actual_total"`
	ReimbDescription       *string   `json:"reimb_description"`

	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type LineItemTotalsResponse struct {
	EstimatedTotal int64 `json:"estimated_total"`
	ActualTotal    int64 `json:"actual_total"`
}

type createLineItemRequest struct {
	EstimatedQuantity    int64  `json:"estimated_quantity"`
	EstimatedCostPerItem int64  `json:"estimated_cost_per_item"`
	EstimatedUnitCost    int64  `json:"estimated_unit_cost"`
	Description          string `json:"description"`
}


type lineItemEstimatesRequest struct {
	EstimatedQuantity    int64  `json:"estimated_quantity"`
	EstimatedCostPerItem int64  `json:"estimated_cost_per_item"`
	EstimatedUnitCost    int64  `json:"estimated_unit_cost"`
	Description          string `json:"description"`
}

// can be null so pointer is used
type lineItemActualsRequest struct {
	ActualQuantity    *int64  `json:"actual_quantity"`
	ActualCostPerItem *int64  `json:"actual_cost_per_item"`
	ActualUnitCost    *int64  `json:"actual_unit_cost"`
	Description       *string `json:"description"`
}

type lineItemTransitionRequest struct {
	Status string `json:"status"`
}

type lineItemAttachRequest struct {
	ReimbursementRequestID int64 `json:"reimbursement_request_id"`
}

// Create opens a line item on a purchase request.
func (h *LineItemHandler) CreateLineItem(c *gin.Context) {
	prID, ok := h.pathID(c, "prID")
	if !ok {
		return
	}
	var req createLineItemRequest
	if !bindJSON(c, &req) {
		return
	}

	item, err := h.service.Create(c.Request.Context(), service.CreateLineItemInput{
		PurchaseRequestID:    prID,
		EstimatedQuantity:    req.EstimatedQuantity,
		EstimatedCostPerItem: req.EstimatedCostPerItem,
		EstimatedUnitCost:    req.EstimatedUnitCost,
		Description:          req.Description,
	})
	if err != nil {
		failLineItem(c, err)
		return
	}
	c.JSON(http.StatusCreated, toLineItemResponse(item))
}

// GetByID returns one line item.
func (h *LineItemHandler) GetByID(c *gin.Context) {
	id, ok := h.pathID(c, "id")
	if !ok {
		return
	}
	item, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		failLineItem(c, err)
		return
	}
	c.JSON(http.StatusOK, toLineItemResponse(item))
}

// ListByPurchaseRequest returns the active line items on a purchase request.
func (h *LineItemHandler) ListByPurchaseRequest(c *gin.Context) {
	prID, ok := h.pathID(c, "prID")
	if !ok {
		return
	}
	items, err := h.service.ListByPurchaseRequest(c.Request.Context(), prID)
	if err != nil {
		failLineItem(c, err)
		return
	}
	c.JSON(http.StatusOK, toLineItemResponses(items))
}

// ListByReimbursementRequest returns the line items on a reimbursement.
func (h *LineItemHandler) ListByReimbursementRequest(c *gin.Context) {
	reimbID, ok := h.pathID(c, "reimbID")
	if !ok {
		return
	}
	items, err := h.service.ListByReimbursementRequest(c.Request.Context(), reimbID)
	if err != nil {
		failLineItem(c, err)
		return
	}
	c.JSON(http.StatusOK, toLineItemResponses(items))
}

// Totals returns the summed estimated and actual cost of a purchase request.
func (h *LineItemHandler) Totals(c *gin.Context) {
	prID, ok := h.pathID(c, "prID")
	if !ok {
		return
	}
	totals, err := h.service.TotalsByPurchaseRequest(c.Request.Context(), prID)
	if err != nil {
		failLineItem(c, err)
		return
	}
	c.JSON(http.StatusOK, LineItemTotalsResponse(totals))
}

// UpdateEstimates revises what the requester expects to spend.
func (h *LineItemHandler) UpdateEstimates(c *gin.Context) {
	id, ok := h.pathID(c, "id")
	if !ok {
		return
	}
	var req lineItemEstimatesRequest
	if !bindJSON(c, &req) {
		return
	}

	item, err := h.service.UpdateEstimates(
		c.Request.Context(), id, service.LineItemEstimatesInput(req),
	)
	if err != nil {
		failLineItem(c, err)
		return
	}
	c.JSON(http.StatusOK, toLineItemResponse(item))
}

// UpdateActuals records what was really spent.
func (h *LineItemHandler) UpdateActuals(c *gin.Context) {
	id, ok := h.pathID(c, "id")
	if !ok {
		return
	}
	var req lineItemActualsRequest
	if !bindJSON(c, &req) {
		return
	}

	item, err := h.service.UpdateActuals(
		c.Request.Context(), id, service.LineItemActualsInput(req),
	)
	if err != nil {
		failLineItem(c, err)
		return
	}
	c.JSON(http.StatusOK, toLineItemResponse(item))
}

// AttachReimbursementRequest moves an approved line item into the
// reimbursement phase.
func (h *LineItemHandler) AttachReimbursementRequest(c *gin.Context) {
	id, ok := h.pathID(c, "id")
	if !ok {
		return
	}
	var req lineItemAttachRequest
	if !bindJSON(c, &req) {
		return
	}

	item, err := h.service.AttachReimbursementRequest(
		c.Request.Context(), id, req.ReimbursementRequestID,
	)
	if err != nil {
		failLineItem(c, err)
		return
	}
	c.JSON(http.StatusOK, toLineItemResponse(item))
}

// Transition moves a line item along the lifecycle. The statuses a given item
// will accept are published as allowed_transitions on every response, so a
// client does not have to encode the lifecycle itself.
func (h *LineItemHandler) Transition(c *gin.Context) {
	id, ok := h.pathID(c, "id")
	if !ok {
		return
	}
	var req lineItemTransitionRequest
	if !bindJSON(c, &req) {
		return
	}

	item, err := h.service.Transition(
		c.Request.Context(), id, db.LineItemStatus(req.Status),
	)
	if err != nil {
		failLineItem(c, err)
		return
	}
	c.JSON(http.StatusOK, toLineItemResponse(item))
}

// Archive soft deletes a line item.
func (h *LineItemHandler) Archive(c *gin.Context) {
	id, ok := h.pathID(c, "id")
	if !ok {
		return
	}
	if err := h.service.Archive(c.Request.Context(), id); err != nil {
		failLineItem(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// pathID reads a positive integer path parameter, answering 400 itself if it
// is missing or malformed. The bool reports whether the caller should carry on.
func (h *LineItemHandler) pathID(c *gin.Context, name string) (int64, bool) {
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

// failLineItem maps a service error to a status code. It is the only place that
// translation happens, so the handlers stay free of policy.
//
// Anything unrecognised is a 500 and its detail is withheld from the client,
// but attached to the context so logging middleware can record it.
func failLineItem(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrLineItemNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidLineItem):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidLineItemTransition),
		errors.Is(err, service.ErrImmutableLineItemField):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

func toLineItemResponse(item db.LineItem) LineItemResponse {
	next := service.NextLineItemStates(item.Status)
	allowed := make([]string, len(next))
	for i, status := range next {
		allowed[i] = string(status)
	}

	res := LineItemResponse{
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

// toLineItemResponses always returns a non-nil slice, so an empty result marshals as
// [] rather than null and clients can iterate it unconditionally.
func toLineItemResponses(items []db.LineItem) []LineItemResponse {
	out := make([]LineItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toLineItemResponse(item))
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
