package domain

type PostRepo interface {
	FetchPosts() ([]*Post, error)
}
