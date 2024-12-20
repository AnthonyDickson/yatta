package models

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// A PasswordHash stores a hashed and salted password.
type PasswordHash struct {
	// The Hash and salt.
	Hash []byte
}

// NewPasswordHash hashes and salts a plaintext password with the specified cost.
//
// See [golang.org/x/crypto/bcrypt.DefaultCost],
// [golang.org/x/crypto/bcrypt.MinCost] and [golang.org/x/crypto/bcrypt.MaxCost]
// for the possible values for cost.
func NewPasswordHash(password string, cost int) (*PasswordHash, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)

	return &PasswordHash{hash}, err
}

// Compare compares the password hash against a plaintext password.
//
// Returns nil if the password matches, or an error otherwise.
func (p *PasswordHash) Compare(password string) error {
	return bcrypt.CompareHashAndPassword(p.Hash, []byte(password))
}

// MarshalJSON encodes the password hash as a quoted string.
func (p *PasswordHash) MarshalJSON() ([]byte, error) {
	// Marshal the hash as a quoted string to prevent it being encoded with Go struct syntax ('&{...}').
	return []byte(fmt.Sprintf("%q", p.Hash)), nil
}

// UnmarshalJSON decodes the password hash from a quoted string.
func (p *PasswordHash) UnmarshalJSON(data []byte) error {
	dataWithoutQuotes := data[1 : len(data)-1]
	p.Hash = dataWithoutQuotes
	return nil
}
