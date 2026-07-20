package main

import (
	"context"
	"log"
	"backend/db"
	"backend/internal/event"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

)

func main() {
	// database conn
	connStr := "postgres://postgres:password@localhost:5432/warchest?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer pool.Close()
	queries := db.New(pool)

	router := gin.Default()

	// each module wires it own
	event.Wire(queries, router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.Run(":8080")

}
