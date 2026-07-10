package model

import "time"

type User struct {
	ID          uint
	Username    string
	Email       string
	Password    string
	Role        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
