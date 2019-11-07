package qiita

import "time"

type Item struct {
	ID        string    `json:"id"`
	User      *User     `json:"user"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
