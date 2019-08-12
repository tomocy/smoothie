package domain

type PostRepo interface {
	FetchPosts() (Posts, error)
}
