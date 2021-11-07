package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"microblogging-service/internal/handlers"
	"microblogging-service/internal/storage/inmemory"
	"microblogging-service/internal/storage/mongo"
	"microblogging-service/internal/storage/redis"
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

const (
	defaultTimeout = 15 * time.Second
	defaultServerPort = "8080"
	defaultStorageMode = "inmemory"
)

func SetConfig() *Config {
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = defaultServerPort
	}
	storageMode := os.Getenv("STORAGE_MODE")
	if storageMode == "" {
		storageMode = defaultStorageMode
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
		handler.Storage = inmemory.NewStorage()
	} else {
		mongoStorage := mongo.NewStorage(config.mongoUrl, config.mongoDbName)
		if config.storageMode == "mongo" {
			handler.Storage = mongoStorage
		} else {
			handler.Storage = redis.NewCacheStorage(mongoStorage, config.redisUrl)
		}
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/posts", handler.HandleCreatePost).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/posts/{postId:[A-Za-z0-9_\\-]+}", handler.HandleGetPost).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/posts/{postId:[A-Za-z0-9_\\-]+}", handler.HandleUpdatePost).Methods(http.MethodPatch)
	r.HandleFunc("/api/v1/users/{userId:[0-9a-f]+}/posts", handler.HandleGetUserPosts).Methods(http.MethodGet)
	r.HandleFunc("/maintenance/ping", func (rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	return &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("0.0.0.0:%s", config.serverPort),
		ReadTimeout:  defaultTimeout,
		WriteTimeout: defaultTimeout,
	}
}

func main() {
	server := NewServer()
	log.Printf("Start serving on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
