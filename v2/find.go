package v2

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func FindCursor(coll string, filter bson.M, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	c := Coll(coll)
	if c == nil {
		return nil, errCollNil
	}
	return c.Find(ctx, filter, opts...)
}

func FindAll(coll string, filter bson.M, result interface{}, opts ...*options.FindOptions) error {
	cur, err := FindCursor(coll, filter, opts...)
	if err != nil {
		return err
	}
	if !cur.Next(ctx) {
		result = []bson.M{}
		return nil
	}
	return cur.All(ctx, result)
}

func FindOne(coll string, id interface{}, result interface{}, opts ...*options.FindOneOptions) error {
	c := Coll(coll)
	if c == nil {
		return errCollNil
	}
	filter := bson.M{"_id": id}
	res := c.FindOne(ctx, filter, opts...)
	if res.Err() != nil {
		return res.Err()
	}
	return res.Decode(result)
}
