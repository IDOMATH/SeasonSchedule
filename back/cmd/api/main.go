package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"seasonschedule/internal/database"
	"seasonschedule/internal/mailer"
	"seasonschedule/internal/router"
	"seasonschedule/internal/scheduler"
	"seasonschedule/internal/store"
	"seasonschedule/internal/subscribers"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(log.Writer(), &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	config := database.LoadConfig()

	db, err := database.Connect(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	log.Println("Database connection established successfully")

	eventStore := store.NewEventStore(db)
	subStore := subscribers.NewSubscriberStore(db)

	m := mailer.New(mailer.LoadConfig())
	scheduler.Start(subStore, eventStore, m)

	appRouter := router.NewRouter(db, eventStore, subStore)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: appRouter,
	}

	go func() {
		log.Printf("Server starting on http://localhost:8080")
		log.Printf("API endpoints:")
		log.Printf("  GET    /health (health check)")
		log.Printf("  GET    /openapi.yaml (API documentation)")
		log.Printf("  GET    /docs (redirect to openapi.yaml)")
		log.Printf("  POST   /api/login")
		log.Printf("  GET    /api/events")
		log.Printf("  GET    /api/events?week=YYYY-MM-DD")
		log.Printf("  POST   /api/events (requires auth)")
		log.Printf("  PUT    /api/events/{id} (requires auth)")
		log.Printf("  DELETE /api/events/{id} (requires auth)")
		log.Printf("  POST   /api/subscribe")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("Received signal %v, initiating graceful shutdown...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	scheduler.Stop()
	log.Println("Server shutdown complete")
}