package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"microblogging-service/data"
	"microblogging-service/storage"
	"net/http"
	"strconv"
	"strings"
)

type HTTPHandler struct {
	Storage storage.Storage
}

type CreatePostRequestData struct {
	Text string `json:"text"`
}

type GetUserPostsResponseData struct {
	Posts    []*data.Post   `json:"posts,omitempty"`
	NextPage data.PageToken `json:"nextPage,omitempty"`
}

func (handler *HTTPHandler) HandleCreatePost(rw http.ResponseWriter, r *http.Request) {
	var reqData CreatePostRequestData
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	userId := data.UserId(r.Header.Get("System-Design-User-Id"))
	if userId == "" {
		http.Error(rw, "userId not found", http.StatusUnauthorized)
		return
	}
	post := data.NewPost(reqData.Text, userId)
	if err = handler.Storage.AddPost(r.Context(), post); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	rawResponse, _ := json.Marshal(post)
	rw.Header().Set("Content-Type", "application/json")
	if _, err = rw.Write(rawResponse); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (handler *HTTPHandler) HandleGetPost(rw http.ResponseWriter, r *http.Request) {
	postId := data.PostId(mux.Vars(r)["postId"])
	post, err := handler.Storage.GetPost(r.Context(), postId)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
		return
	}
	rawResponse, _ := json.Marshal(post)
	rw.Header().Set("Content-Type", "application/json")
	if _, err = rw.Write(rawResponse); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (handler *HTTPHandler) HandleGetUserPosts(rw http.ResponseWriter, r *http.Request) {
	var err error
	userId := mux.Vars(r)["userId"]
	if userId == "" {
		http.Error(rw, "userId not found", http.StatusBadRequest)
		return
	}
	query := r.URL.Query()
	token := ""
	if pageToken := query.Get("page"); pageToken != "" {
		if !strings.HasPrefix(pageToken, userId + ":") {
			http.Error(rw, "invalid page parameter", http.StatusBadRequest)
			return
		}
		token = pageToken[len(userId) + 1:]
	}
	limit := 10
	if numPosts := query.Get("size"); numPosts != "" {
		if limit, err = strconv.Atoi(numPosts); err != nil || limit < 1 || limit > 100 {
			http.Error(rw, "invalid size parameter", http.StatusBadRequest)
			return
		}
	}
	posts, nextToken, err := handler.Storage.GetUserPosts(
		r.Context(), data.UserId(userId), data.PageToken(token), limit)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	response := &GetUserPostsResponseData{
		Posts:    posts,
		NextPage: "",
	}
	if nextToken != "" {
		response.NextPage = data.GeneratePageToken(data.UserId(userId), nextToken)
	}
	rawResponse, _ := json.Marshal(response)
	rw.Header().Set("Content-Type", "application/json")
	if _, err = rw.Write(rawResponse); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (handler *HTTPHandler) HandlePing(rw http.ResponseWriter, _ *http.Request) {
	rw.WriteHeader(http.StatusOK)
}
