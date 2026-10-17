// Wire up the line item repository, service and handler, and register its
// routes.

package wire

import (
	"backend/db"
	"backend/internal/handler"
	"backend/internal/repository"
	"backend/internal/router"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

func WireLineItem(queries *db.Queries, r *gin.Engine) {
	lineItemRepo := repository.NewLineItemRepository(queries)
	lineItemService := service.NewLineItemService(lineItemRepo)
	lineItemHandler := handler.NewLineItemHandler(lineItemService)
	router.RegisterLineItemRoutes(r, lineItemHandler)
}
