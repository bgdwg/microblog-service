package data

import (
	"github.com/google/uuid"
	"time"
)

type PostId string
type UserId string
type ISOTimestamp string
type PageToken string

func GeneratePostId() PostId {
	return PostId(uuid.New().String())
}

func GenerateTimestamp() ISOTimestamp {
	return ISOTimestamp(time.Now().UTC().Format(time.RFC3339))
}

func GeneratePageToken(userId UserId, token PageToken) PageToken {
	return PageToken(string(userId) + ":" + string(token))
}

type Post struct {
	Id             PostId       `json:"id"             bson:"_id,omitempty"`
	Text           string       `json:"text"           bson:"text"`
	AuthorId       UserId       `json:"authorId"       bson:"authorId"`
	CreatedAt      ISOTimestamp `json:"createdAt"      bson:"createdAt"`
	LastModifiedAt ISOTimestamp `json:"lastModifiedAt" bson:"lastModifiedAt"`
}

func NewPost(text string, userID UserId) *Post {
	now := GenerateTimestamp()
	return &Post{
		Text:      		text,
		AuthorId:  		userID,
		CreatedAt: 		now,
		LastModifiedAt: now,
	}
}