package domain

import "time"

type Post struct {
	ID        string
	Driver    string
	Text      string
	CreatedAt time.Time
}
