package mongo

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/tidwall/pretty"
	log "github.com/z26100/log-go"
	rest "github.com/z26100/rest-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const (
	MaxUploadSize = 1024 * 1024
	DefaultTag    = "latest"
)

func GetRoutes(mongoClient *Client) []rest.Route {
	routes := []rest.Route{
		{Path: "/{database:[a-z,-]+}/_admin/collections", HandlerFc: getManagedCollections(mongoClient), Methods: "GET"},
		{Path: "/{database:[a-z,-]+}/_admin/clone", HandlerFc: Clone(mongoClient), Methods: "POST"},
		{Path: "/{database:[a-z,-]+}/_admin/backup", HandlerFc: Backup(mongoClient), Methods: "POST"},
		{Path: "/{database:[a-z,-]+}/_admin/filebackup", HandlerFc: FileBackup(mongoClient), Methods: "GET"},
		{Path: "/{database:[a-z,-]+}/_admin/filerestore", HandlerFc: FileRestore(mongoClient), Methods: "POST"},
		{Path: "/{database:[a-z,-]+}/_admin/history", HandlerFc: DeleteHistory(mongoClient), Methods: "DELETE"},
		{Path: "/{database:[a-z,-]+}/{collection:[a-z,_]+}/{document:[a-z,0-9,-]+}", HandlerFc: GetDocument(mongoClient), Methods: "GET"},
		{Path: "/{database:[a-z,-]+}/{collection:[a-z,_]+}/{document:[a-z,0-9,-]+}", HandlerFc: PutDocument(mongoClient), Methods: "POST,PUT"},
		{Path: "/{database:[a-z,-]+}/{collection}/{document:[a-z,0-9,-]+}", HandlerFc: DeleteDocument(mongoClient), Methods: "DELETE"},
		{Path: "/{database:[a-z,-]+}/{collection:[a-z,_]+}", HandlerFc: PutDocument(mongoClient), Methods: "POST,PUT"},
		{Path: "/{database:[a-z,-]+}/{collection:[a-z,_]+}", HandlerFc: GetDocument(mongoClient), Methods: "GET"},
		{Path: "/{database:[a-z,-]+}", HandlerFc: getCollections(mongoClient), Methods: "GET"},
		{Path: "/{database:[a-z,-]+}/{collection:[a-z,_]+}", HandlerFc: deleteCollection(mongoClient), Methods: "DELETE"},
		{Path: "/{database:[a-z,-]+}", HandlerFc: deleteDatabase(mongoClient), Methods: "DELETE"},
		{Path: "/", HandlerFc: getDatabases(mongoClient), Methods: "GET"},
	}
	return routes
}

func getCollections(mongoClient *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		v := r.URL.Query()["nameOnly"]
		database := vars["database"]
		nameOnly := false
		if v != nil && len(v) > 0 {
			nameOnly, _ = strconv.ParseBool(v[0])
		}
		if check(func() bool { return database == "" }, w) {
			return
		}
		data, err := mongoClient.GetCollections(database, nameOnly)
		if data == nil {
			w.Write([]byte("[]"))
			return
		}
		if checkError(err, w) {
			return
		}
		jsonData, err := bson.MarshalExtJSON(bson.M{"body": data}, true, true)
		if checkError(err, w) {
			return
		}
		_, err = w.Write(jsonData)
		if checkError(err, w) {
			return
		}
	}
}

func deleteDatabase(mongoClient *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		database := vars["database"]

		if check(func() bool { return database == "" }, w) {
			return
		}
		err := mongoClient.DropDatabase(database)
		if checkError(err, w) {
			return
		}
	}
}

func deleteCollection(mongoClient *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		database := vars["database"]
		collection := vars["collection"]
		if check(func() bool { return collection == "" || database == "" }, w) {
			return
		}
		err := mongoClient.DropCollection(database, collection)
		if checkError(err, w) {
			return
		}
	}
}

func getDatabases(mongoClient *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query()["nameOnly"]
		nameOnly := false
		if v != nil && len(v) > 0 {
			nameOnly, _ = strconv.ParseBool(v[0])
		}
		data, err := mongoClient.GetDatabases(&options.DatabaseOptions{}, nameOnly)
		if checkError(err, w) {
			return
		}
		if data == nil {
			return
		}
		jsonData, err := bson.MarshalExtJSON(bson.M{"body": data}, true, true)
		if checkError(err, w) {
			return
		}
		_, err = w.Write(jsonData)
		if checkError(err, w) {
			return
		}
	}
}

func GetDocument(mongoClient *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		database := vars["database"]
		collection := vars["collection"]
		document := vars["document"]
		if check(func() bool { return collection == "" || database == "" }, w) {
			return
		}

		filter := bson.M{}
		data := make([]bson.M, 0)
		var err error
		switch document {
		case "search":
			for k, v := range r.URL.Query() {
				if strings.HasPrefix(v[0], "_d") {
					numValue, err := strconv.ParseInt(strings.TrimPrefix(v[0], "_d"), 10, 0)
					if err != nil {
						break
					}
					filter[k] = numValue
				} else {
					filter[k] = v[0]
				}
			}
		default:
			if document != "" {
				filter = bson.M{"_id": document}
			} else {
				filter = bson.M{}
			}
		}
		data, err = mongoClient.FindMany(database, collection, filter)
		if checkError(err, w) {
			return
		}
		if data == nil {
			return
		}
		var jsonData []byte
		tag := r.URL.Query().Get("tag")
		if tag == "" {
			tag = DefaultTag
		}
		if tag != "all" {
			jsonData, err = bson.MarshalExtJSON(renderResponse("body", getDocumentVersion(data, tag), now()), false, true)
		} else {
			jsonData, err = bson.MarshalExtJSON(renderResponse("body", data, now()), false, true)
		}
		if r.URL.Query().Get("pretty") == "true" {
			jsonData = pretty.Pretty(jsonData)
		}
		if checkError(err, w) {
			return
		}
		_, err = w.Write(jsonData)
		if checkError(err, w) {
			return
		}
	}
}

