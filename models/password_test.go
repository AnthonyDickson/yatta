package models_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/AnthonyDickson/yatta/models"
	"github.com/AnthonyDickson/yatta/yattatest"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordHash_New(t *testing.T) {
	t.Run("can create hashed password", func(t *testing.T) {
		password := "averysecurepassword"

		_, err := models.NewPasswordHash(password, bcrypt.MinCost)
		yattatest.AssertNoError(t, err)
	})

	// Do not need to test for cost below the minimum since the minimum cost is
	// automatically used instead.
	t.Run("cannot use cost above maximum", func(t *testing.T) {
		_, err := models.NewPasswordHash("mypassword", bcrypt.MaxCost+1)

		if err == nil {
			t.Error("got nil error, want InvalidCostError")
		}
	})
}

func TestPasswordHash_compare(t *testing.T) {
	t.Run("can validate password from hash", func(t *testing.T) {
		password := "GnomeChompy"

		hash := yattatest.MustCreatePasswordHash(t, password)

		err := hash.Compare(password)
		yattatest.AssertNoError(t, err)
	})

	t.Run("wrong password fails validation with hash", func(t *testing.T) {
		password := "FrederickAngles"
		wrongPassword := "FrederickKayak"

		hash := yattatest.MustCreatePasswordHash(t, password)

		err := hash.Compare(wrongPassword)

		if err == nil {
			t.Errorf("got nil error (password matches), want error for passwords %q and %q", password, wrongPassword)
		}
	})
}

func TestPasswordHash_JSON(t *testing.T) {
	t.Run("can marshal and unmarshal JSON", func(t *testing.T) {
		want := yattatest.MustCreatePasswordHash(t, "EdwardBerenstein")

		bytes, err := json.Marshal(&want)
		yattatest.AssertNoError(t, err)

		var got models.PasswordHash
		err = json.Unmarshal(bytes, &got)
		yattatest.AssertNoError(t, err)

		if !slices.Equal(got, want) {
			t.Errorf("got %s, want %s", got, want)
		}
	})
}
