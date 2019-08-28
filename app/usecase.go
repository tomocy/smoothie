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

func (u *PostUsecase) StreamPostsOfDrivers(ctx context.Context, ds ...Driver) (<-chan domain.Posts, <-chan error) {
	psChs, errChs := make([]<-chan domain.Posts, len(ds)), make([]<-chan error, len(ds))
	for i, d := range ds {
		psChs[i], errChs[i] = u.streamPosts(ctx, d)
	}

	return u.fanInPosts(ctx, psChs...), u.fanInErrors(ctx, errChs...)
}

func (u *PostUsecase) streamPosts(ctx context.Context, d Driver) (<-chan domain.Posts, <-chan error) {
	repo, ok := u.repos[d.Name]
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

	return repo.StreamPosts(ctx, d.Args)
}

func (u *PostUsecase) fanInPosts(ctx context.Context, chs ...<-chan domain.Posts) <-chan domain.Posts {
	fannedInCh := make(chan domain.Posts)
	go func() {
		defer close(fannedInCh)
		fanIn := func(dst chan<- domain.Posts, srces []<-chan domain.Posts, doSkip bool) {
			var fannedIn domain.Posts
			for _, src := range srces {
				if doSkip {
					select {
					case ps := <-src:
						fannedIn = append(fannedIn, ps...)
					default:
					}
				} else {
					ps := <-src
					fannedIn = append(fannedIn, ps...)
				}
			}
			if len(fannedIn) <= 0 {
				return
			}

			fannedIn.SortByNewest()
			dst <- fannedIn
		}

		fanIn(fannedInCh, chs, false)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Minute):
				fanIn(fannedInCh, chs, true)
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

func (u *PostUsecase) FetchPostsOfDrivers(ds ...Driver) (domain.Posts, error) {
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

func (u *PostUsecase) fetchPost(d Driver) (domain.Posts, error) {
	repo, ok := u.repos[d.Name]
	if !ok {
		return nil, fmt.Errorf("unknown driver: %s", d)
	}

	return repo.FetchPosts(d.Args)
}

type Driver struct {
	Name string
	Args []string
}
