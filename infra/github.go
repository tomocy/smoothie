package infra

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/tomocy/deverr"

	"github.com/tomocy/smoothie/domain"
	"github.com/tomocy/smoothie/infra/github"
)

type GitHub struct{}

func (g *GitHub) StreamPosts(ctx context.Context) (<-chan domain.Posts, <-chan error) {
	psCh, errCh := make(chan domain.Posts), make(chan error)
	go func() {
		defer func() {
			close(psCh)
			close(errCh)
		}()

		select {
		case <-ctx.Done():
			return
		default:
			errCh <- deverr.NotImplemented
		}
	}()

	return psCh, errCh
}

func (g *GitHub) streamIssues(ctx context.Context, owner, repo string, params url.Values) (<-chan github.Issues, <-chan error) {
	isCh, errCh := make(chan github.Issues), make(chan error)
	go func() {
		defer func() {
			close(isCh)
			close(errCh)
		}()
		lastCreatedAt := g.fetchAndSendIssues(owner, repo, params, isCh, errCh)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Minute):
				if !lastCreatedAt.IsZero() {
					if params == nil {
						params = make(url.Values)
					}
					params.Set("since", lastCreatedAt.Format(time.RFC3339))
				}
				if createdAt := g.fetchAndSendIssues(owner, repo, params, isCh, errCh); !createdAt.IsZero() {
					lastCreatedAt = createdAt
				}
			}
		}
	}()

	return isCh, errCh
}

func (g *GitHub) fetchAndSendIssues(owner, repo string, params url.Values, isCh chan<- github.Issues, errCh chan<- error) time.Time {
	is, err := g.fetchIssues(owner, repo, params)
	if err != nil {
		errCh <- err
		return time.Time{}
	}
	if len(is) <= 0 {
		return time.Time{}
	}

	isCh <- is
	return is[0].CreatedAt
}

func (g *GitHub) FetchPosts() (domain.Posts, error) {
	is, err := g.fetchIssues("golang", "go", nil)
	if err != nil {
		return nil, err
	}

	return is.Adapt(), nil
}

func (g *GitHub) fetchIssues(owner, repo string, params url.Values) (github.Issues, error) {
	var is github.Issues
	if err := g.do(req{
		method: http.MethodGet, url: g.endpoint("repos", owner, repo, "issues"), params: params,
	}, &is); err != nil {
		return nil, err
	}

	return is, nil
}

func (g *GitHub) do(r req, dst interface{}) error {
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

func (g *GitHub) endpoint(ps ...string) string {
	parsed, _ := url.Parse("https://api.github.com")
	ss := append([]string{parsed.Path}, ps...)
	parsed.Path = filepath.Join(ss...)
	return parsed.String()
}
