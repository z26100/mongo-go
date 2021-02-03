package mongo

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"
)

type ResponseHeader struct {
	RequestTimestamp *time.Time `json:"request,omitempty" bson:"request,omitempty"`
	Duration         int64      `json:"duration,omitempty"`
	Version          string     `json:"version"`
}

type ResponseType map[string]interface{}

func renderResponse(attributeName string, body interface{}, requestTime *time.Time) ResponseType {
	now := now()
	duration := int64(0)
	if requestTime != nil {
		duration = now.Sub(*requestTime).Microseconds()
	}
	resp := make(ResponseType)
	resp["header"] = ResponseHeader{
		RequestTimestamp: requestTime,
		Duration:         duration,
		Version:          version,
	}
	if body != nil {
		switch reflect.TypeOf(body).String() {
		case "[]primitive.M":
			if len(body.([]primitive.M)) > 0 {
				resp[attributeName] = body
			}
		default:
			resp[attributeName] = body
		}
	}
	return resp
}

func now() *time.Time {
	t := time.Now()
	return &t
}

func readBodyWithJsonDecoder(in io.Reader) (interface{}, error) {
	body, err := ioutil.ReadAll(in)
	var doc interface{}
	if err != nil {
		return doc, err
	}
	err = json.Unmarshal(body, &doc)
	return doc, err
}

func ReturnDocument(val options.ReturnDocument) *options.ReturnDocument {
	return &val
}

func readBodyWithBsonDecoder(in io.Reader) (bson.M, error) {
	body, err := ioutil.ReadAll(in)
	var doc bson.M
	if err != nil {
		return doc, err
	}
	err = bson.UnmarshalExtJSON(body, false, &doc)
	return doc, err
}

func Write(w http.ResponseWriter, data interface{}) error {
	jsonData, err := bson.MarshalExtJSON(data, false, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	_, err = w.Write(jsonData)
	return err
}
