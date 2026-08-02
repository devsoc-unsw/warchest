package lineitem

import (
	"backend/db"

	"github.com/gin-gonic/gin"
)

// Wire assembles the module and mounts its routes, so main only has to hand
// over the queries and a router.
//
// The layers are built here rather than in main because nothing outside this
// package needs to know the repository or the service exist.
func Wire(queries *db.Queries, router gin.IRouter) {
	repo := NewRepository(queries)
	svc := NewService(repo)
	RegisterRoutes(router, NewHandler(svc))
}
