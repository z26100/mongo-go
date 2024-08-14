package mongo

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func RecoverOne(coll string, id *primitive.ObjectID) (BsonType, error) {
	hColl := getHistoryCollection(coll)
	res := Coll(hColl).FindOne(ctx, asFilter(BsonType{"_id": *id}))
	if res == nil {
		return nil, errors.New("no result found")
	}
	if res.Err() != nil {
		return nil, res.Err()
	}
	obj := BsonType{}
	err := res.Decode(&obj)
	if err != nil {
		return nil, err
	}
	obj["recoveryId"] = obj["_id"]
	obj["_id"] = obj["id"]
	obj["recoveredAt"] = time.Now()
	delete(obj, "id")
	return InsertOrReplace(coll, obj)
}
