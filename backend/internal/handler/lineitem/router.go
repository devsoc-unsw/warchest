package lineitem

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the line item endpoints.
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
func RegisterRoutes(router gin.IRouter, handler *Handler) {
	purchaseRequests := router.Group("/purchase-requests/:prID/line-items")
	{
		purchaseRequests.POST("", handler.Create)
		purchaseRequests.GET("", handler.ListByPurchaseRequest)
		purchaseRequests.GET("/totals", handler.Totals)
	}

	router.GET(
		"/reimbursement-requests/:reimbID/line-items",
		handler.ListByReimbursementRequest,
	)

	lineItems := router.Group("/line-items")
	{
		lineItems.GET("/:id", handler.GetByID)
		lineItems.PATCH("/:id/estimates", handler.UpdateEstimates)
		lineItems.PATCH("/:id/actuals", handler.UpdateActuals)
		lineItems.POST("/:id/reimbursement", handler.AttachReimbursementRequest)
		lineItems.POST("/:id/transition", handler.Transition)
		lineItems.DELETE("/:id", handler.Archive)
	}
}
