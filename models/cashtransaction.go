package models

type CashTransaction struct {
	AccountNo string `json:"account_no"`
	Amount float64 `json:"amount"`
	UserId string `json:"user_id"`
}