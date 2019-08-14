package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/tomocy/smoothie/infra/reddit"
	"golang.org/x/oauth2"
)

type Reddit struct {
	oauth oauth2Config
}

func (r *Reddit) fetchPosts(dst string, params url.Values) (reddit.Posts, error) {
	tok, err := r.retreiveAuthorization()
	if err != nil {
		return nil, err
	}

	var ps reddit.Posts
	if err := r.do(oauth2Req{
		tok: tok, method: http.MethodGet, url: dst, params: params,
	}, &ps); err != nil {
		return nil, err
	}

	return ps, nil
}

func (r *Reddit) retreiveAuthorization() (*oauth2.Token, error) {
	if cnf, err := r.loadConfig(); err == nil {
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

func (r *Reddit) handleAuthorizationRedirect() (*oauth2.Token, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
		Transport: new(withUserAgent),
	})

	return r.oauth.handleRedirect(ctx, nil, "/smoothie/reddit/authorization")
}

func (r *Reddit) authCodeURL(params ...oauth2.AuthCodeOption) string {
	r.setRandomState()
	return r.oauth.cnf.AuthCodeURL(r.oauth.state, params...)
}

func (r *Reddit) setRandomState() {
	r.oauth.state = fmt.Sprintf("%d", rand.Intn(10000))
}

func (r *Reddit) do(req oauth2Req, dst interface{}) error {
	resp, err := req.do(oauth2.NoContext, r.oauth.cnf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(dst)
}
