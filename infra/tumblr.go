package infra

import (
	"context"
	"fmt"
	"net/http"

	"github.com/garyburd/go-oauth/oauth"
	"github.com/tomocy/deverr"

	"github.com/tomocy/smoothie/domain"
)

type Tumblr struct {
	oauthClient oauth.Client
}

func (t *Tumblr) StreamPostsOfDrivers(ctx context.Context) (<-chan domain.Posts, <-chan error) {
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
	return nil, deverr.NotImplemented
}

func (t *Tumblr) retreiveAuthorization() (*oauth.Credentials, error) {
	temp, err := t.oauthClient.RequestTemporaryCredentials(http.DefaultClient, "", nil)
	if err != nil {
		return nil, err
	}

	return t.requestClientAuthorization(temp)
}

func (t *Tumblr) requestClientAuthorization(temp *oauth.Credentials) (*oauth.Credentials, error) {
	url := t.oauthClient.AuthorizationURL(temp, nil)
	fmt.Printf("open this url: %s\n", url)

	fmt.Print("PIN: ")
	var pin string
	fmt.Scan(&pin)

	token, _, err := t.oauthClient.RequestToken(http.DefaultClient, temp, pin)
	return token, err
}
