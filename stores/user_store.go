package stores

type UserStore interface {
	CreateUser(email, password string) error
}
