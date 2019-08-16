package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/garyburd/go-oauth/oauth"
	"github.com/tomocy/deverr"

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

func (t *Tumblr) FetchPosts() (domain.Posts, error) {
	ps, err := t.fetchPosts(t.endpoint("/user/dashboard"), nil)
	if err != nil {
		return nil, err
	}

	return ps.Adapt(), nil
}

func (t *Tumblr) fetchPosts(dst string, params url.Values) (*tumblr.Posts, error) {
	cred, err := t.retreiveAuthorization()
	if err != nil {
		return nil, err
	}

	var ps *tumblr.Posts
	if err := t.do(oauthReq{
		cred: cred, method: http.MethodGet, url: dst, params: params,
	}, &ps); err != nil {
		return nil, err
	}
	if err := t.saveAccessToken(cred); err != nil {
		return nil, err
	}

	return ps, nil
}

func (t *Tumblr) retreiveAuthorization() (*oauth.Credentials, error) {
	if cnf, err := t.loadConfig(); err == nil && !cnf.isZero() {
		return cnf.AccessCredentials, nil
	}

	url, err := t.authorizationURL()
	if err != nil {
		return nil, err
	}
	fmt.Printf("open this url: %s\n", url)

	return t.handleAuthorizationRedirect()
}

func (t *Tumblr) loadConfig() (tumblrConfig, error) {
	cnf, err := loadConfig()
	if err != nil {
		return tumblrConfig{}, err
	}

	return cnf.Tumblr, nil
}

func (t *Tumblr) authorizationURL() (string, error) {
	temp, err := t.oauth.client.RequestTemporaryCredentials(http.DefaultClient, "", nil)
	if err != nil {
		return "", err
	}
	t.oauth.temp = temp

	return t.oauth.client.AuthorizationURL(temp, nil), nil
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
