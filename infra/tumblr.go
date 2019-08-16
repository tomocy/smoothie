package infra

import (
	"context"
	"fmt"
	"net/http"

	"github.com/garyburd/go-oauth/oauth"
	"github.com/tomocy/deverr"

	"github.com/tomocy/smoothie/domain"
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
	_, err := t.retreiveAuthorization()
	if err != nil {
		return nil, err
	}

	return nil, deverr.NotImplemented
}

func (t *Tumblr) retreiveAuthorization() (*oauth.Credentials, error) {
	temp, err := t.oauth.client.RequestTemporaryCredentials(http.DefaultClient, "", nil)
	if err != nil {
		return nil, err
	}

	return t.requestClientAuthorization(temp)
}

func (t *Tumblr) requestClientAuthorization(temp *oauth.Credentials) (*oauth.Credentials, error) {
	url := t.oauth.client.AuthorizationURL(temp, nil)
	fmt.Printf("open this url: %s\n", url)

	fmt.Print("PIN: ")
	var pin string
	fmt.Scan(&pin)

	token, _, err := t.oauth.client.RequestToken(http.DefaultClient, temp, pin)
	return token, err
}
