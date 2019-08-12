package domain

import "time"

type User struct {
	Drivers []string
}

type Post struct {
	ID        string
	Driver    string
	Text      string
	CreatedAt time.Time
}
