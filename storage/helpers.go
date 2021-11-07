package storage

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
			return nil, "", fmt.Errorf("%w: decoding error", ErrBase)
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
