package models

import (
	"time"
)

type StatementResponse struct {
	Account any `json:"account"`
	Transactions []Entry `json:"transactions"`
}

type Entry struct {
	AccountNo string `json:"account_no"`
	Amount float64 `json:"amount"`
	Type string `json:"type"`
	Date time.Time `json:"date"`
}