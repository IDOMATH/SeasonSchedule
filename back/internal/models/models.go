package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Organization represents a top-level organization
type Organization struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// Team represents a team within an organization. A team effectively has one schedule.
type Team struct {
	ID               uuid.UUID `json:"id" db:"id"`
	OrganizationID   uuid.UUID `json:"organizationId" db:"organization_id"`
	Name             string    `json:"name" db:"name"`
	OrganizationName string    `json:"organizationName,omitempty"`
	CreatedAt        time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time `json:"updatedAt" db:"updated_at"`
}

// EventItem represents an event in the schedule for a team
type EventItem struct {
	ID            uuid.UUID `json:"id" db:"id"`
	Title         string    `json:"title" db:"title"`
	StartDateTime time.Time `json:"startDateTime" db:"start_date_time"`
	EndDateTime   time.Time `json:"endDateTime,omitempty" db:"end_date_time"`
	Location      string    `json:"location,omitempty" db:"location"`
	TeamID        uuid.UUID `json:"teamId" db:"team_id"`
	TeamName      string    `json:"teamName,omitempty"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}

// User represents an authenticated application user
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email,omitempty" db:"email"`
	IsAdmin   bool      `json:"isAdmin" db:"is_admin"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents login response with token
type LoginResponse struct {
	Token string `json:"token"`
}

// ErrorResponse for returning error responses
type ErrorResponse struct {
	Error string `json:"error"`
}

// Claims represents JWT claims
type Claims struct {
	UserID   uuid.UUID `json:"userId"`
	IsAdmin  bool      `json:"isAdmin"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

// Permission represents a user's permission on a resource
type Permission struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"userId" db:"user_id"`
	ResourceType string    `json:"resourceType" db:"resource_type"`
	ResourceID   uuid.UUID `json:"resourceId" db:"resource_id"`
	Permission   string    `json:"permission" db:"permission"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}
