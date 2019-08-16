package infra

import (
	"context"

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
