package doto

import "time"

type Event struct {
	ChatID      string
	Timestamp   time.Time
	Title       string
	Description string
}
