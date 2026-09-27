// event router

package router

import (
	"backend/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, eventhandler *handler.EventHandler) {
	events := router.Group("/events")

	{
		events.POST("", eventhandler.CreateEvent)
		// add more GET/PUT...
	}

}
