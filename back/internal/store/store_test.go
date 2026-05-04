package store

import (
	"context"
	"regexp"
	"testing"
	"time"

	"seasonschedule/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestGetAllEvents(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := NewEventStore(db)

	now := time.Now()
	testEventUUID := uuid.New()
	testTeamUUID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "title", "start_date_time", "end_date_time", "location", "team_id", "created_at", "updated_at", "team_name"}).
		AddRow(testEventUUID, "Test Event", now, now.Add(time.Hour), "Test Location", testTeamUUID, now, now, "Test Team")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT e.id, e.title, e.start_date_time, e.end_date_time, e.location, e.team_id, e.created_at, e.updated_at, t.name as team_name FROM events e JOIN teams t ON e.team_id = t.id ORDER BY e.start_date_time`)).
		WillReturnRows(rows)

	events, err := s.GetAllEvents(context.Background(), uuid.Nil)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].Title != "Test Event" {
		t.Errorf("Expected title 'Test Event', got '%s'", events[0].Title)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := NewEventStore(db)

	now := time.Now()
	testEventUUID := uuid.New()
	teamUUID := uuid.New()
	event := models.EventItem{
		Title:         "New Event",
		StartDateTime: now,
		EndDateTime:   now.Add(time.Hour),
		Location:      "New Location",
		TeamID:        teamUUID,
	}

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(testEventUUID, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO events (title, start_date_time, end_date_time, location, team_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`)).
		WithArgs(event.Title, event.StartDateTime, event.EndDateTime, event.Location, event.TeamID, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	createdEvent, err := s.CreateEvent(context.Background(), event)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if createdEvent.ID != testEventUUID {
		t.Errorf("Expected ID %s, got %s", testEventUUID, createdEvent.ID)
	}
	if createdEvent.Title != "New Event" {
		t.Errorf("Expected Title 'New Event', got '%s'", createdEvent.Title)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDeleteEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := NewEventStore(db)

	testEventUUID := uuid.New()
	
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM events WHERE id = $1`)).
		WithArgs(testEventUUID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = s.DeleteEvent(context.Background(), testEventUUID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
