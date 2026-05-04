package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"seasonschedule/internal/models"
)

// EventStore manages events and teams in the database
type EventStore struct {
	DB *sql.DB
}

// NewEventStore creates a new event store with database connection
func NewEventStore(db *sql.DB) *EventStore {
	return &EventStore{
		DB: db,
	}
}

// GetOrganizations returns all organizations
func (s *EventStore) GetOrganizations(ctx context.Context) ([]models.Organization, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM organizations
		ORDER BY name
	`

	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []models.Organization
	for rows.Next() {
		var org models.Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orgs, nil
}

// GetTeams returns all teams with their organization names
func (s *EventStore) GetTeams(ctx context.Context) ([]models.Team, error) {
	query := `
		SELECT t.id, t.organization_id, t.name, t.created_at, t.updated_at, o.name as organization_name
		FROM teams t
		JOIN organizations o ON t.organization_id = o.id
		ORDER BY o.name, t.name
	`

	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var team models.Team
		if err := rows.Scan(&team.ID, &team.OrganizationID, &team.Name, &team.CreatedAt, &team.UpdatedAt, &team.OrganizationName); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return teams, nil
}

// GetTeamsByOrganization returns all teams for a specific organization
func (s *EventStore) GetTeamsByOrganization(ctx context.Context, orgID uuid.UUID) ([]models.Team, error) {
	query := `
		SELECT t.id, t.organization_id, t.name, t.created_at, t.updated_at, o.name as organization_name
		FROM teams t
		JOIN organizations o ON t.organization_id = o.id
		WHERE t.organization_id = $1
		ORDER BY t.name
	`

	rows, err := s.DB.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var team models.Team
		if err := rows.Scan(&team.ID, &team.OrganizationID, &team.Name, &team.CreatedAt, &team.UpdatedAt, &team.OrganizationName); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return teams, nil
}

// GetEventByID returns a single event by ID
func (s *EventStore) GetEventByID(ctx context.Context, id uuid.UUID) (models.EventItem, error) {
	query := `
		SELECT e.id, e.title, e.start_date_time, e.end_date_time, e.location, e.team_id, e.created_at, e.updated_at, t.name as team_name
		FROM events e
		JOIN teams t ON e.team_id = t.id
		WHERE e.id = $1
	`

	var event models.EventItem
	if err := s.DB.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.Title,
		&event.StartDateTime,
		&event.EndDateTime,
		&event.Location,
		&event.TeamID,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.TeamName,
	); err != nil {
		return models.EventItem{}, err
	}

	return event, nil
}

// GetEventsForWeek returns all events for a given week (Sunday-based)
// week parameter should be ISO date (yyyy-mm-dd) for the Sunday of the week
func (s *EventStore) GetEventsForWeek(ctx context.Context, sundayDate string, teamID uuid.UUID) ([]models.EventItem, error) {
	sunday, err := time.Parse("2006-01-02", sundayDate)
	if err != nil {
		return nil, err
	}

	endOfWeek := sunday.AddDate(0, 0, 7)

	query := `
		SELECT e.id, e.title, e.start_date_time, e.end_date_time, e.location, e.team_id, e.created_at, e.updated_at, t.name as team_name
		FROM events e
		JOIN teams t ON e.team_id = t.id
		WHERE e.start_date_time >= $1 AND e.start_date_time < $2
	`

	args := []any{sunday, endOfWeek}
	
	// Check if we have a valid non-nil UUID for filter
	if teamID != uuid.Nil {
		query += " AND e.team_id = $3"
		args = append(args, teamID)
	}
	query += " ORDER BY e.start_date_time"

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.EventItem
	for rows.Next() {
		var event models.EventItem
		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.StartDateTime,
			&event.EndDateTime,
			&event.Location,
			&event.TeamID,
			&event.CreatedAt,
			&event.UpdatedAt,
			&event.TeamName,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// GetAllEvents returns all events from the database filtered by teamID if provided
func (s *EventStore) GetAllEvents(ctx context.Context, teamID uuid.UUID) ([]models.EventItem, error) {
	query := `
		SELECT e.id, e.title, e.start_date_time, e.end_date_time, e.location, e.team_id, e.created_at, e.updated_at, t.name as team_name
		FROM events e
		JOIN teams t ON e.team_id = t.id
	`

	args := []any{}
	if teamID != uuid.Nil {
		query += " WHERE e.team_id = $1 "
		args = append(args, teamID)
	}
	query += " ORDER BY e.start_date_time"

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.EventItem
	for rows.Next() {
		var event models.EventItem
		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.StartDateTime,
			&event.EndDateTime,
			&event.Location,
			&event.TeamID,
			&event.CreatedAt,
			&event.UpdatedAt,
			&event.TeamName,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// GetEventsByOrganization returns all events for teams within a specific organization
func (s *EventStore) GetEventsByOrganization(ctx context.Context, orgID uuid.UUID) ([]models.EventItem, error) {
	query := `
		SELECT e.id, e.title, e.start_date_time, e.end_date_time, e.location, e.team_id, e.created_at, e.updated_at, t.name as team_name
		FROM events e
		JOIN teams t ON e.team_id = t.id
		WHERE t.organization_id = $1
		ORDER BY e.start_date_time
	`

	rows, err := s.DB.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.EventItem
	for rows.Next() {
		var event models.EventItem
		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.StartDateTime,
			&event.EndDateTime,
			&event.Location,
			&event.TeamID,
			&event.CreatedAt,
			&event.UpdatedAt,
			&event.TeamName,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}


// CreateEvent inserts a new event into the database
func (s *EventStore) CreateEvent(ctx context.Context, event models.EventItem) (models.EventItem, error) {
	query := `
		INSERT INTO events (title, start_date_time, end_date_time, location, team_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	var id uuid.UUID
	var createdAt, updatedAt time.Time

	err := s.DB.QueryRowContext(ctx, query, event.Title, event.StartDateTime, event.EndDateTime, event.Location, event.TeamID, now, now).
		Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return models.EventItem{}, err
	}

	event.ID = id
	event.CreatedAt = createdAt
	event.UpdatedAt = updatedAt

	return event, nil
}

// UpdateEvent updates an existing event in the database
func (s *EventStore) UpdateEvent(ctx context.Context, id uuid.UUID, event models.EventItem) (models.EventItem, error) {
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)"
	err := s.DB.QueryRowContext(ctx, checkQuery, id).Scan(&exists)
	if err != nil {
		return models.EventItem{}, err
	}
	if !exists {
		return models.EventItem{}, sql.ErrNoRows
	}

	query := `
		UPDATE events
		SET title = $1, start_date_time = $2, end_date_time = $3, location = $4, team_id = $5, updated_at = $6
		WHERE id = $7
		RETURNING created_at, updated_at
	`

	now := time.Now()
	var createdAt, updatedAt time.Time

	err = s.DB.QueryRowContext(ctx, query, event.Title, event.StartDateTime, event.EndDateTime, event.Location, event.TeamID, now, id).
		Scan(&createdAt, &updatedAt)
	if err != nil {
		return models.EventItem{}, err
	}

	event.ID = id
	event.CreatedAt = createdAt
	event.UpdatedAt = updatedAt

	return event, nil
}

// DeleteEvent removes an event from the database
func (s *EventStore) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	result, err := s.DB.ExecContext(ctx, "DELETE FROM events WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
