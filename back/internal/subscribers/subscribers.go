package subscribers

import (
	"context"
	"database/sql"
	"errors"
)

// ErrAlreadySubscribed is returned when the email is already registered.
var ErrAlreadySubscribed = errors.New("email is already subscribed")

// SubscriberStore manages email subscribers in the database.
type SubscriberStore struct {
	DB *sql.DB
}

// NewSubscriberStore creates a SubscriberStore using the injected DB connection.
func NewSubscriberStore(db *sql.DB) *SubscriberStore {
	return &SubscriberStore{DB: db}
}

// AddSubscriber inserts a new subscriber email. Returns ErrAlreadySubscribed if duplicate.
func (s *SubscriberStore) AddSubscriber(ctx context.Context, email string) error {
	query := `
		INSERT INTO subscribers (email)
		VALUES ($1)
		ON CONFLICT (email) DO NOTHING
	`

	result, err := s.DB.ExecContext(ctx, query, email)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAlreadySubscribed
	}

	return nil
}

// GetAllSubscribers returns all subscriber email addresses.
func (s *SubscriberStore) GetAllSubscribers(ctx context.Context) ([]string, error) {
	query := `SELECT email FROM subscribers ORDER BY subscribed_at`

	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return emails, nil
}
