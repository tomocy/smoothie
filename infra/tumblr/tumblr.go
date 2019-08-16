package tumblr

import (
	"fmt"
	"strings"
	"time"

	"github.com/tomocy/smoothie/domain"
)

type Resp struct {
	Resp struct {
		Posts []*Post `json:"posts"`
	} `json:"response"`
}

type Posts []*Post

func (ps Posts) Adapt() domain.Posts {
	adapteds := make(domain.Posts, len(ps))
	for i, p := range ps {
		adapteds[i] = p.Adapt()
	}

	return adapteds
}

type Post struct {
	ID       int      `json:"id"`
	BlogName string   `json:"blog_name"`
	Summary  string   `json:"summary"`
	Tags     []string `json:"tags"`
	Date     date     `json:"date"`
}

func (p *Post) Adapt() *domain.Post {
	return &domain.Post{
		ID:     fmt.Sprintf("%d", p.ID),
		Driver: "tumblr",
		User: &domain.User{
			Name: p.BlogName,
		},
		Text:      p.joinText(),
		CreatedAt: time.Time(p.Date),
	}
}

func (p *Post) joinText() string {
	return fmt.Sprintf("%s\n%s", p.Summary, strings.Join(p.Tags, " "))
}

type date time.Time

func (d *date) UnmarshalJSON(data []byte) error {
	s := string(data)
	withoutQuotes := s[1 : len(s)-1]
	parsed, err := time.Parse("2006-01-02 15:04:05 MST", withoutQuotes)
	if err != nil {
		return err
	}
	*d = date(parsed.Local())

	return nil
}
