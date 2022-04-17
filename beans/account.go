package beans

type Account struct {
	AccountNo string `json:"account_no"`
	UserId string `json:"user_id"`
	Balance float64 `json:"balance"`
	Name string `json:"name"`
	Aadhar string `json:"aadhar"`
	Mobile string `json:"mobile"`
}