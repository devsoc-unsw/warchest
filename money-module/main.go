package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	userName := os.Getenv("USER_NAME")
	dbName := os.Getenv("DB_NAME")
	dbPassword := os.Getenv("DB_PASSWORD")

	log.Printf("Connecting as %s to database %s\n", userName, dbName)

	connString := "postgres://" + userName + ":" + dbPassword + "@localhost:5432/" + dbName
	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	defer dbpool.Close()

	carRepo := car.NewPgxRepository(dbpool)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
}