func getDocumentVersion(data []bson.M, version string) []bson.M {
	var result []bson.M
	for _, doc := range data {
		current := doc[version]
		if current != nil {
			result = append(result, current.(bson.M))
		}
	}
	return result
}

func PutDocument(mongoClient *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		database := vars["database"]
		collection := vars["collection"]
		if check(func() bool { return collection == "" || database == "" }, w) {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if checkError(err, w) {
			return
		}
		doc := bson.M{}
		err = bson.UnmarshalExtJSON(body, true, &doc)
		if checkError(err, w) {
			return
		}

		document := vars["document"]
		// overwrite document id in the document if found in the url
		if document == "" && doc["_id"] != nil {
			document = doc["_id"].(string)
		}
		// still no document id, then we generate a new id
		now := now()
		newDocument := false
		if document == "" {
			newDocument = true
			document = uuid.New().String()
			doc["_id"] = document
			doc["created"] = now
		}

		var data bson.M
		filter := bson.M{"_id": document}

		tag := r.URL.Query().Get("tag")
		if tag == "" {
			tag = DefaultTag
		}

		history := r.URL.Query().Get("history") == "true"

		updateOpts := &options.FindOneAndUpdateOptions{
			Upsert:         proto.Bool(true),
			ReturnDocument: ReturnDocument(options.After),
		}
		updatePipeline := bson.A{}
		if history {
			updatePipeline = append(updatePipeline,
				bson.M{"$set": bson.M{
					"history" + tag: bson.M{"$ifNull": bson.A{bson.M{"$concatArrays": bson.A{bson.A{"$" + tag}, "$history." + tag}}, bson.A{}}},
					"version":       bson.M{"$ifNull": bson.A{bson.M{"$add": bson.A{"$version", 1}}, 0}},
				}})
		}
		doc["lastModified"] = now
		updatePipeline = append(updatePipeline, bson.M{"$unset": tag})
		updatePipeline = append(updatePipeline,
			bson.M{"$set": bson.M{
				"_id":          document,
				"lastModified": now,
				tag:            doc,
			},
			})
		if newDocument {
			updatePipeline = append(updatePipeline, bson.M{"$set": bson.M{"created": now}})
		}

		data, err = mongoClient.UpdateOne(database, collection, filter, updatePipeline, updateOpts)
		if checkError(err, w) {
			return
		}
		jsonData, err := bson.MarshalExtJSON(renderResponse("body", data[tag], now), false, true)
		if checkError(err, w) {
			return
		}
		_, err = w.Write(jsonData)
		if checkError(err, w) {
			return
		}
	}
}

func DeleteDocument(mongoClient *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		database := vars["database"]
		collection := vars["collection"]
		id := vars["document"]
		if check(func() bool { return collection == "" || database == "" || id == "" }, w) {
			return
		}

		if r.URL.Query().Get("purge") == "true" {
			err := purgeDocument(mongoClient, database, collection, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		updateOpts := &options.FindOneAndUpdateOptions{
			Upsert:         proto.Bool(true),
			ReturnDocument: ReturnDocument(options.Before),
		}

		tag := r.URL.Query().Get("tag")
		if tag == "" {
			tag = DefaultTag
		}

		now := now()

		updatePipeline := bson.A{
			bson.M{"$set": bson.M{
				"lastModified":   now,
				"history." + tag: bson.M{"$ifNull": bson.A{bson.M{"$concatArrays": bson.A{bson.A{"$" + tag}, "$history." + tag}}, bson.A{}}},
			},
			},
			bson.M{"$unset": tag},
		}

		data, err := mongoClient.UpdateOne(database, collection,
			bson.M{"_id": id},
			updatePipeline,
			updateOpts)
		if checkError(err, w) {
			return
		}

		if data == nil {
			return
		}
		jsonData, err := bson.MarshalExtJSON(renderResponse("body", data[tag], now), true, true)
		if checkError(err, w) {
			return
		}
		_, err = w.Write(jsonData)
		if checkError(err, w) {
			return
		}
	}
}

func purgeDocument(mongoClient *Client, database, collection, documentId string) error {
	if collection == "" || database == "" || documentId == "" {
		return errors.New("bad request")
	}
	filter := bson.M{"_id": documentId}
	return mongoClient.DeleteOne(database, collection, filter)
}

func check(condition func() bool, w http.ResponseWriter) bool {
	if condition() {
		http.Error(w, "BadRequest", http.StatusBadRequest)
		return true
	}
	return false
}

func checkError(err error, w http.ResponseWriter) bool {
	return check(func() bool {
		if err != nil {
			log.Println(err)
		}
		return err != nil
	}, w)
}

func checkDataAndError(data interface{}, err error, w http.ResponseWriter) bool {
	return check(func() bool {
		if err != nil {
			log.Println(err)
		}
		return err != nil || data == nil
	}, w)
}
