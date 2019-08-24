package github

import "time"

type Issue struct {
	ID        int       `json:"id"`
	User      *User     `json:"user"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	Login string `json:"login"`
}
