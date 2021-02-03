package mongo

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/****
	GetDatabase
 ****/
func (b Client) GetDatabase(database string, opts *options.DatabaseOptions) (*mongo.Database, error) {
	if b.client == nil {
		return nil, errors.New("Mongo client must not be nil")
	}
	return b.client.Database(database, opts), nil
}

func (b Client) GetDatabases(databaseOptions *options.DatabaseOptions, nameonly bool) (interface{}, error) {
	if b.client == nil {
		return nil, errors.New("Mongo client must not be nil")
	}
	res, err := _getDatabases(b.client)
	return res.Databases, err
}

func (b Client) DropDatabase(database string) error {
	db, err := b.GetDatabase(database, nil)
	if err != nil {
		return err
	}
	return db.Drop(Ctx())
}
