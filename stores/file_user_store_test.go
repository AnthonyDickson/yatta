package stores_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/AnthonyDickson/yatta/models"
	"github.com/AnthonyDickson/yatta/stores"
	"github.com/AnthonyDickson/yatta/yattatest"
)

func TestFileUserStore_New(t *testing.T) {
	t.Run("load store from file", func(t *testing.T) {
		want_users := []models.User{
			{ID: 0, Email: "alice@example.com", Password: yattatest.MustCreatePasswordHash(t, "averysecretpassword")},
			{ID: 1, Email: "bob@example.com", Password: yattatest.MustCreatePasswordHash(t, "anotherverysecretpassword")},
		}

		database, cleanup := yattatest.CreateTempFile(t, fmt.Sprintf(`[
        {
          "ID": 0,
          "Email": "alice@example.com",
          "Password": %q
        },
        {
          "ID": 1,
          "Email": "bob@example.com",
          "Password": %q
        }
      ]`, want_users[0].Password.Hash, want_users[1].Password.Hash))
		defer cleanup()

		store := mustCreateFileUserStore(t, database)

		assertStoreHasUsers(t, store, want_users)
	})
}

func TestFileUserStore_Add(t *testing.T) {
	t.Run("adding new user updates store and database file", func(t *testing.T) {
		database, cleanup := yattatest.CreateTempFile(t, "")
		defer cleanup()
		store := mustCreateFileUserStore(t, database)

		want := models.User{ID: 1, Email: "test@example.com", Password: yattatest.MustCreatePasswordHash(t, "averysecretpassword")}

		err := store.AddUser(want.Email, want.Password)
		yattatest.AssertNoError(t, err)
		assertStoreHasUsers(t, store, []models.User{want})

		storeAfterAdd := mustCreateFileUserStore(t, database)
		assertStoreHasUsers(t, storeAfterAdd, []models.User{want})
	})

	t.Run("adding users increments ID", func(t *testing.T) {
		database, cleanup := yattatest.CreateTempFile(t, "")
		defer cleanup()

		store := mustCreateFileUserStore(t, database)
		want := []models.User{
			{ID: 1, Email: "test@example.com", Password: yattatest.MustCreatePasswordHash(t, "averysecretpassword")},
			{ID: 2, Email: "test2@example.com", Password: yattatest.MustCreatePasswordHash(t, "anotherverysecretpassword")},
		}

		for _, user := range want {
			err := store.AddUser(user.Email, user.Password)
			yattatest.AssertNoError(t, err)
		}

		assertStoreHasUsers(t, store, want)
	})
}

func mustCreateFileUserStore(t *testing.T, database *os.File) *stores.FileUserStore {
	t.Helper()

	store, err := stores.NewFileUserStore(database)

	if err != nil {
		t.Fatalf("could not create FileUserStore: %v", err)
	}

	return store
}

func assertStoreHasUsers(t *testing.T, store *stores.FileUserStore, want_users []models.User) {
	t.Helper()

	for _, want := range want_users {
		got, err := store.GetUser(want.ID)
		yattatest.AssertNoError(t, err)

		if got == nil {
			t.Fatalf("got nil user, want %v", want)
		}

		if got.ID != want.ID || got.Email != want.Email {
			t.Errorf("got user %v, want %v", *got, want)
		}
	}
}
