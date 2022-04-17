package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
	"github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/beans"
	"github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/config"
	"github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/models"
	"github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/repository"
	"github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/validators"
	"math"
	"strconv"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func CreateBankAccount(accountCreateRequest *models.AccountCreate) (*beans.Account, *beans.User, error) {
	err := validators.AccountCreateValidation(accountCreateRequest)
	if err != nil{
		return nil, nil, err
	}
	timestamp := time.Now().UnixMilli()
	userId, err := doesAccountExistsWithSameAadhar(&accountCreateRequest.Aadhar)
	if err != nil {
		return nil, nil, err
	}
	acc := &beans.Account{
		Balance: 0,
		Mobile: accountCreateRequest.Mobile,
		Aadhar: accountCreateRequest.Aadhar,
		Name: accountCreateRequest.Name,
		UserId: fmt.Sprint(timestamp/1000),
		AccountNo: fmt.Sprintf("%v%v", timestamp%1000, timestamp/1000),
	}
	if userId != "" {
		acc.UserId = userId
	}
	err = save(acc)
	if err != nil{
		return nil, nil, err
	}
	if userId == "" {
		usr, err := createUserAccount(acc)
		if err != nil{
			return nil, nil, err
		}
		return acc, usr, err
	} else {
		return acc, nil, err
	}
}

func doesAccountExistsWithSameAadhar(aadhar *string) (string, error){
	query_string := fmt.Sprintf(`{"query":{"bool":{"must":{"match_phrase":{"aadhar":"%s"}}}}}`, *aadhar)
	resp, err := repository.SearchQuery(&query_string, "accounts", config.GetESClient())
	if err != nil {
		return "", err
	}
	var resMapper map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&resMapper)
	for _, hit := range resMapper["hits"].(map[string]interface{})["hits"].([]interface{}){
		doc := hit.(map[string]interface{})
		acc := doc["_source"].(map[string]interface{})
		return acc["user_id"].(string), nil
	}
	return "", nil
}

func save(acc *beans.Account) (error) {
	var es *elasticsearch.Client = config.GetESClient()
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(acc)
	if err != nil {
		return err
	}
	var request = esapi.IndexRequest{Index:"accounts", Body:&buf}
	resp, err := request.Do(context.Background(), es)
	if err != nil{
		return err
	}
	defer resp.Body.Close()
	if resp.IsError(){
		log.Println(resp.StatusCode)
		return errors.New("Error Occurred with following Status code" + (string)(resp.StatusCode))
	}
	return nil
}

func checkIsDepositable(accountNo *string, amount float64) (any, string, error) {
	query_string := fmt.Sprintf(`{"query":{"bool":{"must":{"match_phrase":{"account_no":"%s"}}}}}`, *accountNo)
	resp, err :=repository.SearchQuery(&query_string, "accounts", config.GetESClient())
	if err != nil {
		return nil, "", err
	}
	var resMapper map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&resMapper)
	hits := resMapper["hits"].(map[string]interface{})["hits"].([]interface{})
	if (len(hits) == 0){
		return nil, "", errors.New("Account Number Invalid")
	}
	id := hits[0].(map[string]interface{})["_id"].(string)
	doc := hits[0].(map[string]interface{})["_source"].(map[string]interface{})
	doc["balance"] = doc["balance"].(float64) + amount
	return doc, id, nil
}

func checkIsWithdrawable(accountNo *string, userId *string, amount float64) (any, string, error){
	query_string := fmt.Sprintf(`{"query":{"bool":{"must":[{"match_phrase":{"user_id":"%s"}}, {"match_phrase":{"account_no":"%s"}}]}}}`, *userId, *accountNo)
	resp, err :=repository.SearchQuery(&query_string, "accounts", config.GetESClient())
	if err != nil {
		return nil, "", err
	}
	var resMapper map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&resMapper)
	hits := resMapper["hits"].(map[string]interface{})["hits"].([]interface{})
	if (len(hits) == 0){
		return nil, "", errors.New("Either Account Number is Invalid or That account doesn't belongs to You")
	}
	id := hits[0].(map[string]interface{})["_id"].(string)
	doc := hits[0].(map[string]interface{})["_source"].(map[string]interface{})
	doc["balance"] = doc["balance"].(float64) - amount
	if(doc["balance"].(float64) < 0){
		return nil, "", errors.New("You cannot withdraw amount more than your balance")
	}
	return doc, id, nil
}

