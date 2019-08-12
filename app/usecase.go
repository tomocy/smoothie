package app

import (
	"fmt"

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
