package main_test

import (
	"io"
	"os"
	"reflect"
	"testing"

	yatta "github.com/AnthonyDickson/yatta"
)

func TestFileTodoStore(t *testing.T) {
	t.Run("load store from reader", func(t *testing.T) {
		database, cleanup := createTempFile(t, `[
        {
          "user": "Alice", 
          "tasks": [
            "send message to Bob",
            "upgrade encryption",
            "read message from Bob"
          ]
        },
        {
          "user": "Bob",
          "tasks": [
            "read message from Alice",
            "send message to Bob"
          ]
        }
      ]`)
		defer cleanup()

		store := yatta.NewFileTodoStore(database)

		assertTodos(t, store, "Alice", []string{
			"send message to Bob",
			"upgrade encryption",
			"read message from Bob",
		})
		assertTodos(t, store, "Bob", []string{
			"read message from Alice",
			"send message to Bob",
		})
	})

	t.Run("add todo for existing user", func(t *testing.T) {
		database, cleanup := createTempFile(t, `[
        {
          "user": "Alice",
          "tasks": []
        }
      ]`)
		defer cleanup()
		store := yatta.NewFileTodoStore(database)

		err := store.AddTodo("Alice", "find the keys")

		assertNoError(t, err)
		assertTodos(t, store, "Alice", []string{"find the keys"})
	})

	t.Run("add todo for new user", func(t *testing.T) {
		database, cleanup := createTempFile(t, `[]`)
		defer cleanup()
		store := yatta.NewFileTodoStore(database)

		err := store.AddTodo("Alice", "find the keys")

		assertNoError(t, err)
		assertTodos(t, store, "Alice", []string{"find the keys"})
	})
}

func assertTodos(t *testing.T, store *yatta.FileTodoStore, user string, want []string) {
	t.Helper()

	got, err := store.GetTodos(user)

	if err != nil {
		t.Fatalf("got error %v, want no error", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got tasks %q, want %q", got, want)
	}
}

func createTempFile(t *testing.T, initialData string) (io.ReadWriteSeeker, func()) {
	t.Helper()

	tempFile, err := os.CreateTemp("", "db")

	if err != nil {
		t.Fatalf("could not create temporary file: %v", err)
	}

	_, err = tempFile.Write([]byte(initialData))

	if err != nil {
		t.Fatalf("could not write initial data to temp file: %v", err)
	}

	removeFile := func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}

	return tempFile, removeFile
}
