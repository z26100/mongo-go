package mongo

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
)

func asFilter(document BsonType) bson.M {
	return bson.M{"_id": document.GetId()}
}

func GetCollections(filter ...bson.M) ([]string, error) {
	if filter == nil {
		filter = []bson.M{{}}
	}
	return client.Database(Database).ListCollectionNames(ctx, filter[0])
}

func contains(indexes []mongo.IndexSpecification, name string) bool {
	for _, index := range indexes {
		if index.Name == name {
			return true
		}
	}
	return false
}

func ReadBody(ctx *gin.Context) ([]BsonType, error) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil || body == nil {
		return nil, err
	}
	if body[0] == '[' {
		var data []BsonType
		err = bson.UnmarshalExtJSON(body, false, &data)
		return data, err
	}
	var data BsonType
	err = bson.UnmarshalExtJSON(body, false, &data)
	if data["body"] != nil {
		var body []BsonType
		for _, v := range data["body"].(primitive.A) {
			body = append(body, v.(BsonType))
		}
		return body, nil
	}
	return nil, errors.New("no body found")
}
