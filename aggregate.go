package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AggregateCursor(coll string, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	c := Coll(coll)
	if c == nil {
		return nil, errCollNil
	}
	return c.Aggregate(ctx, pipeline, opts...)
}

func AggregateAll(coll string, pipeline interface{}, result interface{}, opts ...*options.AggregateOptions) error {
	cur, err := AggregateCursor(coll, pipeline, opts...)
	if err != nil {
		return err
	}
	return cur.All(ctx, result)
}
