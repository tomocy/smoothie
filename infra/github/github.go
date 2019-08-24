package github

import (
	"fmt"
	"time"

	"github.com/tomocy/smoothie/domain"
)

type Issues []*Issue

type Issue struct {
	ID        int       `json:"id"`
	User      *User     `json:"user"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

func (i *Issue) Adapt() *domain.Post {
	return &domain.Post{
		ID: fmt.Sprint(i.ID),
		User: &domain.User{
			Name: i.User.Login,
		},
		Text:      fmt.Sprintf("%s\n%s", i.Title, i.Body),
		CreatedAt: i.CreatedAt,
	}
}

type User struct {
	Login string `json:"login"`
}
