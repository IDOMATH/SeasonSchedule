package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"seasonschedule/internal/auth"
	"seasonschedule/internal/middleware"
	"seasonschedule/internal/models"
	"seasonschedule/internal/store"
	"seasonschedule/internal/subscribers"
)

// EventHandler handles event-related HTTP requests
type EventHandler struct {
	Store *store.EventStore
}

// SubscriberHandler handles subscriber-related HTTP requests
type SubscriberHandler struct {
	Store *subscribers.SubscriberStore
}

// HandleLogin handles POST /api/login requests
func (h *EventHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginReq models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid request body"})
		return
	}

	user, ok := auth.ValidateCredentials(r.Context(), h.Store.DB, loginReq.Username, loginReq.Password)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(user)
	if err != nil {
		slog.Error("Failed to generate token", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.LoginResponse{Token: token})
}

// HandleGetEvents handles GET /api/events requests (can filter by teamId or week)
func (h *EventHandler) HandleGetEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	weekParam := r.URL.Query().Get("week")
	
	var teamID uuid.UUID
	if teamParam := r.URL.Query().Get("teamId"); teamParam != "" {
		if id, err := uuid.Parse(teamParam); err == nil {
			teamID = id
		}
	}

	var events []models.EventItem
	var err error

	if weekParam == "" {
		events, err = h.Store.GetAllEvents(r.Context(), teamID)
	} else {
		events, err = h.Store.GetEventsForWeek(r.Context(), weekParam, teamID)
	}

	if err != nil {
		slog.Error("Failed to get events", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve events"})
		return
	}

	if events == nil {
		events = []models.EventItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// HandleGetTeams handles GET /api/teams requests
func (h *EventHandler) HandleGetTeams(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teams, err := h.Store.GetTeams(r.Context())
	if err != nil {
		slog.Error("Failed to get teams", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve teams"})
		return
	}

	if teams == nil {
		teams = []models.Team{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

// HandleGetOrganizations handles GET /api/organizations requests
func (h *EventHandler) HandleGetOrganizations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orgs, err := h.Store.GetOrganizations(r.Context())
	if err != nil {
		slog.Error("Failed to get organizations", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve organizations"})
		return
	}

	if orgs == nil {
		orgs = []models.Organization{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgs)
}

// HandleGetTeamsByOrg handles GET /api/organizations/{orgId}/teams requests
func (h *EventHandler) HandleGetTeamsByOrg(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orgIDStr := r.PathValue("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid organization ID"})
		return
	}

	teams, err := h.Store.GetTeamsByOrganization(r.Context(), orgID)
	if err != nil {
		slog.Error("Failed to get teams", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve teams"})
		return
	}

	if teams == nil {
		teams = []models.Team{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

// HandleGetEventsByOrg handles GET /api/organizations/{orgId}/events requests
func (h *EventHandler) HandleGetEventsByOrg(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orgIDStr := r.PathValue("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid organization ID"})
		return
	}

	events, err := h.Store.GetEventsByOrganization(r.Context(), orgID)
	if err != nil {
		slog.Error("Failed to get events for org", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve events for organization"})
		return
	}

	if events == nil {
		events = []models.EventItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}


// HandleCreateEvent handles POST /api/events requests (requires auth)
func (h *EventHandler) HandleCreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event models.EventItem
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid request body"})
		return
	}

	if event.TeamID == uuid.Nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "teamId is required"})
		return
	}

	claims, ok := middleware.GetClaims(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Check if user can manage this team
	canManage, err := auth.CanManageTeam(r.Context(), h.Store.DB, claims.UserID, event.TeamID, claims.IsAdmin)
	if err != nil {
		slog.Error("Failed to check team permission", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to verify team permission"})
		return
	}
	if !canManage {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Forbidden - insufficient permissions on this team"})
		return
	}

	newEvent, err := h.Store.CreateEvent(r.Context(), event)
	if err != nil {
		slog.Error("Failed to create event", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to create event"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newEvent)
}

// HandleUpdateEvent handles PUT /api/events/{id} requests (requires auth)
func (h *EventHandler) HandleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid event ID"})
		return
	}

	var event models.EventItem
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid request body"})
		return
	}

	if event.TeamID == uuid.Nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "teamId is required"})
		return
	}

	existingEvent, err := h.Store.GetEventByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Event not found"})
			return
		}
		slog.Error("Failed to load event", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to update event"})
		return
	}

	claims, ok := middleware.GetClaims(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	if event.TeamID != existingEvent.TeamID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Cannot change team"})
		return
	}

	// Check if user can manage this team
	canManage, err := auth.CanManageTeam(r.Context(), h.Store.DB, claims.UserID, event.TeamID, claims.IsAdmin)
	if err != nil {
		slog.Error("Failed to check team permission", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to verify team permission"})
		return
	}
	if !canManage {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Forbidden - insufficient permissions on this team"})
		return
	}

	updatedEvent, err := h.Store.UpdateEvent(r.Context(), id, event)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Event not found"})
			return
		}
		slog.Error("Failed to update event", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to update event"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedEvent)
}

// HandleDeleteEvent handles DELETE /api/events/{id} requests (requires auth)
func (h *EventHandler) HandleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid event ID"})
		return
	}

	existingEvent, err := h.Store.GetEventByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Event not found"})
			return
		}
		slog.Error("Failed to load event", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to delete event"})
		return
	}

	claims, ok := middleware.GetClaims(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Check if user can manage this team
	canManage, err := auth.CanManageTeam(r.Context(), h.Store.DB, claims.UserID, existingEvent.TeamID, claims.IsAdmin)
	if err != nil {
		slog.Error("Failed to check team permission", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to verify team permission"})
		return
	}
	if !canManage {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Forbidden - insufficient permissions on this team"})
		return
	}

	err = h.Store.DeleteEvent(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Event not found"})
			return
		}
		slog.Error("Failed to delete event", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to delete event"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// GetLastSunday returns the last Sunday relative to the given date
func GetLastSunday(t time.Time) time.Time {
	days := int(t.Weekday())
	return t.AddDate(0, 0, -days)
}

// subscribeRequest is the expected payload for POST /api/subscribe.
type subscribeRequest struct {
	Email string `json:"email"`
}

// HandleGetUserPermissions handles GET /api/users/{userId}/permissions requests (admin only)
func (h *EventHandler) HandleGetUserPermissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.PathValue("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	// Only admins can view other users' permissions
	claims, ok := middleware.GetClaims(r)
	if !ok || !claims.IsAdmin {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Admin access required"})
		return
	}

	permissions, err := auth.GetUserPermissions(r.Context(), h.Store.DB, userID)
	if err != nil {
		slog.Error("Failed to get user permissions", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve user permissions"})
		return
	}

	if permissions == nil {
		permissions = []models.Permission{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permissions)
}

// HandleGrantPermission handles POST /api/permissions requests (admin only)
func (h *EventHandler) HandleGrantPermission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID       uuid.UUID `json:"userId"`
		ResourceType string    `json:"resourceType"`
		ResourceID   uuid.UUID `json:"resourceId"`
		Permission   string    `json:"permission"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Check if user can manage permissions for this resource
	claims, ok := middleware.GetClaims(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	canManagePerms, err := auth.CanManagePermissions(r.Context(), h.Store.DB, claims.UserID, req.ResourceType, req.ResourceID, claims.IsAdmin)
	if err != nil {
		slog.Error("Failed to check permission management rights", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to verify permission management rights"})
		return
	}
	if !canManagePerms {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Insufficient permissions to manage this resource"})
		return
	}

	// Validate resource type
	if req.ResourceType != "organization" && req.ResourceType != "team" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid resource type"})
		return
	}

	// Validate permission type
	if req.Permission != "read" && req.Permission != "write" && req.Permission != "admin" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid permission type"})
		return
	}

	err = auth.GrantPermission(r.Context(), h.Store.DB, req.UserID, req.ResourceType, req.ResourceID, req.Permission)
	if err != nil {
		slog.Error("Failed to grant permission", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to grant permission"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Permission granted successfully"})
}

// HandleRevokePermission handles DELETE /api/permissions requests (admin only)
func (h *EventHandler) HandleRevokePermission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID       uuid.UUID `json:"userId"`
		ResourceType string    `json:"resourceType"`
		ResourceID   uuid.UUID `json:"resourceId"`
		Permission   string    `json:"permission"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Check if user can manage permissions for this resource
	claims, ok := middleware.GetClaims(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Unauthorized"})
		return
	}

	canManagePerms, err := auth.CanManagePermissions(r.Context(), h.Store.DB, claims.UserID, req.ResourceType, req.ResourceID, claims.IsAdmin)
	if err != nil {
		slog.Error("Failed to check permission management rights", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to verify permission management rights"})
		return
	}
	if !canManagePerms {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Insufficient permissions to manage this resource"})
		return
	}

	err = auth.RevokePermission(r.Context(), h.Store.DB, req.UserID, req.ResourceType, req.ResourceID, req.Permission)
	if err != nil {
		slog.Error("Failed to revoke permission", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to revoke permission"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Permission revoked successfully"})
}

// HandleSubscribe handles POST /api/subscribe requests
func (h *SubscriberHandler) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req subscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "A valid email address is required"})
		return
	}

	err := h.Store.AddSubscriber(r.Context(), req.Email)
	if err == subscribers.ErrAlreadySubscribed {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "This email is already subscribed"})
		return
	}
	if err != nil {
		slog.Error("Failed to subscribe", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to subscribe, please try again"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully subscribed!"})
}
