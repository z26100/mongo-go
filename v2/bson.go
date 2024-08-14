package v2

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type BsonType bson.M

func (b BsonType) GetId() *primitive.ObjectID {
	id, ok := b["_id"].(primitive.ObjectID)
	if ok == true {
		return &id
	}
	return nil
}
func (b BsonType) SetId(id *primitive.ObjectID) {
	b["_id"] = id
}

func (b BsonType) GetCreated() *time.Time {
	t, ok := b["created"].(primitive.DateTime)
	if ok == true {
		timestamp := t.Time()
		return &timestamp
	}
	return nil
}
func (b BsonType) SetCreated(val time.Time) {
	b["created"] = &val
}
func (b BsonType) SetLastModified(val time.Time) {
	b["lastModified"] = val
}
