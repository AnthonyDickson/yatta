package main

import (
	"log"
	"net/http"
	"os"

	"github.com/AnthonyDickson/yatta/stores"
)

const taskDBFileName = "todos.db.json"
const userDBFileName = "users.db.json"

func main() {
	userStore := createUserStore()
	taskStore := createTaskStore()
	renderer, err := NewHTMLRenderer()

	if err != nil {
		log.Fatalf("an error occurred while creating the HTML renderer: %v", err)
	}

	server, err := NewServer(taskStore, userStore, renderer)

	if err != nil {
		log.Fatalf("an error occurred while creating the server: %v", err)
	}

	handler := http.Handler(server)
	log.Fatal(http.ListenAndServe(":8000", handler))
}

func createTaskStore() *stores.FileTaskStore {
	database, err := os.OpenFile(taskDBFileName, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		log.Fatalf("could not open file %s: %v", taskDBFileName, err)
	}

	store, err := stores.NewFileTaskStore(database)

	if err != nil {
		log.Fatalf("could not load the file task store: %v", err)
	}

	return store
}

func createUserStore() *stores.FileUserStore {
	database, err := os.OpenFile(userDBFileName, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		log.Fatalf("could not open file %s: %v", userDBFileName, err)
	}

	store, err := stores.NewFileUserStore(database)

	if err != nil {
		log.Fatalf("could not load the user task store: %v", err)
	}

	return store
}
