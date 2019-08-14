package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tomocy/smoothie/domain"
)

func NewPostUsecase(repos map[string]domain.PostRepo) *PostUsecase {
	return &PostUsecase{
		repos: repos,
	}
}

type PostUsecase struct {
	repos map[string]domain.PostRepo
}

func (u *PostUsecase) StreamPostsOfDrivers(ctx context.Context, ds ...string) (<-chan domain.Posts, <-chan error) {
	psChs, errChs := make([]<-chan domain.Posts, len(ds)), make([]<-chan error, len(ds))
	for i, d := range ds {
		psChs[i], errChs[i] = u.streamPosts(ctx, d)
	}

	return u.fanInPosts(ctx, psChs...), u.fanInErrors(ctx, errChs...)
}

func (u *PostUsecase) streamPosts(ctx context.Context, d string) (<-chan domain.Posts, <-chan error) {
	repo, ok := u.repos[d]
	if !ok {
		psCh, errCh := make(chan domain.Post), make(chan error)
		go func() {
			defer func() {
				close(psCh)
				close(errCh)
			}()
			errCh <- fmt.Errorf("unknown driver: %s", d)
		}()
		return nil, errCh
	}

	return repo.StreamPosts(ctx)
}

func (u *PostUsecase) fanInPosts(ctx context.Context, chs ...<-chan domain.Posts) <-chan domain.Posts {
	fannedInCh := make(chan domain.Posts)
	go func() {
		defer close(fannedInCh)
		fanIn := func(dst chan<- domain.Posts, srces []<-chan domain.Posts) bool {
			var fannedIn domain.Posts
			for _, src := range srces {
				select {
				case ps := <-src:
					fannedIn = append(fannedIn, ps...)
				default:
				}
			}

			if len(fannedIn) <= 0 {
				return false
			}
			fannedIn.SortByNewest()
			dst <- fannedIn

			return true
		}
	waiting:
		for {
			select {
			case <-time.After(1 * time.Second):
				if fanIn(fannedInCh, chs) {
					break waiting
				}
			}
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Minute):
				fanIn(fannedInCh, chs)
			}
		}
	}()

	return fannedInCh
}

func (u *PostUsecase) fanInErrors(ctx context.Context, chs ...<-chan error) <-chan error {
	fannedInCh := make(chan error)
	go func() {
		defer close(fannedInCh)
		var wg sync.WaitGroup
		for _, ch := range chs {
			wg.Add(1)
			go func(ch <-chan error) {
				defer wg.Done()
				for err := range ch {
					select {
					case <-ctx.Done():
						return
					default:
						fannedInCh <- err
					}
				}
			}(ch)
		}
		wg.Wait()
	}()

	return fannedInCh
}

func (u *PostUsecase) FetchPostsOfDrivers(ds ...string) (domain.Posts, error) {
	var fetcheds domain.Posts
	for _, d := range ds {
		ps, err := u.fetchPost(d)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch post of drivers: %s", err)
		}

		fetcheds = append(fetcheds, ps...)
	}

	fetcheds.SortByNewest()

	return fetcheds, nil
}

func (u *PostUsecase) fetchPost(d string) (domain.Posts, error) {
	repo, ok := u.repos[d]
	if !ok {
		return nil, fmt.Errorf("unknown driver: %s", d)
	}

	return repo.FetchPosts()
}