func DepositMoney(trans *models.CashTransaction, authTok *string) error {
	if(trans.Amount <= 0){
		return errors.New("Invalid Amount")
	}
	err := ValidateAuthToken(&trans.UserId, authTok)
	if err != nil {
		return err
	}
	doc, id, err := checkIsDepositable(&trans.AccountNo, trans.Amount)
	transaction := &beans.Transaction{
		FromAccountNo: "User:"+trans.UserId,
		ToAccountNo: trans.AccountNo,
		Amount: trans.Amount,
		Timestamp: time.Now().UnixMicro(),
	}
	err = repository.Upsert(doc, "accounts", id, config.GetESClient())
	if err != nil {
		log.Println(err)
	}
	return repository.Upsert(transaction, "transactions", "", config.GetESClient())
}

func WithdrawMoney(trans *models.CashTransaction, authTok *string) error {
	if(trans.Amount <= 0){
		return errors.New("Invalid Amount")
	}
	err := ValidateAuthToken(&trans.UserId, authTok)
	if err != nil {
		return err
	}
	doc, id, err := checkIsWithdrawable(&trans.AccountNo, &trans.UserId, trans.Amount)
	if err != nil {
		return err
	}
	transaction := &beans.Transaction{
		FromAccountNo: trans.AccountNo,
		ToAccountNo: "self",
		Amount: -trans.Amount,
		Timestamp: time.Now().UnixMicro(),
	}
	err = repository.Upsert(doc, "accounts", id, config.GetESClient())
	if err != nil {
		log.Println(err)
	}
	return repository.Upsert(transaction, "transactions", "", config.GetESClient())
}

func NeftTransfer(trans *models.NEFT, authTok *string) error {
	if(trans.Amount <= 0){
		return errors.New("Invalid Amount")
	}
	err := ValidateAuthToken(&trans.UserId, authTok)
	if err != nil {
		return err
	}
	// checking from account
	doc1, id1, err := checkIsWithdrawable(&trans.FromAccountNo, &trans.UserId, trans.Amount)
	if err != nil {
		return err
	}

	//checking to account
	doc2, id2, err := checkIsDepositable(&trans.ToAccountNo, trans.Amount)
	if err != nil {
		return err
	}
	transaction := &beans.Transaction{
		FromAccountNo: trans.FromAccountNo,
		ToAccountNo: trans.ToAccountNo,
		Amount: trans.Amount,
		Timestamp: time.Now().UnixMicro(),
	}
	err = repository.Upsert(transaction, "transactions", "", config.GetESClient())
	if err != nil {
		return err
	}
	err = repository.Upsert(doc1, "accounts", id1, config.GetESClient())
	if err != nil {
		return err
	}
	return repository.Upsert(doc2, "accounts", id2, config.GetESClient())
}

func GenerateStatement(stmtReq *models.StatementRequest, authTok *string) (*models.StatementResponse, error) {
	if(stmtReq.LastTransactions <= 0){
		return nil, errors.New("Number of transactions asked must be greater than zero")
	}
	err := ValidateAuthToken(&stmtReq.UserId, authTok)
	if err != nil {
		return nil, err
	}
	acc, _, err := checkIsWithdrawable(&stmtReq.AccountNo, &stmtReq.UserId, 0)
	if err != nil {
		return nil, err
	}
	acc.(map[string]interface{})["password"] = ""
	query_string := fmt.Sprintf(`{"query": {"query_string": {"query": "%s"}},"size": %d,"from": 0,"sort": [{"timestamp": {"unmapped_type": "int","order": "desc"}}]}`, stmtReq.AccountNo, stmtReq.LastTransactions)
	resp, err := repository.SearchQuery(&query_string, "transactions", config.GetESClient())
	if err != nil {
		return nil, err
	}
	stmt := &models.StatementResponse{Account: acc}
	var resMapper map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&resMapper)
	for _, hit := range resMapper["hits"].(map[string]interface{})["hits"].([]interface{}){
		doc := hit.(map[string]interface{})["_source"].(map[string]interface{})
		stmt.Transactions = append(stmt.Transactions, handleStatementEntry(doc, stmtReq.AccountNo))
	}
	return stmt, nil
}

func handleStatementEntry(doc map[string]interface{}, accNo string) models.Entry {
	var entry models.Entry;
	amt := doc["amount"].(float64)
	entry.Amount = math.Abs(amt)
	entry.Date = time.UnixMicro(int64(doc["timestamp"].(float64)))
	if(amt < 0){
		entry.Type = "withdraw"
		entry.AccountNo = "self"
	} else if _, err := strconv.ParseInt(doc["from_account_no"].(string), 10, 64); err != nil {
		entry.Type = "deposit"
		entry.AccountNo = doc["from_account_no"].(string)
	} else {
		if doc["from_account_no"].(string) == accNo {
			entry.Type = "NEFT:debit"
			entry.AccountNo = doc["to_account_no"].(string)
		} else{
			entry.Type = "NEFT:credit"
			entry.AccountNo = doc["from_account_no"].(string)
		}
	}
	return entry
}