package stores

import (
	"github.com/AnthonyDickson/yatta/models"
)

type UserStore interface {
	AddUser(email, password string) error
	GetUser(id uint64) (*models.User, error)
}
