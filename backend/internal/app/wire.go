// Package app assembles the layers into a running application.
//
// It is the only place that knows about all three layers at once. Every layer
// depends downwards or not at all, so nothing else has to name a concrete
// implementation. The compiler settles here that the Postgres repository still
// satisfies the interface the service asks for, and that the handler still
// accepts the service it is handed.
package app

import (
	"backend/db"
	lineitemhandler "backend/internal/handler/lineitem"
	lineitemrepo "backend/internal/repository/lineitem"
	lineitemsvc "backend/internal/service/lineitem"

	"github.com/gin-gonic/gin"
)

// Wire mounts every module's routes on the given router, so main hands over
// the queries and a router and nothing more. Adding a feature is one call here.
func Wire(queries *db.Queries, router gin.IRouter) {
	wireLineItems(queries, router)
}

func wireLineItems(queries *db.Queries, router gin.IRouter) {
	repo := lineitemrepo.New(queries)
	svc := lineitemsvc.NewService(repo)
	lineitemhandler.RegisterRoutes(router, lineitemhandler.NewHandler(svc))
}
