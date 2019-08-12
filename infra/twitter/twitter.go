package twitter

import "time"

type Tweets []*Tweet

type Tweet struct {
	ID        string `json:"id_str"`
	Text      string `json:"text"`
	FullText  string `json:"full_text"`
	CreatedAt date   `json:"created_at"`
}

type date time.Time

func (d *date) UnmarshalJSON(data []byte) error {
	withoutQuotes := (string(data))[1 : len(data)-1]
	parsed, err := time.ParseInLocation(time.RubyDate, withoutQuotes, time.UTC)
	if err != nil {
		return err
	}
	*d = date(parsed.In(time.Local))

	return nil
}
