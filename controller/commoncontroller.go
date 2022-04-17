package controller

import (
	"github.com/gorilla/mux"
)

func CreateRouter() *mux.Router {
	r := mux.NewRouter()
	AddAccountRoutes(r.PathPrefix("/account").Subrouter())
	AddUserRoutes(r.PathPrefix("/user").Subrouter())
	return r
}

