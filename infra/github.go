package infra

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/tomocy/smoothie/domain"
	githubPkg "github.com/tomocy/smoothie/infra/github"
)

type GitHubIssue struct{}

func (g *GitHubIssue) StreamPosts(ctx context.Context) (<-chan domain.Posts, <-chan error) {
	isCh, errCh := g.streamIssues(ctx, "golang", "go", nil)
	ch := make(chan domain.Posts)
	go func() {
		defer close(ch)
		for is := range isCh {
			select {
			case <-ctx.Done():
				return
			default:
				ch <- is.Adapt()
			}
		}
	}()

	return ch, errCh
}

func (g *GitHubIssue) streamIssues(ctx context.Context, owner, repo string, params url.Values) (<-chan githubPkg.Issues, <-chan error) {
	isCh, errCh := make(chan githubPkg.Issues), make(chan error)
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

func (g *GitHubIssue) fetchAndSendIssues(owner, repo string, params url.Values, isCh chan<- githubPkg.Issues, errCh chan<- error) time.Time {
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

func (g *GitHubIssue) FetchPosts() (domain.Posts, error) {
	is, err := g.fetchIssues("golang", "go", nil)
	if err != nil {
		return nil, err
	}

	return is.Adapt(), nil
}

func (g *GitHubIssue) fetchIssues(owner, repo string, params url.Values) (githubPkg.Issues, error) {
	var is githubPkg.Issues
	if err := g.do(req{
		method: http.MethodGet, url: g.endpoint("repos", owner, repo, "issues"), params: params,
	}, &is); err != nil {
		return nil, err
	}

	return is, nil
}

func (g *GitHubIssue) do(r req, dst interface{}) error {
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

func (g *GitHubIssue) endpoint(ps ...string) string {
	parsed, _ := url.Parse("https://api.github.com")
	ss := append([]string{parsed.Path}, ps...)
	parsed.Path = filepath.Join(ss...)
	return parsed.String()
}

type github struct{}

func (g *github) do(r req, dst interface{}) error {
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

func (g *github) endpoint(ps ...string) string {
	parsed, _ := url.Parse("https://api.github.com")
	ss := append([]string{parsed.Path}, ps...)
	parsed.Path = filepath.Join(ss...)
	return parsed.String()
}
