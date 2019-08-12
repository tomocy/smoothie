package app

import "github.com/tomocy/smoothie/domain"

type PostUsecase struct {
	repos map[string]domain.PostRepo
}
