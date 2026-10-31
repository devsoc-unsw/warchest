// line item model

package domain

import "time"

type LineItemStatus string

const (
	LineItemStatusPrDraft       LineItemStatus = "pr_draft"
	LineItemStatusPrPending     LineItemStatus = "pr_pending"
	LineItemStatusPrApproved    LineItemStatus = "pr_approved"
	LineItemStatusPrRejected    LineItemStatus = "pr_rejected"
	LineItemStatusReimbDraft    LineItemStatus = "reimb_draft"
	LineItemStatusReimbPending  LineItemStatus = "reimb_pending"
	LineItemStatusReimbApproved LineItemStatus = "reimb_approved"
	LineItemStatusReimbRejected LineItemStatus = "reimb_rejected"
)

// a nil pointer is a value that has not been recorded yet, money is in cents
type LineItem struct {
	ID                     int64
	PurchaseRequestID      int64
	ReimbursementRequestID *int64
	Status                 LineItemStatus
	IsActive               bool

	EstimatedQuantity    int64
	EstimatedCostPerItem int64
	EstimatedUnitCost    int64
	PrDescription        *string

	ActualQuantity    *int64
	ActualCostPerItem *int64
	ActualUnitCost    *int64
	ReimbDescription  *string

	CreatedAt time.Time
	// the read version guarded writes are checked against
	UpdatedAt time.Time
}
