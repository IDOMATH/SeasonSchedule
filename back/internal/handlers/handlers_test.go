package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"seasonschedule/internal/models"
	"seasonschedule/internal/store"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestHandleGetEvents(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := store.NewEventStore(db)

	now := time.Now()
	testEventUUID := uuid.New()
	testTeamUUID := uuid.New()
	
	rows := sqlmock.NewRows([]string{"id", "title", "start_date_time", "end_date_time", "location", "team_id", "created_at", "updated_at", "team_name"}).
		AddRow(testEventUUID, "Test Event", now, now.Add(time.Hour), "Test Location", testTeamUUID, now, now, "Test Team")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT e.id, e.title, e.start_date_time, e.end_date_time, e.location, e.team_id, e.created_at, e.updated_at, t.name as team_name FROM events e JOIN teams t ON e.team_id = t.id ORDER BY e.start_date_time`)).
		WillReturnRows(rows)

	req, err := http.NewRequest("GET", "/api/events", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := &EventHandler{Store: s}

	handler.HandleGetEvents(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var events []models.EventItem
	err = json.Unmarshal(rr.Body.Bytes(), &events)
	if err != nil {
		t.Fatalf("Could not parse response body: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].Title != "Test Event" {
		t.Errorf("Expected 'Test Event' but got '%s'", events[0].Title)
	}
}

func TestHandleLogin_InvalidMethod(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/login", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := &EventHandler{}

	handler.HandleLogin(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}
}

func TestHandleLogin_InvalidJSON(t *testing.T) {
	req, err := http.NewRequest("POST", "/api/login", bytes.NewBuffer([]byte("{invalid json}")))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := &EventHandler{}

	handler.HandleLogin(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}
