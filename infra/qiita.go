package infra

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
)

type Qiita struct{}

func (q *Qiita) do(r req, dst interface{}) error {
	resp, err := r.do()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if http.StatusBadRequest <= resp.StatusCode {
		return errors.New(resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(dst)
}

func (q *Qiita) endpoint(ps ...string) string {
	parsed, _ := url.Parse("https://qiita.com/api/v2")
	ss := append([]string{parsed.Path}, ps...)
	parsed.Path = filepath.Join(ss...)
	return parsed.String()
}
