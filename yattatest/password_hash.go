package yattatest

import (
	"testing"

	"github.com/AnthonyDickson/yatta/models"
	"golang.org/x/crypto/bcrypt"
)

func MustCreatePasswordHash(t *testing.T, password string) models.PasswordHash {
	t.Helper()

	hash, err := models.NewPasswordHash(password, bcrypt.MinCost)
	AssertNoError(t, err)

	return hash
}
