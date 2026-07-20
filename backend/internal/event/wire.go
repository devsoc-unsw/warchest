// wire up this module's repository, service, handler and register its routes
// 

package event

import(
	"backend/db"
	"github.com/gin-gonic/gin"
)

func Wire(queries *db.Queries, router *gin.Engine) {
	repo := NewEventRepository(queries)
	service := NewEventService(repo)
	handler := NewEventHandler(service)
	RegisterRoutes(router, handler)
}