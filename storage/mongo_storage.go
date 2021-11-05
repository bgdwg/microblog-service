package storage

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
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
	res, err := storage.Posts.InsertOne(ctx, post)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrCollision
		}
		return fmt.Errorf("insertion error - %w", ErrBase)
	}
	post.Id = data.PostId(res.InsertedID.(primitive.ObjectID).Hex())
	return nil
}

func (storage *MongoStorage) GetPost(ctx context.Context, postId data.PostId) (*data.Post, error) {
	var post data.Post
	objectId, err := primitive.ObjectIDFromHex(string(postId))
	if err != nil{
		log.Println("Invalid id")
	}
	if err := storage.Posts.FindOne(ctx, bson.M{"_id": objectId}).Decode(&post); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("not found post with id=%v - %w", postId, ErrNotFound)
		}
		return nil, fmt.Errorf("finding error - %w", ErrBase)
	}
	post.Id = postId
	return &post, nil
}

func (storage *MongoStorage) GetUserPosts(ctx context.Context, userId data.UserId,
										  token data.PageToken, limit int) ([]*data.Post, data.PageToken, error) {
	cursor, err := storage.Posts.Find(ctx, setFilter(userId, token), setOptions(limit))
	if err != nil {
		return nil, "", fmt.Errorf("finding error - %w", ErrBase)
	}
	posts, nextToken, err := copyPosts(ctx, cursor, limit)
	return posts, nextToken, err
}




