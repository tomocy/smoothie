package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/tomocy/deverr"

	"github.com/tomocy/smoothie/domain"
	"github.com/tomocy/smoothie/infra/reddit"
	"golang.org/x/oauth2"
)

func NewReddit(id, secret string) *Reddit {
	return &Reddit{
		oauth: oauth2Config{
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
	}
}

type Reddit struct {
	oauth oauth2Config
}

func (r *Reddit) StreamPosts(context.Context) (<-chan domain.Posts, <-chan error) {
	errCh := make(chan error)
	go func() {
		close(errCh)
		errCh <- deverr.NotImplemented
	}()

	return nil, errCh
}

func (r *Reddit) FetchPosts() (domain.Posts, error) {
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

	var ps *reddit.Posts
	if err := r.do(oauth2Req{
		tok: tok, method: http.MethodGet, url: dst, params: params,
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
	fmt.Printf("reddit: open this link: %s\n", url)

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
	r.setRandomState()
	return r.oauth.cnf.AuthCodeURL(r.oauth.state, oauth2.SetAuthURLParam("duration", "permanent"))
}

func (r *Reddit) setRandomState() {
	r.oauth.state = fmt.Sprintf("%d", rand.Intn(10000))
}

func (r *Reddit) handleAuthorizationRedirect() (*oauth2.Token, error) {
	return r.oauth.handleRedirect(r.contextWithUserAgent(), nil, "/smoothie/reddit/authorization")
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
