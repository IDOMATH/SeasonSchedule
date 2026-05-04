package router

import (
	"database/sql"
	"net/http"

	"seasonschedule/internal/handlers"
	"seasonschedule/internal/middleware"
	"seasonschedule/internal/store"
	"seasonschedule/internal/subscribers"
)

// NewRouter creates and returns a configured router handler
func NewRouter(db *sql.DB, eventStore *store.EventStore, subStore *subscribers.SubscriberStore) http.Handler {
	mux := http.NewServeMux()

	eventHandler := &handlers.EventHandler{Store: eventStore}
	subscriberHandler := &handlers.SubscriberHandler{Store: subStore}
	healthHandler := &handlers.HealthHandler{DB: db}

	mux.HandleFunc("GET /health", healthHandler.HandleHealth)
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		http.ServeFile(w, r, "api/openapi.yaml")
	})
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/openapi.yaml", http.StatusMovedPermanently)
	})

	// Define routes
	mux.HandleFunc("POST /api/login", middleware.CORSMiddleware(eventHandler.HandleLogin))
	mux.HandleFunc("GET /api/events", middleware.CORSMiddleware(eventHandler.HandleGetEvents))
	mux.HandleFunc("GET /api/teams", middleware.CORSMiddleware(eventHandler.HandleGetTeams)) // formerly schedules
	mux.HandleFunc("GET /api/organizations", middleware.CORSMiddleware(eventHandler.HandleGetOrganizations))
	mux.HandleFunc("GET /api/organizations/{orgId}/teams", middleware.CORSMiddleware(eventHandler.HandleGetTeamsByOrg))
	mux.HandleFunc("GET /api/organizations/{orgId}/events", middleware.CORSMiddleware(eventHandler.HandleGetEventsByOrg))
	
	// Permissions endpoints
	mux.HandleFunc("GET /api/users/{userId}/permissions", middleware.CORSMiddleware(middleware.AuthMiddleware(eventHandler.HandleGetUserPermissions)))
	mux.HandleFunc("POST /api/permissions", middleware.CORSMiddleware(middleware.AuthMiddleware(eventHandler.HandleGrantPermission)))
	mux.HandleFunc("DELETE /api/permissions", middleware.CORSMiddleware(middleware.AuthMiddleware(eventHandler.HandleRevokePermission)))
	
	// Event mutations
	mux.HandleFunc("POST /api/events", middleware.CORSMiddleware(middleware.AuthMiddleware(eventHandler.HandleCreateEvent)))
	mux.HandleFunc("PUT /api/events/{id}", middleware.CORSMiddleware(middleware.AuthMiddleware(eventHandler.HandleUpdateEvent)))
	mux.HandleFunc("DELETE /api/events/{id}", middleware.CORSMiddleware(middleware.AuthMiddleware(eventHandler.HandleDeleteEvent)))
	
	// Subscriptions
	mux.HandleFunc("POST /api/subscribe", middleware.CORSMiddleware(subscriberHandler.HandleSubscribe))

	// Handle OPTIONS for CORS
	mux.HandleFunc("OPTIONS /", middleware.CORSMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	return middleware.RequestLoggerMiddleware(middleware.RecoverMiddleware(mux))
}
