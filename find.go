package mongo

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (b Client) FindOne(database string, collection string, filter bson.M) (bson.M, error) {
	col, err := b.GetCollection(database, collection, b.config.databaseOptions, b.config.collectionOptions)
	if err != nil {
		return nil, err
	}

	return FindOne(col, filter)
}

func (b Client) FindMany(database string, collection string, filter bson.M) ([]bson.M, error) {
	col, err := b.GetCollection(database, collection, b.config.databaseOptions, b.config.collectionOptions)
	if err != nil {
		return nil, err
	}
	cursor, err := FindMany(col, filter)
	if err != nil {
		return nil, err
	}
	result := make([]bson.M, 0)
	if cursor.Next(Ctx()) {
		err = cursor.All(Ctx(), &result)
		return result, err
	}
	return nil, nil
}
func (b Client) FindAll(database string, collection string) ([]bson.M, error) {
	col, err := b.GetCollection(database, collection, b.config.databaseOptions, b.config.collectionOptions)
	if err != nil {
		return nil, err
	}
	cursor, err := FindAll(col)
	if err != nil {
		return nil, err
	}
	result := make([]bson.M, 0)
	if cursor.Next(Ctx()) {
		err = cursor.All(Ctx(), &result)
		return result, err
	}
	return nil, nil
}

func FindAll(collection *mongo.Collection) (*mongo.Cursor, error) {
	return FindMany(collection, bson.M{})
}

func FindMany(collection *mongo.Collection, filter bson.M) (*mongo.Cursor, error) {
	return collection.Find(Ctx(), filter, &options.FindOptions{})
}
func FindOne(collection *mongo.Collection, filter bson.M) (bson.M, error) {
	res := collection.FindOne(Ctx(), filter, &options.FindOneOptions{})
	if res == nil {
		return nil, errors.New("Result must not be nil")
	}
	if res.Err() != nil {
		return nil, res.Err()
	}
	var obj bson.M
	err := res.Decode(&obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func FindOneAndDelete(collection *mongo.Collection, filter bson.M, deleteOptions *options.FindOneAndDeleteOptions) *mongo.SingleResult {
	return collection.FindOneAndDelete(Ctx(), filter, deleteOptions)
}

func FindOneAndReplace(collection *mongo.Collection, filter bson.M, replacement interface{}, options ...*options.FindOneAndReplaceOptions) (bson.M, error) {
	res := collection.FindOneAndReplace(Ctx(), filter, replacement, options...)
	if res == nil {
		return nil, errors.New("Result must not be nil")
	}
	if res.Err() != nil {
		return nil, res.Err()
	}
	var result bson.M
	err := res.Decode(&result)
	return result, err
}

func FindOneAndUpdate(collection *mongo.Collection, filter bson.M, update bson.A, opts *options.FindOneAndUpdateOptions) (bson.M, error) {
	res := collection.FindOneAndUpdate(Ctx(), filter, update, opts)
	if res == nil {
		return nil, errors.New("Result must not be nil")
	}
	var result bson.M
	err := res.Decode(&result)
	return result, err
}
