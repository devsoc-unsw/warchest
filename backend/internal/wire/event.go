// wire up event's repository, service, handler and register its routes

package wire

import (
	"backend/db"
	"backend/internal/handler"
	"backend/internal/repository"
	"backend/internal/router"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
)

func WireEvent(queries *db.Queries, r *gin.Engine) {
	eventRepo := repository.NewEventRepository(queries)
	eventService := service.NewEventService(eventRepo)
	eventHandler := handler.NewEventHandler(eventService)
	router.RegisterRoutes(r, eventHandler)
}
