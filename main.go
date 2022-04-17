package main

import (
	"log"
	"net/http"
	"github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/config"
	"github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/controller"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gorilla/mux"
)

type ConfigVars struct{
	es *elasticsearch.Client
	router *mux.Router
}

var configVars ConfigVars

func main() {
	configVars.es = config.GetESClient()
	configVars.router = controller.CreateRouter()
	info, _ := configVars.es.Info()
	log.Println("Server Started:", info)
	http.ListenAndServe("localhost:8081", configVars.router)
}

func GetESClient() *elasticsearch.Client {
	return configVars.es
}

func GetRouter() *mux.Router {
	return configVars.router
}

