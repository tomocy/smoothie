package domain

import (
	"sort"
	"time"
)

type Posts []*Post

func (ps *Posts) SortByNewest() {
	sort.Slice(*ps, func(i, j int) bool {
		pI, pJ := (*ps)[i], (*ps)[j]
		if pI.CreatedAt.Equal(pJ.CreatedAt) {
			return pI.Driver < pJ.Driver
		}

		return pI.CreatedAt.After(pJ.CreatedAt)
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
