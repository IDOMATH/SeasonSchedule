package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"seasonschedule/internal/mailer"
	"seasonschedule/internal/store"
	"seasonschedule/internal/subscribers"
)

var quit chan struct{}

func Start(subStore *subscribers.SubscriberStore, eventStore *store.EventStore, m *mailer.Mailer) {
	quit = make(chan struct{})
	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		log.Printf("Scheduler: failed to load timezone, defaulting to UTC: %v", err)
		loc = time.UTC
	}

	go func() {
		log.Println("Scheduler: weekly digest scheduler started (Sundays 8 PM CT)")
		for {
			next := nextSunday8PM(loc)
			log.Printf("Scheduler: next digest scheduled for %s", next.Format(time.RFC1123))

			select {
			case <-quit:
				log.Println("Scheduler: stopped")
				return
			case <-time.After(time.Until(next)):
				log.Println("Scheduler: firing weekly digest")
				if err := runDigest(subStore, eventStore, m); err != nil {
					log.Printf("Scheduler: digest error: %v", err)
				}
				time.Sleep(2 * time.Minute)
			}
		}
	}()
}

func Stop() {
	if quit != nil {
		close(quit)
	}
}

func runDigest(subStore *subscribers.SubscriberStore, eventStore *store.EventStore, m *mailer.Mailer) error {
	emails, err := subStore.GetAllSubscribers(context.Background())
	if err != nil {
		return err
	}

	now := time.Now()
	sunday := now.AddDate(0, 0, -int(now.Weekday()))
	sundayISO := sunday.Format("2006-01-02")

	events, err := eventStore.GetEventsForWeek(context.Background(), sundayISO, uuid.Nil)
	if err != nil {
		return err
	}
	return m.SendWeeklyDigest(emails, events)
}

func nextSunday8PM(loc *time.Location) time.Time {
	now := time.Now().In(loc)

	daysUntilSunday := (7 - int(now.Weekday())) % 7

	candidate := time.Date(now.Year(), now.Month(), now.Day()+daysUntilSunday, 20, 0, 0, 0, loc)

	if candidate.Before(now) || candidate.Equal(now) {
		candidate = candidate.AddDate(0, 0, 7)
	}

	return candidate
}