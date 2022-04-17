package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"errors"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func SearchQuery(query_string *string, index string, es *elasticsearch.Client)(*esapi.Response, error){
	var b strings.Builder
	b.WriteString(*query_string)
	reading := strings.NewReader(b.String())
	
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(index),
		es.Search.WithBody(reading),
	)
	if err != nil {
		return nil, err
	}
	if res.IsError(){
		return nil, errors.New(res.Status())
	}
	return res, nil
}

func Upsert(doc any, index string, id string, es *elasticsearch.Client) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(doc)
	if err != nil {
		log.Println("upsert1:", err)
		return err
	}
	var request = esapi.IndexRequest{Index:index, Body:&buf}
	if id != "" {
		request.DocumentID = id
	}
	resp, err := request.Do(context.Background(), es)
	if err != nil{
		log.Println("upsert2:", err)
		return err
	}
	defer resp.Body.Close()
	if resp.IsError(){
		log.Println("upsert3:", resp.StatusCode)
		return errors.New("Error Occurred with following Status code" + (string)(resp.StatusCode))
	}
	return nil
}