package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/tomocy/smoothie/domain"
	"github.com/tomocy/smoothie/infra/qiita"
)

type Qiita struct{}

func (q *Qiita) StreamPosts(ctx context.Context, args []string) (<-chan domain.Posts, <-chan error) {
	parsed := q.parseArgs(args)
	isCh, errCh := q.streamItems(ctx, parsed.tag, nil)
	ch := make(chan domain.Posts)
	go func() {
		defer close(ch)

		for is := range isCh {
			ch <- is.Adapt()
		}
	}()

	return ch, errCh
}

func (q *Qiita) streamItems(ctx context.Context, tag string, params url.Values) (<-chan qiita.Items, <-chan error) {
	isCh, errCh := make(chan qiita.Items), make(chan error)
	go func() {
		defer func() {
			close(isCh)
			close(errCh)
		}()

		lastCreatedAt := q.fetchAndSendItems(tag, params, isCh, errCh)
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case <-time.After(time.Second):
				if !lastCreatedAt.IsZero() {
					if params == nil {
						params = make(url.Values)
					}
					params.Add("query", fmt.Sprintf("created:>%s", lastCreatedAt))
					log.Println(params.Get("query"))
				}
				if createdAt := q.fetchAndSendItems(tag, params, isCh, errCh); !createdAt.IsZero() {
					lastCreatedAt = createdAt
				}
			}
		}
	}()

	return isCh, errCh
}

func (q *Qiita) fetchAndSendItems(
	tag string, params url.Values,
	isCh chan<- qiita.Items, errCh chan<- error,
) time.Time {
	is, err := q.fetchItems(tag, params)
	if err != nil {
		errCh <- err
		return time.Time{}
	}
	if len(is) <= 0 {
		return time.Time{}
	}

	lastCreatedAt := is[0].CreatedAt
	isCh <- is

	return lastCreatedAt
}

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
