package models

type UserChangePassword struct {
	UserId string `json:"user_id"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}