package infra

import (
	"context"
	"fmt"

	"github.com/tomocy/deverr"
	"github.com/tomocy/smoothie/domain"
	"golang.org/x/oauth2"
)

type Gmail struct {
	oauth oauth2Manager
}

func (g *Gmail) StreamPosts(ctx context.Context) (<-chan domain.Posts, <-chan error) {
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

func (g *Gmail) FetchPosts() (domain.Posts, error) {
	return nil, deverr.NotImplemented
}

func (g *Gmail) retreiveAuthorization() (*oauth2.Token, error) {
	url := g.oauth.authURL()
	fmt.Printf("open this link: %s\n", url)

	return g.handleAuthorizationRedirect()
}

func (g *Gmail) handleAuthorizationRedirect() (*oauth2.Token, error) {
	return g.oauth.handleRedirect(context.Background(), nil, "/smoothie/gmail/authorization")
}
