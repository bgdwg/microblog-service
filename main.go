package main

import (
	"github.com/gorilla/mux"
	"log"
	"microblogging-service/handlers"
	"microblogging-service/storage"
	"net/http"
	"os"
	"time"
)

func NewServer() *http.Server {
	handler := &handlers.HTTPHandler{Storage: storage.NewStorage()}
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/posts", handler.HandleCreatePost).Methods("POST")
	router.HandleFunc("/api/v1/posts/{postId:[A-Za-z0-9_\\-]+}", handler.HandleGetPost).Methods("GET")
	router.HandleFunc("/api/v1/users/{userId:[0-9a-f]+}/posts", handler.HandleGetUserPosts).Methods("GET")
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	return &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:" + port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
}

func main() {
	server := NewServer()
	log.Printf("Start serving on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
