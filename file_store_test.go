package main_test

import (
	"os"
	"reflect"
	"testing"

	yatta "github.com/AnthonyDickson/yatta"
)

func TestFileTaskStore(t *testing.T) {
	t.Run("load store from reader", func(t *testing.T) {
		database, cleanup := createTempFile(t, `[
        {
          "user": "Alice", 
          "tasks": [
            {"ID": 0, "Description": "send message to Bob"},
            {"ID": 1, "Description": "upgrade encryption"},
            {"ID": 2, "Description": "read message from Bob"}
          ]
        },
        {
          "user": "Bob",
          "tasks": [
            {"ID": 3, "Description": "read message from Alice"},
            {"ID": 4, "Description": "send message to Alice"}
          ]
        }
      ]`)
		defer cleanup()

		store := yatta.NewFileTaskStore(database)

		assertTasks(t, store, "Alice", []yatta.Task{
			{ID: 0, Description: "send message to Bob"},
			{ID: 1, Description: "upgrade encryption"},
			{ID: 2, Description: "read message from Bob"},
		})
		assertTasks(t, store, "Bob", []yatta.Task{
			{ID: 3, Description: "read message from Alice"},
			{ID: 4, Description: "send message to Alice"},
		})
	})

	t.Run("get a task by id", func(t *testing.T) {
		database, cleanup := createTempFile(t, `[
        {
          "user": "Alice", 
          "tasks": [
            {"ID": 0, "Description": "send message to Bob"},
            {"ID": 1, "Description": "upgrade encryption"},
            {"ID": 2, "Description": "read message from Bob"}
          ]
        }
      ]`)
		defer cleanup()

		store := yatta.NewFileTaskStore(database)

		assertGetTask(t, store, 0, yatta.Task{ID: 0, Description: "send message to Bob"})
		assertGetTask(t, store, 1, yatta.Task{ID: 1, Description: "upgrade encryption"})
		assertGetTask(t, store, 2, yatta.Task{ID: 2, Description: "read message from Bob"})
	})

	t.Run("add task for existing user", func(t *testing.T) {
		database, cleanup := createTempFile(t, `[
        {
          "user": "Alice",
          "tasks": []
        }
      ]`)
		defer cleanup()
		store := yatta.NewFileTaskStore(database)

		err := store.AddTask("Alice", "find the keys")

		assertNoError(t, err)
		assertTasks(t, store, "Alice", []yatta.Task{{ID: 0, Description: "find the keys"}})
	})

	t.Run("add task for new user", func(t *testing.T) {
		database, cleanup := createTempFile(t, `[]`)
		defer cleanup()
		store := yatta.NewFileTaskStore(database)

		err := store.AddTask("Alice", "find the keys")

		assertNoError(t, err)
		assertTasks(t, store, "Alice", []yatta.Task{{ID: 0, Description: "find the keys"}})
	})
}

func assertGetTask(t *testing.T, store *yatta.FileTaskStore, id uint64, want yatta.Task) {
	t.Helper()

	got, err := store.GetTask(id)
	assertNoError(t, err)

	if *got != want {
		t.Errorf("got task %v want %v", *got, want)
	}
}

func assertTasks(t *testing.T, store *yatta.FileTaskStore, user string, want []yatta.Task) {
	t.Helper()

	got, err := store.GetTasks(user)

	if err != nil {
		t.Fatalf("got error %v, want no error", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got tasks %q, want %q", got, want)
	}
}

func createTempFile(t *testing.T, initialData string) (*os.File, func()) {
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
