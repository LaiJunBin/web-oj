package userquery

import (
	"errors"
	"time"
)

var (
	ErrorUserNotFound = errors.New("user not found")
)

type User struct {
	ID        int32
	Name      string
	Username  string
	Roles     []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserQueryService interface {
	ListUser() ([]User, error)
	GetUserByID(id int32) (*User, error)
}
