package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/db"
	"backend/internal/app"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("warchest backend: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	pool, err := newPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	router := gin.Default()
	router.GET("/health", health(pool))

	// app is the only place that assembles the layers; main just supplies
	// the dependencies they need.
	app.Wire(db.New(pool), router)

	return serve(router)
}

// newPool connects to Postgres and proves the connection works before the
// server starts accepting traffic, so a bad configuration fails loudly at
// startup instead of as a 500 on the first request.
//
// The connection string is required rather than defaulted: a fallback
// containing credentials is the kind of thing that quietly ends up pointing at
// the wrong database. docker-compose supplies BE_DATABASE_URL.
func newPool(ctx context.Context) (*pgxpool.Pool, error) {
	url := os.Getenv("BE_DATABASE_URL")
	if url == "" {
		return nil, errors.New("BE_DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}
	return pool, nil
}

// serve runs the server until interrupted, then lets in-flight requests finish
// rather than cutting them off.
func serve(handler http.Handler) error {
	addr := ":" + port()
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	failed := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", addr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			failed <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-failed:
		return fmt.Errorf("serving: %w", err)
	case <-stop:
	}

	log.Println("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// health reports whether the process can still reach the database, so a
// container orchestrator can tell "running" apart from "working".
func health(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "database unavailable",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}
