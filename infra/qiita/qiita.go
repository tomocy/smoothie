package qiita

import (
	"fmt"
	"time"

	"github.com/tomocy/smoothie/domain"
)

type Items []*Item

func (is Items) Adapt() domain.Posts {
	adapteds := make(domain.Posts, len(is))
	for i, item := range is {
		adapteds[i] = item.Adapt()
	}

	return adapteds
}

type Item struct {
	ID        string    `json:"id"`
	User      *User     `json:"user"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

func (i *Item) Adapt() *domain.Post {
	return &domain.Post{
		ID:        i.ID,
		Driver:    "qiita",
		User:      i.User.Adapt(),
		Text:      fmt.Sprintf("%s\n\n%s", i.Title, i.Body),
		CreatedAt: i.CreatedAt,
	}
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
