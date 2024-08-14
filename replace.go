package mongo

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ReplaceOne(coll string, replacement BsonType, opts ...*options.ReplaceOptions) (BsonType, error) {
	c := Coll(coll)
	if c == nil {
		return nil, errCollNil
	}
	if opts == nil {
		opts = append(opts, &options.ReplaceOptions{})
	}
	_, err := c.ReplaceOne(ctx, asFilter(replacement), replacement, opts[0].SetUpsert(true))
	if err != nil {
		return nil, err
	}
	result := c.FindOne(ctx, asFilter(replacement))
	if result == nil {
		return nil, errors.New("not found")
	}
	if result.Err() != nil {
		return nil, result.Err()
	}
	var returnDoc = BsonType{}
	err = result.Decode(&returnDoc)
	return returnDoc, err
}
