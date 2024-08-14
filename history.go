package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"

	"strings"
	"time"
)

func WriteToHistory(coll string, document BsonType) (BsonType, error) {
	sess, err := Client().StartSession()
	if err != nil {
		return nil, err
	}
	defer sess.EndSession(context.TODO())
	result, err := sess.WithTransaction(context.TODO(), func(ctx mongo.SessionContext) (interface{}, error) {
		c := Coll(coll)
		collHistory := getHistoryCollection(coll)
		res := c.FindOne(ctx, asFilter(document))

		result := BsonType{}
		err = res.Decode(&result)
		if err != nil {
			return nil, err
		}
		result["id"] = result.GetId()
		delete(result, "_id")
		_, err := InsertOne(collHistory, result)
		if err != nil {
			return nil, err
		}
		document.SetLastModified(time.Now())
		replace, err := ReplaceOne(coll, document)
		return replace, err
	})
	if err != nil {
		return nil, err
	}
	return result.(BsonType), err
}
func MoveToHistory(coll string, document BsonType) (*mongo.DeleteResult, error) {
	sess, err := Client().StartSession()
	if err != nil {
		return nil, err
	}
	defer sess.EndSession(context.TODO())
	res, err := sess.WithTransaction(context.TODO(), func(ctx mongo.SessionContext) (interface{}, error) {
		c := Coll(coll)
		collHistory := getHistoryCollection(coll)
		res := c.FindOne(ctx, asFilter(document))
		if err != nil {
			return nil, err
		}
		result := BsonType{}
		err = res.Decode(&result)
		if err != nil {
			return nil, err
		}
		result["id"] = result.GetId()
		delete(result, "_id")
		_, err := InsertOne(collHistory, result)
		if err != nil {
			return nil, err
		}
		return c.DeleteOne(ctx, asFilter(document))
	})
	if err != nil {
		return nil, err
	}
	return res.(*mongo.DeleteResult), err
}
func getHistoryCollection(coll string) string {
	return fmt.Sprintf("%s%s", coll, "HISTORY")
}
func ignoreHistoryCollections(colls []string) []string {
	var results []string
	for _, coll := range colls {
		if strings.HasSuffix(strings.ToUpper(coll), "HISTORY") {
			continue
		}
		results = append(results, coll)
	}
	return results
}
