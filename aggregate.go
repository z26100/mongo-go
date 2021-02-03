package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func (b Client) Query(database string, collection string, pipeline interface{}, opts *options.AggregateOptions) (*mongo.Cursor, error) {
	log.Printf("Querying %s, %s", database, collection)
	col, err := b.GetCollection(database, collection, nil, nil)
	if err != nil {
		return nil, err
	}
	return col.Aggregate(Ctx(), pipeline, opts)
}
