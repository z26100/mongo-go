package mongo

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (b Client) GetCollection(database, collection string, databaseOptions *options.DatabaseOptions, collectionOptions *options.CollectionOptions) (*mongo.Collection, error) {
	if b.client == nil {
		return nil, errors.New("Mongo client must not be nil")
	}
	db, err := b.GetDatabase(database, databaseOptions)
	if err != nil {
		return nil, err
	}
	return db.Collection(collection, collectionOptions), nil
}

func (b Client) GetCollections(database string, nameOnly bool) (interface{}, error) {
	if b.client == nil {
		return nil, errors.New("Mongo client must not be nil")
	}
	db, err := b.GetDatabase(database, nil)
	if err != nil {
		return nil, err
	}
	cursor, err := db.ListCollections(Ctx(), bson.M{}, &options.ListCollectionsOptions{NameOnly: proto.Bool(nameOnly)})
	if !cursor.Next(Ctx()) {
		return nil, nil
	}
	var result []interface{}
	err = cursor.All(Ctx(), &result)
	return result, err
}

func (b Client) DropCollection(database string, collection string) error {
	col, err := b.GetCollection(database, collection, nil, nil)
	if err != nil {
		return err
	}
	return col.Drop(Ctx())
}
