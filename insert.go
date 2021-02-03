package mongo

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func returnDocument(document options.ReturnDocument) *options.ReturnDocument {
	return &document
}

func (b Client) InsertOrReplace(database, collection string, filter bson.M, update interface{}) (bson.M, error) {
	opts := &options.FindOneAndReplaceOptions{
		Upsert:         aws.Bool(true),
		ReturnDocument: returnDocument(options.After),
	}
	col, err := b.GetCollection(database, collection, b.config.databaseOptions, b.config.collectionOptions)
	if err != nil {
		return nil, err
	}
	result := col.FindOneAndReplace(Ctx(), filter, update, opts)
	if result.Err() != nil {
		return nil, result.Err()
	}
	var resp bson.M
	err = result.Decode(&resp)
	return resp, err
}

func (b Client) InsertOne(database string, collection string, doc bson.M) (bson.M, error) {
	col, err := b.GetCollection(database, collection, b.config.databaseOptions, b.config.collectionOptions)
	if err != nil {
		return nil, err
	}
	res, err := col.InsertOne(Ctx(), doc)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errors.New("result must not be nil")
	}
	id := res.InsertedID
	doc, err = FindOne(col, bson.M{documentIDField: id})
	return doc, err
}

func (b Client) ReplaceOne(database string, collection string, filter bson.M, replacement bson.A, opts ...*options.FindOneAndReplaceOptions) (bson.M, error) {
	col, err := b.GetCollection(database, collection, b.config.databaseOptions, b.config.collectionOptions)
	if err != nil {
		return nil, err
	}
	result, err := FindOneAndReplace(col, filter, replacement, opts...)
	return result, err
}

func (b Client) UpdateOne(database string, collection string, filter bson.M, update bson.A, opts *options.FindOneAndUpdateOptions) (bson.M, error) {
	col, err := b.GetCollection(database, collection, b.config.databaseOptions, b.config.collectionOptions)
	if err != nil {
		return nil, err
	}
	result, err := FindOneAndUpdate(col, filter, update, opts)
	return result, err
}
