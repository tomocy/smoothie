package qiita

import (
	"time"

	"github.com/tomocy/smoothie/domain"
)

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

func (u *User) Adapt() *domain.User {
	return &domain.User{
		Name:     u.Name,
		Username: u.ID,
	}
}
