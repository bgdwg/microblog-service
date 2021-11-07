package storage

import (
	"context"
	"errors"
	"microblogging-service/data"
	"strconv"
	"sync"
)

type MemoryStorage struct {
	Posts       map[data.PostId]*data.Post
	UserPosts   map[data.UserId][]*data.Post
	Mutex       *sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Posts:     make(map[data.PostId]*data.Post),
		UserPosts: make(map[data.UserId][]*data.Post),
		Mutex:     new(sync.RWMutex),
	}
}

func (storage *MemoryStorage) AddPost(_ context.Context, post *data.Post) error {
	storage.Mutex.Lock()
	defer storage.Mutex.Unlock()
	post.Id = data.GeneratePostId()
	storage.Posts[post.Id] = post
	posts, found := storage.UserPosts[post.AuthorId]
	if !found {
		posts = make([]*data.Post, 0)
	}
	storage.UserPosts[post.AuthorId] = append([]*data.Post{post}, posts...)
	return nil
}

func (storage *MemoryStorage) GetPost(_ context.Context, postId data.PostId) (*data.Post, error) {
	storage.Mutex.RLock()
	defer storage.Mutex.RUnlock()
	post, found := storage.Posts[postId]
	if !found {
		return nil, errors.New("post not found")
	}
	return post, nil
}

func (storage *MemoryStorage) GetUserPosts(_ context.Context, userId data.UserId,
										   token data.PageToken, limit int) ([]*data.Post, data.PageToken, error) {
	storage.Mutex.RLock()
	defer storage.Mutex.RUnlock()
	posts, found := storage.UserPosts[userId]
	if !found {
		return nil, "", nil
	}
	offset := 0
	if token != "" {
		var err error
		offset, err = strconv.Atoi(string(token))
		if err != nil {
			return nil, "", err
		}
	}
	postsSlice := make([]*data.Post, 0)
	for i := offset; i < offset+limit && i < len(posts); i++ {
		postsSlice = append(postsSlice, posts[i])
	}
	nextToken := ""
	if offset + limit < len(posts) {
		nextToken = strconv.Itoa(offset + limit)
	}
	return postsSlice, data.PageToken(nextToken), nil
}

func (storage *MemoryStorage) UpdatePost(_ context.Context, post *data.Post) error {
	storage.Mutex.Lock()
	defer storage.Mutex.Unlock()
	storage.Posts[post.Id] = post
	for i, userPost := range storage.UserPosts[post.AuthorId] {
		if userPost.Id == post.Id {
			storage.UserPosts[post.AuthorId][i] = post
			break
		}
	}
	return nil
}

