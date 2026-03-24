package server

import "time"

const AUTH_DISABLE int32 = -1

type TerminalUser struct {
	Username   string
	Password   string
	Expiration *time.Time
}

type UserRepository interface {
	CreateUser(string, password string) (*TerminalUser, error)
	AddUser(user *TerminalUser) (bool, error)
	ClearUsers() error
	FindUser(username string) *TerminalUser
	GetUsers() []*TerminalUser
	Load() error
	RemoveUser(username string) *TerminalUser
	Save()
}
