package tumblr

import (
	"fmt"
	"strings"
	"time"

	"github.com/tomocy/smoothie/domain"
)

type Posts struct {
	Resp struct {
		Posts []*Post `json:"posts"`
	} `json:"response"`
}

type Post struct {
	ID       string    `json:"id"`
	BlogName string    `json:"blog_name"`
	Summary  string    `json:"summary"`
	Tags     []string  `json:"tags"`
	Date     time.Time `json:"date"`
}

func (p *Post) Adapt() *domain.Post {
	return &domain.Post{
		ID: p.ID,
		User: &domain.User{
			Name: p.BlogName,
		},
		Text:      p.joinText(),
		CreatedAt: p.Date,
	}
}

func (p *Post) joinText() string {
	return fmt.Sprintf("%s\n%s", p.Summary, strings.Join(p.Tags, " "))
}
