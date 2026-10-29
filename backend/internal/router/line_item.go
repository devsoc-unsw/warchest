// Line item routes.

package router

import (
	"backend/internal/handler"

	"github.com/gin-gonic/gin"
)


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
