package tumblr

import "time"

type Post struct {
	ID      string    `json:"id"`
	Summary string    `json:"summary"`
	Tags    []string  `json:"tags"`
	Date    time.Time `json:"date"`
}
