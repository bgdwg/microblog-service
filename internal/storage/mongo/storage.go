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
	"microblogging-service/internal/storage"
	"time"
)

type Storage struct {
	Posts *mongo.Collection
}

const collectionName = "postsCollection"

func NewStorage(mongoUrl, mongoDbName string) *Storage {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUrl))
	if err != nil {
		panic(err)
	}
	collection := client.Database(mongoDbName).Collection(collectionName)
	ensureIndexes(ctx, collection)
	return &Storage{Posts: collection}
}

func (s *Storage) AddPost(ctx context.Context, post *data.Post) error {
	res, err := s.Posts.InsertOne(ctx, post)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return storage.ErrCollision
		}
		return fmt.Errorf("%w: insertion error", storage.ErrBase)
	}
	post.Id = data.PostId(res.InsertedID.(primitive.ObjectID).Hex())
	return nil
}

func (s *Storage) GetPost(ctx context.Context, postId data.PostId) (*data.Post, error) {
	var post data.Post
	objectId, err := primitive.ObjectIDFromHex(string(postId))
	if err != nil{
		log.Println("Invalid id")
	}
	if err := s.Posts.FindOne(ctx, bson.M{"_id": objectId}).Decode(&post); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("%w: not found post with id %v ", storage.ErrNotFound, postId)
		}
		return nil, fmt.Errorf("%w: finding error", storage.ErrBase)
	}
	post.Id = postId
	return &post, nil
}

func (s *Storage) GetUserPosts(ctx context.Context, userId data.UserId,
	token data.PageToken, limit int) ([]*data.Post, data.PageToken, error) {
	cursor, err := s.Posts.Find(ctx, setFilter(userId, token), setOptions(limit))
	if err != nil {
		return nil, "", fmt.Errorf("%w: finding error", storage.ErrBase)
	}
	return copyPosts(ctx, cursor, limit)
}

func (s *Storage) UpdatePost(ctx context.Context, post *data.Post) error {
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
	if _, err = s.Posts.UpdateOne(ctx, filter, update); err != nil {
		return fmt.Errorf("update error - %w", storage.ErrBase)
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
			return nil, "", fmt.Errorf("%w: decoding error", storage.ErrBase)
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




