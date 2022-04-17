package config

import (
	"github.com/elastic/go-elasticsearch/v8"
)

func GetESClient() (*elasticsearch.Client) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		panic(err)
	}
	return es
}