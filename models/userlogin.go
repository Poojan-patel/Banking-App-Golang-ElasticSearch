package models

type UserLogin struct{
	UserId string `json:"user_id"`
	Password string `json:"password"`
}