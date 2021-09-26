package data

import (
	"github.com/google/uuid"
	"time"
)

type PostId string
type UserId string
type ISOTimestamp string
type PageToken string

func generatePostId() PostId {
	return PostId(uuid.New().String())
}

func generateTimestamp() ISOTimestamp {
	return ISOTimestamp(time.Now().UTC().Format(time.RFC3339))
}

type Post struct {
	Id        PostId       `json:"id"         bson:"id"`
	Text      string       `json:"text"       bson:"text"`
	AuthorId  UserId       `json:"authorId"   bson:"authorId"`
	CreatedAt ISOTimestamp `json:"createdAt"  bson:"createdAt"`
}

func NewPost(text string, userID UserId) *Post {
	return &Post{
		Id:        generatePostId(),
		Text:      text,
		AuthorId:  userID,
		CreatedAt: generateTimestamp(),
	}
}