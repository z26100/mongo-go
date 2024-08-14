package v2

import (
	"context"
	"errors"
	"fmt"
	"github.com/z26100/log-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Hostname = "localhost"
	Port     = 27017
	User     = ""
	Password = ""
	Database = ""

	ctx    = context.TODO()
	client *mongo.Client
)

func Connect() error {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?retryWrites=true&w=majority&ssl=false&authSource=admin", User, Password, Hostname, Port, Database)
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}
	err = Ping()
	if err == nil {
		log.Infof("Connected to host %s database %s.\n", Hostname, Database)
	}
	return err
}
func Client() *mongo.Client {
	return client
}

func Ping() error {
	if client == nil {
		return errors.New("client must not be nil")
	}
	return client.Ping(ctx, nil)
}

func Close() error {
	if client == nil {
		return nil
	}
	return client.Disconnect(context.TODO())
}

func Coll(coll string) *mongo.Collection {
	if client == nil {
		return nil
	}
	return client.Database(Database).Collection(coll)
}
