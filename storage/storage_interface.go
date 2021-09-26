package storage

import "microblogging-service/data"

type Storage interface {
	AddPost(post *data.Post) error
	GetPost(postId data.PostId) (*data.Post, error)
	GetUserPosts (userId data.UserId, offset, limit int) ([]*data.Post, int, error)
}
