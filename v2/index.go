package v2

import (
	"errors"
	"github.com/z26100/log-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetIndex(coll string) ([]mongo.IndexSpecification, error) {
	col := Coll(coll)
	if col == nil {
		return nil, errors.New("no collection found")
	}
	cursor, err := col.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	var results []mongo.IndexSpecification
	err = cursor.All(ctx, &results)
	return results, err
}
func DropIndex(coll string, name string) error {
	col := Coll(coll)
	if col == nil {
		return errors.New("no collection found")
	}
	_, err := col.Indexes().DropOne(ctx, name)
	return err
}

func CreateIndex(coll string, field string) error {
	log.Infof("Create index for colleciton %s on field %s", coll, field)
	return nil

}
func CreateFulltextIndex(coll string) error {
	log.Infof("Create full text index for coll %s", coll)
	col := Coll(coll)
	if col == nil {
		return errors.New("no collection found")
	}
	opts := options.IndexOptions{}
	opts.SetName("TextIndex")
	col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"$**": "text"},
		Options: &opts,
	})
	return nil
}

func EnsureFulltextIndexes(ignoreHistory bool, filter ...bson.M) error {
	colls, err := GetCollections(filter...)
	if err != nil {
		return err
	}
	if ignoreHistory {
		colls = ignoreHistoryCollections(colls)
	}
	for _, coll := range colls {
		indexes, err := GetIndex(coll)
		if err != nil {
			log.Error(err)
			continue
		}
		if !contains(indexes, "TextIndex") {
			CreateFulltextIndex(coll)
		}
	}
	return nil
}
