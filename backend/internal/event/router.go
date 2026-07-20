// router

package event

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, handler *EventHandler) {
	events := router.Group("/events") 

	{
		events.POST("", handler.CreateEvent)
		// add more GET/PUT...



}
	
}