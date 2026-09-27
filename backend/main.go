package main

import (
	"backend/db"
	"backend/internal/wire"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
)

func main() {
	// database conn
	connStr := "postgres://postgres:password@localhost:5432/warchest?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer pool.Close()
	// sqlc queries
	queries := db.New(pool)
	// Gin HTTP router
	router := gin.Default()

	// each module wires it own
	wire.WireEvent(queries, router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run(":8080")
}
