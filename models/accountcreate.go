package models

type AccountCreate struct {
	Name string `json:"name"`
	Mobile string `json:"mobile"`
	Aadhar string `json:"aadhar"`
}