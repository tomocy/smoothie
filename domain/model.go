package domain

import (
	"sort"
	"time"
)

type Posts []*Post

func (ps *Posts) SortByNewest() {
	sort.Slice(*ps, func(i, j int) bool {
		return (*ps)[i].CreatedAt.After((*ps)[j].CreatedAt)
	})
}

type Post struct {
	ID        string
	Driver    string
	User      *User
	Text      string
	CreatedAt time.Time
}

type User struct {
	ID       string
	Name     string
	Username string
}
