package domain

import "context"

type PostRepo interface {
	StreamPosts(context.Context, []string) (<-chan Posts, <-chan error)
	FetchPosts([]string) (Posts, error)
}
