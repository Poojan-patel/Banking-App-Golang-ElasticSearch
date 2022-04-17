package models

import "github.com/dgrijalva/jwt-go"

type JWTClaim struct{
	UserId string `json:"user_id"`
	jwt.StandardClaims
}