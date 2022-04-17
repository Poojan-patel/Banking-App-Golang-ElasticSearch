package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/models"
	"github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/services"
	"github.com/gorilla/mux"
)


func AddUserRoutes(r *mux.Router){
	r.HandleFunc("/login", AccountLogin).Methods("POST")
	r.HandleFunc("/change_password", ChangePassword).Methods("POST")
	r.HandleFunc("/find_user", FindTheUser).Methods("POST")
}

func AccountLogin(w http.ResponseWriter, r *http.Request){
	var loginRequest models.UserLogin
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&loginRequest)
	if err != nil {
		fmt.Fprintf(w, "Login Unsuccessful")
		return	
	} 
	authTok := r.Header.Get("auth_tok_banking_system")
	if authTok == "" {
		authTok, err := services.GenerateAuthToken(&loginRequest)
		if err != nil{
			fmt.Fprintf(w, "Error Thrown While Authenticating: %s", err)
		} else{
			w.Header().Set("auth_tok_banking_system", *authTok)
			fmt.Fprintf(w, "You are Authenticated\nPlease Add this key-value pair in your headers\n'auth_tok_banking_system':%s", *authTok)
		}
	} else{
		err := services.ValidateAuthToken(&loginRequest.UserId, &authTok)
		if err != nil {
			fmt.Fprintf(w, "Error Thrown While Authenticating: %s", err)
		} else{
			fmt.Fprintf(w, "You are already authenticated")
		}
	}
}

func ChangePassword(w http.ResponseWriter, r *http.Request){
	var changePassword models.UserChangePassword
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&changePassword)
	if err != nil {
		fmt.Fprintln(w, "ChangePassword Unsuccessful: ", err.Error())
		return
	}
	err = services.ChangePassword(&changePassword)
	if err != nil {
		fmt.Fprintln(w, "Change Password Unsuccessful: ", err.Error())
	} else{
		fmt.Fprintln(w, "Change Password Success")
	}
}

func FindTheUser(w http.ResponseWriter, r *http.Request){
	var loginRequest models.UserLogin
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&loginRequest)
	if err != nil {
		fmt.Fprintln(w, "ChangePassword Unsuccessful: ", err.Error())
		return
	}
	doc, err := services.FindUser(&loginRequest.UserId, &loginRequest.Password)
	if err != nil || doc == nil {
		fmt.Fprintln(w, "User Not Found")	
	} else {
		fmt.Fprintln(w, "User Found")
	}
}