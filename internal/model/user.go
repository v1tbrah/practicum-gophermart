package model

import "strconv"

type User struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password,omitempty" binding:"required"`
	ID       int64  `json:"id"`
}

func (u *User) String() string {
	if u == nil {
		return "user is nil pointer"
	}

	return "ID: " + strconv.Itoa(int(u.ID)) +
		" Login: " + u.Login
}
