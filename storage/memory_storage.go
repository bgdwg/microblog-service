package storage

import (
	"errors"
	"microblogging-service/data"
	"sync"
)

type MemoryStorage struct {
	Posts       map[data.PostId]*data.Post
	UserPosts   map[data.UserId][]*data.Post
	Mutex       *sync.RWMutex
}

func NewStorage() *MemoryStorage {
	return &MemoryStorage{
		Posts:     make(map[data.PostId]*data.Post),
		UserPosts: make(map[data.UserId][]*data.Post),
		Mutex:     new(sync.RWMutex),
	}
}

func (storage *MemoryStorage) AddPost(post *data.Post) error {
	storage.Mutex.Lock()
	defer storage.Mutex.Unlock()
	storage.Posts[post.Id] = post
	posts, found := storage.UserPosts[post.AuthorId]
	if !found {
		posts = make([]*data.Post, 0)
	}
	storage.UserPosts[post.AuthorId] = append([]*data.Post{post}, posts...)
	return nil
}

func (storage *MemoryStorage) GetPost(postId data.PostId) (*data.Post, error) {
	storage.Mutex.RLock()
	defer storage.Mutex.RUnlock()
	post, found := storage.Posts[postId]
	if !found {
		return nil, errors.New("post not found")
	}
	return post, nil
}

func (storage *MemoryStorage) GetUserPosts(userId data.UserId, offset, limit int) ([]*data.Post, int, error) {
	storage.Mutex.RLock()
	defer storage.Mutex.RUnlock()
	posts, found := storage.UserPosts[userId]
	if !found {
		return nil, 0, nil
	}
	postsSlice := make([]*data.Post, 0)
	for i := offset; i < offset+limit && i < len(posts); i++ {
		postsSlice = append(postsSlice, posts[i])
	}
	return postsSlice, len(posts), nil
}

