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
	"github.com/tomocy/smoothie/infra/reddit"
	"golang.org/x/oauth2"
)

func NewReddit(id, secret string, presenter authURLPresenter) *Reddit {
	return &Reddit{
		oauth: oauth2Manager{
			cnf: oauth2.Config{
				ClientID:     id,
				ClientSecret: secret,
				RedirectURL:  "http://localhost/smoothie/reddit/authorization",
				Endpoint: oauth2.Endpoint{
					AuthURL:   "https://www.reddit.com/api/v1/authorize",
					TokenURL:  "https://www.reddit.com/api/v1/access_token",
					AuthStyle: oauth2.AuthStyleInHeader,
				},
				Scopes: []string{
					"read", "identity", "mysubreddits",
				},
			},
		},
		presenter: presenter,
	}
}

type Reddit struct {
	oauth     oauth2Manager
	presenter authURLPresenter
}

func (r *Reddit) StreamPosts(ctx context.Context, args []string) (<-chan domain.Posts, <-chan error) {
	psCh, errCh := r.streamPosts(ctx, r.endpoint("/new"), nil)
	ch := make(chan domain.Posts)
	go func() {
		defer close(ch)
		for ps := range psCh {
			select {
			case <-ctx.Done():
				return
			default:
				ch <- ps.Adapt()
			}
		}
	}()

	return ch, errCh
}

func (r *Reddit) streamPosts(ctx context.Context, dst string, params url.Values) (<-chan *reddit.Posts, <-chan error) {
	psCh, errCh := make(chan *reddit.Posts), make(chan error)
	go func() {
		defer func() {
			close(psCh)
			close(errCh)
		}()

		lastID := r.fetchAndSendPosts(dst, params, psCh, errCh)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Minute):
				if lastID != "" {
					if params == nil {
						params = make(url.Values)
					}
					params.Set("before", lastID)
				}
				if id := r.fetchAndSendPosts(dst, params, psCh, errCh); id != "" {
					lastID = id
				}
			}
		}
	}()

	return psCh, errCh
}

func (r *Reddit) fetchAndSendPosts(dst string, params url.Values, psCh chan<- *reddit.Posts, errCh chan<- error) string {
	ps, err := r.fetchPosts(dst, params)
	if err != nil {
		return ""
	}
	if len(ps.Data.Children) <= 0 {
		return ""
	}

	lastID := ps.Data.Children[0].Data.Name
	psCh <- ps

	return lastID
}

func (r *Reddit) FetchPosts(args []string) (domain.Posts, error) {
	ps, err := r.fetchPosts(r.endpoint("/new"), nil)
	if err != nil {
		return nil, err
	}

	return ps.Adapt(), nil
}

func (r *Reddit) fetchPosts(dst string, params url.Values) (*reddit.Posts, error) {
	tok, err := r.retreiveAuthorization()
	if err != nil {
		return nil, err
	}

	assured := r.assureDefaultParams(params)
	var ps *reddit.Posts
	if err := r.do(oauth2Req{
		tok: tok,
		req: req{method: http.MethodGet, url: dst, params: assured},
	}, &ps); err != nil {
		return nil, err
	}
	if err := r.saveAccessToken(tok); err != nil {
		return nil, err
	}

	return ps, nil
}

func (r *Reddit) retreiveAuthorization() (*oauth2.Token, error) {
	if cnf, err := r.loadConfig(); err == nil && !cnf.isZero() {
		return cnf.AccessToken, nil
	}

	url := r.authCodeURL()
	r.presenter.ShowAuthURL(url)

	return r.handleAuthorizationRedirect()
}

func (r *Reddit) loadConfig() (redditConfig, error) {
	cnf, err := loadConfig()
	if err != nil {
		return redditConfig{}, err
	}

	return cnf.Reddit, nil
}

func (r *Reddit) authCodeURL() string {
	return r.oauth.authURL(oauth2.SetAuthURLParam("duration", "permanent"))
}

func (r *Reddit) handleAuthorizationRedirect() (*oauth2.Token, error) {
	return r.oauth.handleRedirect(r.contextWithUserAgent(), nil, "/smoothie/reddit/authorization")
}

func (r *Reddit) assureDefaultParams(params url.Values) url.Values {
	assured := params
	if assured == nil {
		assured = make(url.Values)
	}
	assured.Set("limit", "100")

	return assured
}

func (r *Reddit) do(req oauth2Req, dst interface{}) error {
	resp, err := req.do(r.contextWithUserAgent(), r.oauth.cnf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if http.StatusBadRequest <= resp.StatusCode {
		return errors.New(resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(dst)
}

func (r *Reddit) contextWithUserAgent() context.Context {
	return context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
		Transport: new(withUserAgent),
	})
}

func (r *Reddit) saveAccessToken(tok *oauth2.Token) error {
	loaded, err := r.loadConfig()
	if err != nil {
		return err
	}
	loaded.AccessToken = tok

	return r.saveConfig(loaded)
}

func (r *Reddit) saveConfig(cnf redditConfig) error {
	loaded, err := loadConfig()
	if err != nil {
		return err
	}
	loaded.Reddit = cnf

	return saveConfig(loaded)
}

func (r *Reddit) endpoint(ps ...string) string {
	parsed, _ := url.Parse("https://oauth.reddit.com")
	ss := append([]string{parsed.Path}, ps...)
	parsed.Path = filepath.Join(ss...)
	return parsed.String()
}
