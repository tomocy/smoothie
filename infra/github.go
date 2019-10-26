package infra

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/tomocy/smoothie/domain"
	githubPkg "github.com/tomocy/smoothie/infra/github"
)

type GitHubEvents struct {
	github
}

func (g *GitHubEvents) StreamPosts(ctx context.Context, args []string) (<-chan domain.Posts, <-chan error) {
	parsed := g.parseArgs(args)
	esCh, errCh := g.streamEvents(ctx, parsed.uname, nil, nil)
	ch := make(chan domain.Posts)
	go func() {
		defer close(ch)
		for es := range esCh {
			ch <- es.Adapt()
		}
	}()

	return ch, errCh
}

func (g *GitHubEvents) streamEvents(ctx context.Context, uname string, header http.Header, params url.Values) (<-chan githubPkg.Events, <-chan error) {
	esCh, errCh := make(chan githubPkg.Events), make(chan error)
	go func() {
		defer func() {
			close(esCh)
			close(errCh)
		}()

		lastETag := g.fetchAndSendEvents(uname, header, params, esCh, errCh)
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case <-time.After(time.Minute):
				if lastETag != "" {
					if header == nil {
						header = make(http.Header)
					}
					header.Set("If-None-Match", lastETag)
				}
				if etag := g.fetchAndSendEvents(uname, header, params, esCh, errCh); etag != "" {
					lastETag = etag
				}
			}
		}
	}()

	return esCh, errCh
}

func (g *GitHubEvents) fetchAndSendEvents(uname string, header http.Header, params url.Values, esCh chan<- githubPkg.Events, errCh chan<- error) string {
	es, etag, err := g.fetchEvents(uname, header, params)
	if err != nil {
		errCh <- err
		return ""
	}
	if len(es) <= 0 {
		return ""
	}

	esCh <- es
	return etag
}

func (g *GitHubEvents) FetchPosts(args []string) (domain.Posts, error) {
	parsed := g.parseArgs(args)
	es, _, err := g.fetchEvents(parsed.uname, nil, nil)
	if err != nil {
		return nil, err
	}

	return es.Adapt(), nil
}

func (g *GitHubEvents) fetchEvents(uname string, header http.Header, params url.Values) (githubPkg.Events, string, error) {
	var es githubPkg.Events
	dst := &resp{
		body: &es,
	}
	if err := g.do(req{
		method: http.MethodGet, url: g.endpoint("users", uname, "received_events"), header: header, params: params,
	}, dst); err != nil {
		return nil, "", err
	}

	return es, dst.header.Get("ETag"), nil
}

func (g *GitHubEvents) parseArgs(args []string) githubEventsArgs {
	var parsed githubEventsArgs
	parsed.parse(args)

	return parsed
}

type githubEventsArgs struct {
	uname string
}

func (as *githubEventsArgs) parse(args []string) {
	if len(args) <= 0 {
		return
	}

	as.uname = args[0]
}

type GitHubIssues struct {
	github
}

func (g *GitHubIssues) StreamPosts(ctx context.Context, args []string) (<-chan domain.Posts, <-chan error) {
	parsed := g.parseArgs(args)
	isCh, errCh := g.streamIssues(ctx, parsed.owner, parsed.repo, nil)
	ch := make(chan domain.Posts)
	go func() {
		defer close(ch)
		for is := range isCh {
			ch <- is.Adapt()
		}
	}()

	return ch, errCh
}

func (g *GitHubIssues) streamIssues(ctx context.Context, owner, repo string, params url.Values) (<-chan githubPkg.Issues, <-chan error) {
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

func (g *GitHubIssues) fetchAndSendIssues(owner, repo string, params url.Values, isCh chan<- githubPkg.Issues, errCh chan<- error) time.Time {
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

func (g *GitHubIssues) FetchPosts(args []string) (domain.Posts, error) {
	parsed := g.parseArgs(args)
	is, err := g.fetchIssues(parsed.owner, parsed.repo, nil)
	if err != nil {
		return nil, err
	}

	return is.Adapt(), nil
}

func (g *GitHubIssues) fetchIssues(owner, repo string, params url.Values) (githubPkg.Issues, error) {
	var is githubPkg.Issues
	dst := &resp{
		body: &is,
	}
	if err := g.do(req{
		method: http.MethodGet, url: g.endpoint("repos", owner, repo, "issues"), params: params,
	}, dst); err != nil {
		return nil, err
	}

	return is, nil
}

func (g *GitHubIssues) parseArgs(args []string) githubIssuesArgs {
	var parsed githubIssuesArgs
	parsed.parse(args)

	return parsed
}

type githubIssuesArgs struct {
	owner, repo string
}

func (as *githubIssuesArgs) parse(args []string) {
	if len(args) <= 0 {
		return
	}

	splited := strings.Split(args[0], "/")
	if len(splited) == 2 {
		as.owner = splited[0]
		as.repo = splited[1]
	}
}

type github struct{}

func (g *github) do(r req, dst *resp) error {
	resp, err := r.do()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if http.StatusBadRequest <= resp.StatusCode {
		return errors.New(resp.Status)
	}
	if resp.StatusCode == http.StatusNotModified {
		return nil
	}

	dst.header = resp.Header
	return json.NewDecoder(resp.Body).Decode(dst.body)
}

func (g *github) endpoint(ps ...string) string {
	parsed, _ := url.Parse("https://api.github.com")
	ss := append([]string{parsed.Path}, ps...)
	parsed.Path = filepath.Join(ss...)
	return parsed.String()
}
