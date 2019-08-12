package twitter

import (
	"time"

	"github.com/tomocy/smoothie/domain"
)

type Tweets []*Tweet

func (ts Tweets) Adapt() domain.Posts {
	adapteds := make(domain.Posts, len(ts))
	for i, t := range ts {
		adapteds[i] = t.Adapt()
	}

	return adapteds
}

type Tweet struct {
	ID        string `json:"id_str"`
	Text      string `json:"text"`
	FullText  string `json:"full_text"`
	CreatedAt date   `json:"created_at"`
}

func (t *Tweet) Adapt() *domain.Post {
	text := t.Text
	if t.FullText != "" {
		text = t.FullText
	}
	return &domain.Post{
		ID: t.ID, Driver: "twitter", Text: text, CreatedAt: time.Time(t.CreatedAt),
	}
}

type date time.Time

func (d *date) UnmarshalJSON(data []byte) error {
	withoutQuotes := (string(data))[1 : len(data)-1]
	parsed, err := time.ParseInLocation(time.RubyDate, withoutQuotes, time.UTC)
	if err != nil {
		return err
	}
	*d = date(parsed.Local())

	return nil
}

type User struct {
	ID         string `json:"id_str"`
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
}
