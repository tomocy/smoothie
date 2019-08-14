package reddit

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type Posts []*Post

type Post struct {
	Name                  string        `json:"name"`
	SubredditNamePrefixed string        `json:"subreddit_name_prefixed"`
	Author                string        `json:"author"`
	Title                 string        `json:"title"`
	SelfText              string        `json:"selftext"`
	CreatedUTC            unixTimestamp `json:"created_utc"`
}

type unixTimestamp time.Time

func (t *unixTimestamp) UnmarshalJSON(data []byte) error {
	parsed, err := t.parse(string(data))
	if err != nil {
		return err
	}
	*t = unixTimestamp(parsed.Local())

	return nil
}

func (t *unixTimestamp) parse(ts string) (time.Time, error) {
	splited := strings.Split(ts, ".")
	if len(splited) != 2 {
		return time.Time{}, errors.New("invalid format of unix timestamp: the format should be sec.nsec")
	}
	sec, err := strconv.ParseInt(splited[0], 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	nsec, err := strconv.ParseInt(splited[1], 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(sec, nsec), nil
}
