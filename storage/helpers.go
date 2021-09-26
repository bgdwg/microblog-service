package storage

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"microblogging-service/data"
	"time"
)

func ensureIndexes(ctx context.Context, collection *mongo.Collection) {
	indexModels := []mongo.IndexModel{
		{
			Keys: bsonx.Doc{
				{Key: "id", Value: bsonx.Int32(1)},
			},
		},
		{
			Keys: bsonx.Doc{
				{Key: "authorId", Value: bsonx.Int32(1)},
				{Key: "_id", Value: bsonx.Int32(1)},
			},
		},
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := collection.Indexes().CreateMany(ctx, indexModels, opts)
	if err != nil {
		panic(fmt.Errorf("failed to ensure indexes %w", err))
	}
}

func copyPosts(ctx context.Context, cursor *mongo.Cursor) ([]*data.Post, error) {
	var posts []*data.Post
	for cursor.Next(ctx) {
		var post data.Post
		if err := cursor.Decode(&post); err != nil {
			return nil, fmt.Errorf("decoding error - %w", ErrBase)
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func setOptions(offset, limit int) *options.FindOptions {
	opt := options.Find()
	opt.SetSkip(int64(offset))
	opt.SetLimit(int64(limit))
	opt.SetSort(bson.D{{"createdAt", -1}})
	return opt
}
