package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"microblogging-service/data"
	"microblogging-service/storage"
	"net/http"
	"strconv"
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
	query := r.URL.Query()
	offset := 0
	if pageToken := query.Get("page"); pageToken != "" {
		if offset, err = strconv.Atoi(pageToken); err != nil {
			http.Error(rw, "invalid page parameter", http.StatusBadRequest)
			return
		}
	}
	limit := 10
	if numPosts := query.Get("size"); numPosts != "" {
		if limit, err = strconv.Atoi(numPosts); err != nil || limit < 1 || limit > 10 {
			http.Error(rw, "invalid size parameter", http.StatusBadRequest)
			return
		}
	}
	userId := data.UserId(mux.Vars(r)["userId"])
	posts, size, err := handler.Storage.GetUserPosts(r.Context(), userId, offset, limit)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	response := &GetUserPostsResponseData{
		Posts:    posts,
		NextPage: "",
	}
	if size > offset + limit {
		response.NextPage = data.PageToken(strconv.Itoa(offset + limit))
	}
	rawResponse, _ := json.Marshal(response)
	rw.Header().Set("Content-Type", "application/json")
	if _, err = rw.Write(rawResponse); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (handler *HTTPHandler) HandlePing(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)
	return
}
