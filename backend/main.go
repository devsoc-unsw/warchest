package main

import (
	// "backend/db"
	// "backend/internal/wire"
	// "context"
	// "log"

	// "github.com/gin-gonic/gin"
	// "github.com/jackc/pgx/v5/pgxpool"
	// swaggerFiles "github.com/swaggo/files"
	// ginSwagger "github.com/swaggo/gin-swagger"
	"log/slog"
	"os"
)

func FinalHandler(opts) {

}

func initLogger() {
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func main() {
	// database conn
	connStr := "postgres://postgres:password@localhost:5432/warchest?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
		slog.Error("failed to connect db: %v", "err msg", err)
	}
	defer pool.Close()
	// sqlc queries
	queries := db.New(pool)
	// Gin HTTP router
	router := gin.Default()

	// each module wires it own
	wire.WireEvent(queries, router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	if err := router.Run(":8080"); err != nil {
		panic(err)
	}	
	initLogger()
	slog.Info("meow", "id", 100)
}

