package subscribers

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

var emailRegex = regexp.MustCompile(`^[^@]+@[^@]+\.[^@]+$`)

func TestAddSubscriber_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub db: %v", err)
	}
	defer db.Close()

	s := NewSubscriberStore(db)

	mock.ExpectExec(`INSERT INTO subscribers`).
		WithArgs("test@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := s.AddSubscriber(context.Background(), "test@example.com"); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestAddSubscriber_Duplicate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub db: %v", err)
	}
	defer db.Close()

	s := NewSubscriberStore(db)

	// ON CONFLICT DO NOTHING → 0 rows affected
	mock.ExpectExec(`INSERT INTO subscribers`).
		WithArgs("dup@example.com").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = s.AddSubscriber(context.Background(), "dup@example.com")
	if err != ErrAlreadySubscribed {
		t.Errorf("expected ErrAlreadySubscribed, got %v", err)
	}
}

func TestGetAllSubscribers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub db: %v", err)
	}
	defer db.Close()

	s := NewSubscriberStore(db)

	rows := sqlmock.NewRows([]string{"email"}).
		AddRow("alice@example.com").
		AddRow("bob@example.com")

	mock.ExpectQuery(`SELECT email FROM subscribers`).WillReturnRows(rows)

	emails, err := s.GetAllSubscribers(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(emails) != 2 {
		t.Errorf("expected 2 emails, got %d", len(emails))
	}
	if emails[0] != "alice@example.com" || emails[1] != "bob@example.com" {
		t.Errorf("unexpected email values: %v", emails)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
