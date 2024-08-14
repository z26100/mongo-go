package mongo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func DeleteOne(coll string, id interface{}) (*mongo.DeleteResult, error) {
	if id == nil {
		return nil, errors.New("id must not be nil")
	}
	_id, ok := id.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("invalid id")
	}
	return deleteOne(coll, _id)
}

func deleteOne(coll string, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	c := Coll(coll)
	if c == nil {
		return nil, errCollNil
	}
	if !history {
		return c.DeleteOne(ctx, bson.M{"_id": id})
	}
	return MoveToHistory(coll, BsonType{"_id": id})
}

func DeleteManyRecursively(coll string, id interface{}, filter bson.M) error {

	sess, err := Client().StartSession()
	if err != nil {
		return err
	}
	defer sess.EndSession(context.TODO())
	ctx := context.TODO()
	c := Coll(coll)
	_, err = sess.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		pipeline := []bson.M{{"$match": filter}}
		res, err := c.Aggregate(ctx, pipeline)
		if err != nil {
			return res, err
		}
		for res.Next(ctx) == true {
			var data BsonType
			err := res.Decode(&data)
			if err != nil {
				return nil, err
			}
			if data.GetId() != nil {
				err = deleteNode(ctx, c, data.GetId())
				if err != nil {
					return nil, err
				}
			}

		}
		_, err = c.DeleteOne(ctx, bson.M{"_id": id})
		return nil, err
	})
	return err
}

func deleteNode(ctx context.Context, c *mongo.Collection, oid *primitive.ObjectID) error {
	pipeline := []bson.M{{"$match": bson.M{"parent": oid}}}
	res, err := c.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	for res.Next(ctx) {
		var data BsonType
		err := res.Decode(&data)
		if err != nil {
			return err
		}
		err = deleteNode(ctx, c, data.GetId())
		if err != nil {
			return err
		}
	}
	_, err = c.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}
