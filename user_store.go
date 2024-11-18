package main

type UserStore interface {
	CreateUser(email, password string) error
}
