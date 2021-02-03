package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (b Client) DeleteOne(database string, collection string, filter bson.M) error {
	col, err := b.GetCollection(database, collection, b.config.databaseOptions, b.config.collectionOptions)
	if err != nil {
		return err
	}
	_, err = DeleteOne(col, filter, &options.DeleteOptions{})
	return err
}

func DeleteOne(collection *mongo.Collection, filter bson.M, deleteOptions *options.DeleteOptions) (*mongo.DeleteResult, error) {
	return collection.DeleteOne(Ctx(), filter, deleteOptions)
}

func DeleteMany(collection *mongo.Collection, filter bson.M, deleteOptions *options.DeleteOptions) (*mongo.DeleteResult, error) {
	return collection.DeleteMany(Ctx(), filter, deleteOptions)
}
