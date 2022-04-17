package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/models"
	"github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/services"
	"github.com/gorilla/mux"
)

func AddAccountRoutes(r *mux.Router){
	r.HandleFunc("/create", AccountCreationRequest).Methods("POST")
	r.HandleFunc("/deposit", Deposit).Methods("POST")
	r.HandleFunc("/withdraw", Withdraw).Methods("POST")
	r.HandleFunc("/neft", Neft).Methods("POST")
	r.HandleFunc("/statement", AccountStatement).Methods("POST")
}

func AccountCreationRequest(w http.ResponseWriter, r *http.Request){
	var accountCreationRequest models.AccountCreate
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&accountCreationRequest)
	if err != nil {
		fmt.Fprintf(w, "Failed to Parse Account Creation Request")
		return	
	}
	acc, usr, err := services.CreateBankAccount(&accountCreationRequest)
	if err != nil{
		fmt.Fprintf(w, "Request Failed With Error: %+v", err.Error())
		return
	}
	if usr != nil {
		fmt.Fprintf(w, "Account: %+v\nUser: %+v", acc, usr)	
	} else {
		fmt.Fprintf(w, "Account: %+v\nLogin with your existing Netbanking account, to see all your accounts", acc)	
	}
}

func Deposit(w http.ResponseWriter, r *http.Request){
	var trans models.CashTransaction
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&trans)
	if err != nil {
		fmt.Fprintf(w, "Failed to Parse Account Creation Request")
		return
	}
	authTok := r.Header.Get("auth_tok_banking_system")
	if authTok == "" {
		fmt.Fprintf(w, "Unauthorized to Process Request\nPlease Login First\n")
		return
	}
	err = services.DepositMoney(&trans, &authTok)
	if err != nil {
		fmt.Fprintf(w, "Unable to Process Transaction with following error:%s", err)
	} else{
		fmt.Fprintf(w, "Transaction Successful")
	}
}

func Withdraw(w http.ResponseWriter, r *http.Request){
	var trans models.CashTransaction
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&trans)
	if err != nil {
		fmt.Fprintf(w, "Failed to Parse Account Creation Request")
		return
	}
	authTok := r.Header.Get("auth_tok_banking_system")
	if authTok == "" {
		fmt.Fprintf(w, "Unauthorized to Process Request\nPlease Login First\n")
		return
	}
	err = services.WithdrawMoney(&trans, &authTok)
	if err != nil {
		fmt.Fprintf(w, "Unable to Process Transaction with following error:%s", err)
	} else{
		fmt.Fprintf(w, "Transaction Successful")
	}
}

func Neft(w http.ResponseWriter, r *http.Request){
	var trans models.NEFT
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&trans)
	if err != nil {
		fmt.Fprintf(w, "Failed to Parse Account Creation Request")
		return
	}
	authTok := r.Header.Get("auth_tok_banking_system")
	if authTok == "" {
		fmt.Fprintf(w, "Unauthorized to Process Request\nPlease Login First\n")
		return
	}
	err = services.NeftTransfer(&trans, &authTok)
	if err != nil {
		fmt.Fprintf(w, "Unable to Process Transaction with following error:%s", err)
	} else{
		fmt.Fprintf(w, "Transaction Successful")
	}
}

func AccountStatement(w http.ResponseWriter, r *http.Request){
	var stmtReq models.StatementRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&stmtReq)
	if err != nil {
		fmt.Fprintf(w, "Failed to Parse Account Creation Request")
		return
	}
	authTok := r.Header.Get("auth_tok_banking_system")
	if authTok == "" {
		fmt.Fprintf(w, "Unauthorized to Process Request\nPlease Login First\n")
		return
	}
	stmt, err := services.GenerateStatement(&stmtReq, &authTok)
	if err != nil {
		fmt.Fprintf(w, "Error Thrown While Processing Your Request: %s", err)
	} else {
		jsonResp, err := json.Marshal(stmt)
		if err != nil {
			fmt.Fprintf(w, "%+v", stmt)
		} else{
			w.Write(jsonResp)
		}
	}
}