package storage

import (
	"context"
	"errors"
	"fmt"
	"microblogging-service/internal/data"
)

var (
	ErrBase = errors.New("storage")
	ErrCollision = fmt.Errorf("%w.collision", ErrBase)
	ErrNotFound = fmt.Errorf("%w.not_found", ErrBase)
)

type Storage interface {
	AddPost(ctx context.Context, post *data.Post) error
	GetPost(ctx context.Context, postId data.PostId) (*data.Post, error)
	GetUserPosts (ctx context.Context, userId data.UserId,
		token data.PageToken, limit int) ([]*data.Post, data.PageToken, error)
	UpdatePost(ctx context.Context, post *data.Post) error
}
