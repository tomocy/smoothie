package infra

import (
	"context"

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
