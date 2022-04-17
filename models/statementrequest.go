package models

type StatementRequest struct {
	AccountNo string `json:"account_no"`
	UserId string `json:"user_id"`
	LastTransactions int `json:"last_transactions"`
}