package stores_test

import (
	"reflect"
	"testing"

	"github.com/AnthonyDickson/yatta/models"
	"github.com/AnthonyDickson/yatta/stores"
	"github.com/AnthonyDickson/yatta/yattatest"
)

func TestFileTaskStore(t *testing.T) {
	t.Run("load store from reader", func(t *testing.T) {
		database, cleanup := yattatest.CreateTempFile(t, `[
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

		store := stores.NewFileTaskStore(database)

		assertTasks(t, store, "Alice", []models.Task{
			{ID: 0, Description: "send message to Bob"},
			{ID: 1, Description: "upgrade encryption"},
			{ID: 2, Description: "read message from Bob"},
		})
		assertTasks(t, store, "Bob", []models.Task{
			{ID: 3, Description: "read message from Alice"},
			{ID: 4, Description: "send message to Alice"},
		})
	})

	t.Run("get a task by id", func(t *testing.T) {
		database, cleanup := yattatest.CreateTempFile(t, `[
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

		store := stores.NewFileTaskStore(database)

		assertGetTask(t, store, 0, models.Task{ID: 0, Description: "send message to Bob"})
		assertGetTask(t, store, 1, models.Task{ID: 1, Description: "upgrade encryption"})
		assertGetTask(t, store, 2, models.Task{ID: 2, Description: "read message from Bob"})
	})

	t.Run("add task for existing user", func(t *testing.T) {
		database, cleanup := yattatest.CreateTempFile(t, `[
        {
          "user": "Alice",
          "tasks": []
        }
      ]`)
		defer cleanup()
		store := stores.NewFileTaskStore(database)

		err := store.AddTask("Alice", "find the keys")

		yattatest.AssertNoError(t, err)
		assertTasks(t, store, "Alice", []models.Task{{ID: 0, Description: "find the keys"}})
	})

	t.Run("add task for new user", func(t *testing.T) {
		database, cleanup := yattatest.CreateTempFile(t, `[]`)
		defer cleanup()
		store := stores.NewFileTaskStore(database)

		err := store.AddTask("Alice", "find the keys")

		yattatest.AssertNoError(t, err)
		assertTasks(t, store, "Alice", []models.Task{{ID: 0, Description: "find the keys"}})
	})
}

func assertGetTask(t *testing.T, store *stores.FileTaskStore, id uint64, want models.Task) {
	t.Helper()

	got, err := store.GetTask(id)
	yattatest.AssertNoError(t, err)

	if *got != want {
		t.Errorf("got task %v want %v", *got, want)
	}
}

func assertTasks(t *testing.T, store *stores.FileTaskStore, user string, want []models.Task) {
	t.Helper()

	got, err := store.GetTasks(user)

	if err != nil {
		t.Fatalf("got error %v, want no error", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got tasks %q, want %q", got, want)
	}
}
