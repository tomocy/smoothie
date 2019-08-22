package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/garyburd/go-oauth/oauth"

	"github.com/tomocy/smoothie/domain"
	"github.com/tomocy/smoothie/infra/tumblr"
)

func NewTumblr(id, secret string) *Tumblr {
	return &Tumblr{
		oauth: oauthManager{
			client: oauth.Client{
				TemporaryCredentialRequestURI: "https://www.tumblr.com/oauth/request_token",
				ResourceOwnerAuthorizationURI: "https://www.tumblr.com/oauth/authorize",
				TokenRequestURI:               "https://www.tumblr.com/oauth/access_token",
				Credentials: oauth.Credentials{
					Token: id, Secret: secret,
				},
			},
		},
	}
}

type Tumblr struct {
	oauth oauthManager
}

func (t *Tumblr) StreamPosts(ctx context.Context) (<-chan domain.Posts, <-chan error) {
	psCh, errCh := t.streamPosts(ctx, t.endpoint("/user/dashboard"), nil)
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

func (t *Tumblr) streamPosts(ctx context.Context, dst string, params url.Values) (<-chan tumblr.Posts, <-chan error) {
	psCh, errCh := make(chan tumblr.Posts), make(chan error)
	go func() {
		defer func() {
			close(psCh)
			close(errCh)
		}()
		lastID := t.fetchAndSendPosts(dst, params, psCh, errCh)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Minute):
				if lastID != "" {
					if params == nil {
						params = make(url.Values)
					}
					params.Set("since_id", lastID)
				}
				if id := t.fetchAndSendPosts(dst, params, psCh, errCh); id != "" {
					lastID = id
				}
			}
		}
	}()

	return psCh, errCh
}

func (t *Tumblr) fetchAndSendPosts(dst string, params url.Values, psCh chan<- tumblr.Posts, errCh chan<- error) string {
	ps, err := t.fetchPosts(dst, params)
	if err != nil {
		errCh <- err
		return ""
	}
	if len(ps) <= 0 {
		return ""
	}

	psCh <- ps
	return fmt.Sprintf("%d", ps[0].ID)
}

func (t *Tumblr) FetchPosts() (domain.Posts, error) {
	ps, err := t.fetchPosts(t.endpoint("/user/dashboard"), nil)
	if err != nil {
		return nil, err
	}

	return ps.Adapt(), nil
}

func (t *Tumblr) fetchPosts(dst string, params url.Values) (tumblr.Posts, error) {
	cred, err := t.retreiveAuthorization()
	if err != nil {
		return nil, err
	}

	var resp *tumblr.Resp
	if err := t.do(oauthReq{
		cred: cred, method: http.MethodGet, url: dst, params: params,
	}, &resp); err != nil {
		return nil, err
	}
	if err := t.saveAccessToken(cred); err != nil {
		return nil, err
	}

	return resp.Resp.Posts, nil
}

func (t *Tumblr) retreiveAuthorization() (*oauth.Credentials, error) {
	if cnf, err := t.loadConfig(); err == nil && !cnf.isZero() {
		return cnf.AccessCredentials, nil
	}

	url, err := t.oauth.authURL(nil)
	if err != nil {
		return nil, err
	}
	fmt.Printf("tumblr: open this url: %s\n", url)

	return t.handleAuthorizationRedirect()
}

func (t *Tumblr) loadConfig() (tumblrConfig, error) {
	cnf, err := loadConfig()
	if err != nil {
		return tumblrConfig{}, err
	}

	return cnf.Tumblr, nil
}

func (t *Tumblr) handleAuthorizationRedirect() (*oauth.Credentials, error) {
	return t.oauth.handleRedirect(context.Background(), "/smoothie/tumblr/authorization")
}

func (t *Tumblr) do(r oauthReq, dst interface{}) error {
	resp, err := r.do(t.oauth.client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(dst)
}

func (t *Tumblr) saveAccessToken(cred *oauth.Credentials) error {
	loaded, err := t.loadConfig()
	if err != nil {
		return err
	}
	loaded.AccessCredentials = cred

	return t.saveConfig(loaded)
}

func (t *Tumblr) saveConfig(cnf tumblrConfig) error {
	loaded, err := loadConfig()
	if err != nil {
		return err
	}
	loaded.Tumblr = cnf

	return saveConfig(loaded)
}

func (t *Tumblr) endpoint(ps ...string) string {
	parsed, _ := url.Parse("https://api.tumblr.com/v2")
	ss := append([]string{parsed.Path}, ps...)
	parsed.Path = filepath.Join(ss...)
	return parsed.String()
}
