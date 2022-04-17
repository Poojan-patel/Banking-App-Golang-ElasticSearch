package models

type JWTToken struct {
	UserId string `json:"user_id"`
	Token string `json:"token"`
}