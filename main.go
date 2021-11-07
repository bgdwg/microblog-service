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

type Config struct {
	serverPort 	string
	storageMode string
	mongoUrl 	string
	mongoDbName string
	redisUrl 	string
}

func SetConfig() *Config {
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	storageMode := os.Getenv("STORAGE_MODE")
	if storageMode == "" {
		storageMode = "inmemory"
	}

	return &Config{
		serverPort: 	serverPort,
		storageMode: 	storageMode,
		mongoUrl: 		os.Getenv("MONGO_URL"),
		mongoDbName: 	os.Getenv("MONGO_DBNAME"),
		redisUrl: 		os.Getenv("REDIS_URL"),
	}
}

func NewServer() *http.Server {
	config := SetConfig()

	var handler handlers.HTTPHandler
	if config.storageMode == "inmemory" {
		handler.Storage = storage.NewMemoryStorage()
	} else {
		mongoStorage := storage.NewMongoStorage(config.mongoUrl, config.mongoDbName)
		if config.storageMode == "mongo" {
			handler.Storage = mongoStorage
		} else {
			handler.Storage = storage.NewCacheStorage(mongoStorage, config.redisUrl)
		}
	}

	router := mux.NewRouter()

	router.HandleFunc("/api/v1/posts",
		handler.HandleCreatePost).Methods("POST")

	router.HandleFunc("/api/v1/posts/{postId:[A-Za-z0-9_\\-]+}",
		handler.HandleGetPost).Methods("GET")
	router.HandleFunc("/api/v1/posts/{postId:[A-Za-z0-9_\\-]+}",
		handler.HandleUpdatePost).Methods("PATCH")

	router.HandleFunc("/api/v1/users/{userId:[0-9a-f]+}/posts",
		handler.HandleGetUserPosts).Methods("GET")

	router.HandleFunc("/maintenance/ping",
		handler.HandlePing).Methods("GET")

	return &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:" + config.serverPort,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
}

func main() {
	server := NewServer()
	log.Printf("Start serving on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
