package user

import (
	"time"
)

type User struct {
	ID             ID
	Username       Username
	Email          Email
	HashedPassword HashedPassword
	Role           Role
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ID uint
type HashedPassword string
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)
