package infra

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/tomocy/smoothie/domain"
	"github.com/tomocy/smoothie/infra/qiita"
)

type Qiita struct{}

func (q *Qiita) FetchPosts(args []string) (domain.Posts, error) {
	parsed := q.parseArgs(args)
	is, err := q.fetchItems(parsed.tag, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %s", err)
	}

	return is.Adapt(), nil
}

func (q *Qiita) parseArgs(args []string) qiitaArgs {
	var parsed qiitaArgs
	parsed.parse(args)

	return parsed
}

func (q *Qiita) fetchItems(tag string, params url.Values) (qiita.Items, error) {
	var is qiita.Items
	if err := q.do(req{
		method: http.MethodGet, url: q.endpoint("tags", tag, "items"), params: params,
	}, &is); err != nil {
		return nil, err
	}

	return is, nil
}

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

type qiitaArgs struct {
	tag string
}

func (as *qiitaArgs) parse(args []string) {
	if len(args) <= 0 {
		return
	}

	as.tag = args[0]
}
