package v2

import (
	"context"
	"github.com/z26100/log-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UpdatedField struct {
	Key   string `json:"key" bson:"key"`
	Value any    `json:"value" bson:"value"`
}

type UpdateDescription struct {
	UpdatedFields bson.M `bson:"updatedFields" json:"updatedFields"`
}
type ChangeEventDocument struct {
	Id                bson.M            `bson:"_id" json:"_id"`
	OperationType     string            `bson:"operationType" json:"operationType"`
	DocumentKey       BsonType          `bson:"documentKey" json:"documentKey"`
	FullDocument      BsonType          `bson:"fullDocument" json:"fullDocument"`
	UpdateDescription UpdateDescription `bson:"updateDescription" json:"updateDescription"`
}

type WatchChannel chan ChangeEventDocument

func WatchCollection(coll string, pipeline []bson.M, dest WatchChannel, opts ...*options.ChangeStreamOptions) error {
	ctx := context.TODO()
	c := Coll(coll)
	if c == nil {
		return errCollNil
	}
	changeStream, err := c.Watch(ctx, pipeline, opts...)
	if err != nil {
		return err
	}
	// Process change stream events
	defer changeStream.Close(ctx)
	for changeStream.Next(ctx) {
		changeDocument := ChangeEventDocument{}
		if err := changeStream.Decode(&changeDocument); err != nil {
			log.Error(err)
			continue
		} else {
			dest <- changeDocument
		}
	}
	return changeStream.Err()
}
