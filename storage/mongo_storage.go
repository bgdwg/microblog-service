package storage

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"microblogging-service/data"
)

type MongoStorage struct {
	Posts *mongo.Collection
}

const collName = "postsCollection"

func NewMongoStorage(mongoUrl, dbName string) *MongoStorage {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUrl))
	if err != nil {
		panic(err)
	}
	collection := client.Database(dbName).Collection(collName)
	ensureIndexes(ctx, collection)
	return &MongoStorage{
		Posts: collection,
	}
}

func (storage *MongoStorage) AddPost(ctx context.Context, post *data.Post) error {
	_, err := storage.Posts.InsertOne(ctx, post)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrCollision
		}
		return fmt.Errorf("insertion error - %w", ErrBase)
	}
	return nil
}

func (storage *MongoStorage) GetPost(ctx context.Context, postId data.PostId) (*data.Post, error) {
	var post data.Post
	if err := storage.Posts.FindOne(ctx, bson.M{"id": postId}).Decode(&post); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("not found post with id=%v - %w", postId, ErrNotFound)
		}
		return nil, fmt.Errorf("finding error - %w", ErrBase)
	}
	return &post, nil
}

func (storage *MongoStorage) GetUserPosts(ctx context.Context, userId data.UserId,
										  offset, limit int) ([]*data.Post, int, error) {
	numPosts, err := storage.Posts.CountDocuments(ctx, bson.M{"authorId": userId})
	if err != nil {
		return nil, 0, fmt.Errorf("count documents error - %w", ErrBase)
	}
	cursor, err := storage.Posts.Find(ctx, bson.M{"authorId": userId}, setOptions(offset, limit))
	if err != nil {
		return nil, 0, fmt.Errorf("finding error - %w", ErrBase)
	}
	posts, err := copyPosts(ctx, cursor)
	return posts, int(numPosts), err
}




