package models

type NEFT struct {
	FromAccountNo string `json:"from_account_no"`
	ToAccountNo string `json:"to_account_no"`
	Amount float64 `json:"amount"`
	UserId string `json:"user_id"`
}