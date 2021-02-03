package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tidwall/pretty"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const version = "v2"

type BackupCollection struct {
	Name    string        `json:"name"`
	Records []interface{} `json:"records"`
}
type BackupResponse struct {
	Datestamp   time.Time          `json:"datestamp"`
	Collections []BackupCollection `json:"collections"`
}

var managedCollections = bson.A{"tags", "trends", "contacts", "organisations", "portfolios", "offerings", "radars", "projects", "assets", "verticals", "documents"}

func getManagedCollections(mongoClient *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestTime := now()
		vars := mux.Vars(r)
		database := vars["database"]
		filter := bson.M{"name": bson.M{"$in": managedCollections}}
		collections, err := mongoClient.client.Database(database).ListCollectionNames(context.Background(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp, err := json.Marshal(renderResponse("body", collections, requestTime))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, _ = w.Write(resp)
	}
}

func Clone(mongoClient *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Receiving clone request")
		vars := mux.Vars(r)
		database := vars["database"]
		commit := r.URL.Query().Get("commit")

		cols := managedCollections
		body, _ := readBodyWithBsonDecoder(r.Body)

		sourceTag := DefaultTag
		targetTag := ""

		if body != nil {
			if body["collections"] != nil {
				cols = body["collections"].(bson.A)
			}
			if body["sourceTag"] != nil {
				sourceTag = body["sourceTag"].(string)
			}
			if body["targetTag"] == nil {
				http.Error(w, "Target tag must not be nil", http.StatusBadRequest)
				return
			}
			targetTag = body["targetTag"].(string)
		} else {
			http.Error(w, "Body must not be nil", http.StatusBadRequest)
			return
		}

		filter := bson.M{"name": bson.M{"$in": cols}}

		collections, err := mongoClient.client.Database(database).ListCollectionNames(context.Background(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		results := renderResponse("body", bson.A{}, now())

		for _, collection := range collections {
			pipeline := bson.A{
				bson.M{"$set": bson.M{targetTag: "$" + sourceTag}},
				bson.M{"$set": bson.M{"history." + targetTag: "$history." + sourceTag}},
			}
			out := collection
			if commit != "" {
				pipeline = append(pipeline, bson.M{"$out": out})
			}

			col, err := mongoClient.GetCollection(database, collection, nil, nil)
			if err != nil {
				log.Println(err)
				break
			}
			cursor, err := col.Aggregate(Ctx(), pipeline)
			if err != nil {
				log.Println(err)
				break
			}
			var records bson.A
			err = cursor.All(Ctx(), &records)
			if err != nil {
				log.Println(err)
				break
			}
			results["body"] = append(results["body"].(bson.A), bson.M{"name": collection, "records": records})
		}
		jsonDoc, err := bson.MarshalExtJSON(results, false, false)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		_, _ = w.Write(pretty.Pretty(jsonDoc))
	}
}

func Backup(mongoClient *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Receiving backup request")
		vars := mux.Vars(r)
		database := vars["database"]
		tag := r.URL.Query().Get("tag")
		if tag == "" {
			tag = fmt.Sprintf("%d", time.Now().Unix())
		}

		managedColls := managedCollections
		body, _ := readBodyWithBsonDecoder(r.Body)
		if body != nil && body["collections"] != nil {
			managedColls = body["collections"].(bson.A)
		}
		filter := bson.M{"name": bson.M{"$in": managedColls}}

		collections, err := mongoClient.client.Database(database).ListCollectionNames(context.Background(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var response bson.A
		for _, collection := range collections {
			pipeline := bson.A{
				bson.M{"$match": bson.M{}},
				bson.M{"$addFields": bson.M{tag: "$current"}},
				bson.M{"$out": collection},
			}
			col, err := mongoClient.GetCollection(database, collection, nil, nil)
			if err != nil {
				log.Println(err)
				break
			}
			cursor, err := col.Aggregate(Ctx(), pipeline)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			var resp bson.A
			err = cursor.All(Ctx(), &resp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			response = append(response, resp...)
		}
		_ = Write(w, bson.M{"response": response})
	}
}

func Restore(mongoClient *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func FileBackup(mongoClient *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Receiving backup request")
		vars := mux.Vars(r)
		database := vars["database"]
		filter := bson.M{}
		collections, err := mongoClient.client.Database(database).ListCollectionNames(context.Background(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result := BackupResponse{
			Datestamp:   time.Now(),
			Collections: make([]BackupCollection, 0),
		}
		for _, collection := range collections {
			c := BackupCollection{
				Name:    collection,
				Records: nil,
			}
			data, err := mongoClient.FindAll(database, collection)
			if err == nil {
				for _, d := range data {
					jsonData, err := bson.MarshalExtJSON(d, false, false)
					if checkError(err, w) {
						break
					}
					var object interface{}
					err = json.Unmarshal(jsonData, &object)
					if checkError(err, w) {
						break
					}
					c.Records = append(c.Records, object)
				}
			}
			result.Collections = append(result.Collections, c)
		}
		jsonDoc, err := json.Marshal(result)
		if checkError(err, w) {
			return
		}
		w.Header().Set("Content-Disposition", "attachment; filename="+database+"_"+result.Datestamp.Format(time.RFC3339)+".json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonDoc)
	}
}

func DeleteHistory(mongoClient *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Receiving delete history request")
		vars := mux.Vars(r)
		database := vars["database"]
		filter := bson.M{}
		collections, err := mongoClient.client.Database(database).ListCollectionNames(context.Background(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for _, collection := range collections {
			col, err := mongoClient.GetCollection(database, collection, nil, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			pipeline := bson.A{
				bson.M{"$unset": bson.A{"history", "current", "default.casestudies", "default.contacts",
					"default.prospects"}},
			}
			_, err = col.Aggregate(Ctx(), pipeline, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

	}
}

func FileRestore(mongoClient *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received restore request")
		vars := mux.Vars(r)
		database := vars["database"]

		if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
			http.Error(w, "The uploaded file is too big. Please choose an file that's less than 1MB in size", http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer func() { _ = file.Close() }()
		data, err := ioutil.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var bsonData bson.M
		err = bson.UnmarshalExtJSON(data, false, &bsonData)
		for _, item := range bsonData["collections"].(bson.A) {
			col := item.(bson.M)
			collectionName := col["name"].(string)
			log.Printf("deleting %s", collectionName)
			err = mongoClient.DropCollection(database, collectionName)
			if err != nil {
				log.Println(err)
			}
			log.Printf("restoring %s", collectionName)
			for _, record := range col["records"].(bson.A) {
				id := bson.M{"_id": record.(bson.M)["_id"].(string)}
				log.Printf("Restoring %s", id)
				_, err := mongoClient.InsertOrReplace(database, collectionName, id, record)
				if checkError(err, w) {
					return
				}
			}

		}

	}
}
