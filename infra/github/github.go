package github

import (
	"fmt"
	"time"

	"github.com/tomocy/smoothie/domain"
)

type Issues []*Issue

func (is Issues) Adapt() domain.Posts {
	adapteds := make(domain.Posts, len(is))
	for j, i := range is {
		adapteds[j] = i.Adapt()
	}

	return adapteds
}

type Issue struct {
	ID        int       `json:"id"`
	User      *User     `json:"user"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

func (i *Issue) Adapt() *domain.Post {
	return &domain.Post{
		ID:     fmt.Sprint(i.ID),
		Driver: "github",
		User: &domain.User{
			Username: i.User.Login,
		},
		Text:      fmt.Sprintf("%s\n%s", i.Title, i.Body),
		CreatedAt: i.CreatedAt,
	}
}

type User struct {
	Login string `json:"login"`
}
