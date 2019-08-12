package domain

import (
	"sort"
	"time"
)

type User struct {
	ID       string
	Name     string
	Username string
}

type Posts []*Post

func (ps *Posts) SortByNewest() {
	sort.Slice(*ps, func(i, j int) bool {
		return (*ps)[i].CreatedAt.After((*ps)[j].CreatedAt)
	})
}

type Post struct {
	ID        string
	Driver    string
	Text      string
	CreatedAt time.Time
}
