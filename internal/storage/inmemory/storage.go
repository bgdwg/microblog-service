package inmemory

import (
	"context"
	"errors"
	"microblogging-service/internal/data"
	"strconv"
	"sync"
)

type Storage struct {
	Posts       map[data.PostId]*data.Post
	UserPosts   map[data.UserId][]*data.Post
	Mutex       *sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		Posts:     make(map[data.PostId]*data.Post),
		UserPosts: make(map[data.UserId][]*data.Post),
		Mutex:     new(sync.RWMutex),
	}
}

func (s *Storage) AddPost(_ context.Context, post *data.Post) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	post.Id = data.GeneratePostId()
	s.Posts[post.Id] = post
	posts, found := s.UserPosts[post.AuthorId]
	if !found {
		posts = make([]*data.Post, 0)
	}
	s.UserPosts[post.AuthorId] = append([]*data.Post{post}, posts...)
	return nil
}

func (s *Storage) GetPost(_ context.Context, postId data.PostId) (*data.Post, error) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	post, found := s.Posts[postId]
	if !found {
		return nil, errors.New("post not found")
	}
	return post, nil
}

func (s *Storage) GetUserPosts(_ context.Context, userId data.UserId,
										   token data.PageToken, limit int) ([]*data.Post, data.PageToken, error) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	posts, found := s.UserPosts[userId]
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

func (s *Storage) UpdatePost(_ context.Context, post *data.Post) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Posts[post.Id] = post
	for i, userPost := range s.UserPosts[post.AuthorId] {
		if userPost.Id == post.Id {
			s.UserPosts[post.AuthorId][i] = post
			break
		}
	}
	return nil
}

