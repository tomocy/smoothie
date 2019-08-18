package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/tomocy/deverr"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/tomocy/smoothie/domain"
)

func NewYouTube(id, secret string) *YouTube {
	return &YouTube{
		oauth: oauth2Manager{
			cnf: oauth2.Config{
				ClientID: id, ClientSecret: secret,
				RedirectURL: "http://localhost/smoothie/youtube/authorization",
				Endpoint:    google.Endpoint,
				Scopes: []string{
					"https://www.googleapis.com/auth/youtube",
				},
			},
		},
	}
}

type YouTube struct {
	oauth oauth2Manager
}

func (y *YouTube) StreamPosts(ctx context.Context) (<-chan domain.Posts, <-chan error) {
	psCh, errCh := make(chan domain.Posts), make(chan error)
	go func() {
		defer func() {
			close(psCh)
			close(errCh)
		}()
		select {
		case <-ctx.Done():
		default:
			errCh <- deverr.NotImplemented
		}
	}()

	return psCh, errCh
}

func (y *YouTube) FetchPosts() (domain.Posts, error) {
	return nil, deverr.NotImplemented
}

func (y *YouTube) retreiveAuthorization() (*oauth2.Token, error) {
	url := y.authCodeURL()
	fmt.Printf("youtube: open this link: %s\n", url)

	return y.handleAuthorizationRedirect()
}

func (y *YouTube) authCodeURL() string {
	return y.oauth.authURL()
}

func (y *YouTube) handleAuthorizationRedirect() (*oauth2.Token, error) {
	return y.oauth.handleRedirect(context.Background(), nil, "/smoothie/youtube/authorization")
}

func (y *YouTube) do(r oauth2Req, dst interface{}) error {
	resp, err := r.do(context.Background(), y.oauth.cnf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if http.StatusBadRequest <= resp.StatusCode {
		return errors.New(resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(dst)
}
