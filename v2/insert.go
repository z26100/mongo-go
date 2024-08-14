package v2

import (
	"github.com/z26100/log-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"time"
)

func InsertOne(coll string, document BsonType, opts ...*options.InsertOneOptions) (BsonType, error) {
	c := Coll(coll)
	if c == nil {
		return nil, errCollNil
	}
	document.SetCreated(time.Now())

	res, err := c.InsertOne(ctx, document, opts...)
	if err != nil {
		return nil, err
	}
	var result BsonType
	err = FindOne(coll, res.InsertedID, &result)
	return result, err
}

const (
	history = false
)

func InsertOrReplace(coll string, document BsonType, opts ...*options.InsertOneOptions) (BsonType, error) {

	c := Coll(coll)
	if c == nil {
		return nil, errCollNil
	}
	if document.GetCreated() == nil {
		document.SetCreated(time.Now())
	}

	if document.GetId() == nil {
		id := primitive.NewObjectID()
		document.SetId(&id)
	}
	if !history {
		o := &options.UpdateOptions{}
		o.SetUpsert(true)
		c.UpdateOne(ctx, asFilter(document), bson.M{"$set": document}, o)
	}
	found := c.FindOne(ctx, asFilter(document))
	switch found.Err() {
	case nil:
		return WriteToHistory(coll, document)
	default:
		return InsertOne(coll, document, opts...)
	}
}

func InsertMany(coll string, documents []BsonType, opts ...*options.InsertOneOptions) ([]any, error) {
	c := Coll(coll)
	if c == nil {
		return nil, errCollNil
	}
	var results []interface{}
	for _, doc := range documents {
		var err error
		res, err := InsertOrReplace(coll, doc, opts...)
		if err == nil {
			results = append(results, res)
		} else {
			log.Error(err)
		}
	}
	return results, nil
}
