package stores

import (
	"github.com/AnthonyDickson/yatta/models"
)

// UserStore is an interface for storing and retrieving users.
type UserStore interface {
	// AddUser adds a new user to the store.
	//
	// Email addresses must be unique.
	//
	// Returns an error if the email address is already in use.
	AddUser(email string, password *models.PasswordHash) error

	// GetUser retrieves a user by their ID.
	GetUser(id uint64) (*models.User, error)

	// GetUsers retrieves all users.
	GetUsers() ([]models.User, error)

	// CheckEmailInUse checks if a user is already using `email`.
	EmailInUse(email string) bool
}
