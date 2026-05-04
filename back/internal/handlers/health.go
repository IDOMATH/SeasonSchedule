package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type HealthHandler struct {
	DB *sql.DB
}

type HealthResponse struct {
	Status   string    `json:"status"`
	Database string    `json:"database"`
	Time     time.Time `json:"time"`
}

func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.DB.PingContext(ctx); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(HealthResponse{
			Status:   "unhealthy",
			Database: "down",
			Time:     time.Now().UTC(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{
		Status:   "healthy",
		Database: "up",
		Time:     time.Now().UTC(),
	})
}
