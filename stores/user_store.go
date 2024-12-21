package stores

import (
	"github.com/AnthonyDickson/yatta/models"
)

// UserStore is an interface for storing and retrieving users.
type UserStore interface {
	// AddUser adds a new user to the store.
	AddUser(email string, password *models.PasswordHash) error

	// GetUser retrieves a user by their ID.
	GetUser(id uint64) (*models.User, error)

	// GetUsers retrieves all users.
	GetUsers() ([]models.User, error)
}
