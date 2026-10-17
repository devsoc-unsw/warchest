// Line item routes.

package router

import (
	"backend/internal/handler"

	"github.com/gin-gonic/gin"
)

// RegisterLineItemRoutes mounts the line item endpoints.
//
// Line items are nested under the request that owns them wherever the owner is
// what the caller has in hand, and addressed directly by id once it holds one.
// Estimates and actuals are separate endpoints rather than one PATCH, matching
// the service split: the two halves of a line item are editable at different
// points in the lifecycle, and a single endpoint would have to accept fields it
// then refuses.
//
// It takes a gin.IRouter rather than a *gin.Engine so the routes can be mounted
// under a versioned or authenticated group later without changing this file.
func RegisterLineItemRoutes(r gin.IRouter, h *handler.LineItemHandler) {
	purchaseRequests := r.Group("/purchase-requests/:prID/line-items")
	{
		purchaseRequests.POST("", h.Create)
		purchaseRequests.GET("", h.ListByPurchaseRequest)
		purchaseRequests.GET("/totals", h.Totals)
	}

	r.GET(
		"/reimbursement-requests/:reimbID/line-items",
		h.ListByReimbursementRequest,
	)

	lineItems := r.Group("/line-items")
	{
		lineItems.GET("/:id", h.GetByID)
		lineItems.PATCH("/:id/estimates", h.UpdateEstimates)
		lineItems.PATCH("/:id/actuals", h.UpdateActuals)
		lineItems.POST("/:id/reimbursement", h.AttachReimbursementRequest)
		lineItems.POST("/:id/transition", h.Transition)
		lineItems.DELETE("/:id", h.Archive)
	}
}
