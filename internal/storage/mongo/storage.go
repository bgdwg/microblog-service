package mongo

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"log"
	"microblogging-service/internal/data"
	storage2 "microblogging-service/internal/storage"
	"time"
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
			return storage2.ErrCollision
		}
		return fmt.Errorf("%w: insertion error", storage2.ErrBase)
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
			return nil, fmt.Errorf("%w: not found post with id %v ", storage2.ErrNotFound, postId)
		}
		return nil, fmt.Errorf("%w: finding error", storage2.ErrBase)
	}
	post.Id = postId
	return &post, nil
}

func (storage *MongoStorage) GetUserPosts(ctx context.Context, userId data.UserId,
	token data.PageToken, limit int) ([]*data.Post, data.PageToken, error) {
	cursor, err := storage.Posts.Find(ctx, setFilter(userId, token), setOptions(limit))
	if err != nil {
		return nil, "", fmt.Errorf("%w: finding error", storage2.ErrBase)
	}
	posts, nextToken, err := copyPosts(ctx, cursor, limit)
	return posts, nextToken, err
}

func (storage *MongoStorage) UpdatePost(ctx context.Context, post *data.Post) error {
	objectId, err := primitive.ObjectIDFromHex(string(post.Id))
	if err != nil{
		log.Println("Invalid id")
	}
	filter := bson.M{"_id": objectId}
	update := bson.M{
		"$set": bson.M{
			"text": post.Text,
			"lastModifiedAt": post.LastModifiedAt,
		},
	}
	if _, err = storage.Posts.UpdateOne(ctx, filter, update); err != nil {
		return fmt.Errorf("update error - %w", storage2.ErrBase)
	}
	return nil
}

func ensureIndexes(ctx context.Context, collection *mongo.Collection) {
	indexModels := []mongo.IndexModel{
		{
			Keys: bsonx.Doc{
				{Key: "_id", Value: bsonx.Int32(-1)},
				{Key: "authorId", Value: bsonx.Int32(1)},
			},
		},
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := collection.Indexes().CreateMany(ctx, indexModels, opts)
	if err != nil {
		panic(fmt.Errorf("failed to ensure indexes %w", err))
	}
}

func copyPosts(ctx context.Context, cursor *mongo.Cursor, limit int) ([]*data.Post, data.PageToken, error) {
	var posts []*data.Post
	for cursor.Next(ctx) {
		var post data.Post
		if err := cursor.Decode(&post); err != nil {
			return nil, "", fmt.Errorf("%w: decoding error", storage2.ErrBase)
		}
		posts = append(posts, &post)
	}
	nextToken := data.PageToken("")
	if limit == len(posts) - 1 {
		nextToken = data.PageToken(posts[limit].Id)
		posts = posts[:limit]
	}
	return posts, nextToken, nil
}

func setOptions(limit int) *options.FindOptions {
	opt := options.Find()
	opt.SetLimit(int64(limit) + 1)
	opt.SetSort(bson.D{{"_id", -1}})
	return opt
}

func setFilter(userId data.UserId, token data.PageToken) bson.M {
	filter := bson.M{"authorId": userId}
	if token != "" {
		objId, _ := primitive.ObjectIDFromHex(string(token))
		filter["_id"] = bson.M{"$lte": objId}
	}
	return filter
}




