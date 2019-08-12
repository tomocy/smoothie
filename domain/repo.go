package domain

import "context"

type PostRepo interface {
	StreamPosts(context.Context) (<-chan Posts, <-chan error)
	FetchPosts() (Posts, error)
}
