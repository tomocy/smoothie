package infra

import (
	"context"

	"github.com/tomocy/deverr"

	"github.com/tomocy/smoothie/domain"
)

type GitHub struct{}

func (g *GitHub) StreamPosts(ctx context.Context) (<-chan domain.Posts, <-chan error) {
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

func (g *GitHub) FetchPosts() (domain.Posts, error) {
	return nil, deverr.NotImplemented
}
