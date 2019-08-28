package github

import (
	"fmt"
	"time"

	"github.com/tomocy/smoothie/domain"
)

type Events []*Event

func (es Events) Adapt() domain.Posts {
	adapteds := make(domain.Posts, len(es))
	for i, e := range es {
		adapteds[i] = e.Adapt()
	}

	return adapteds
}

type Event struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Actor *User  `json:"actor"`
	Repo  struct {
		Name string `json:"name"`
	} `json:"repo"`
	Payload struct {
		Action string `json:"action"`
	} `jsno:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

func (e *Event) Adapt() *domain.Post {
	return &domain.Post{
		ID:        fmt.Sprint(e.ID),
		Driver:    "github event",
		User:      e.Actor.Adapt(),
		Text:      e.joinText(),
		CreatedAt: e.CreatedAt,
	}
}

func (e *Event) joinText() string {
	if e.Type != "WatchEvent" {
		return ""
	}

	return fmt.Sprintf("%s %s", e.Payload.Action, e.Repo.Name)
}

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
		ID:        fmt.Sprint(i.ID),
		Driver:    "github",
		User:      i.User.Adapt(),
		Text:      fmt.Sprintf("%s\n%s", i.Title, i.Body),
		CreatedAt: i.CreatedAt,
	}
}

type User struct {
	Login string `json:"login"`
}

func (u *User) Adapt() *domain.User {
	return &domain.User{
		Username: u.Login,
	}
}
