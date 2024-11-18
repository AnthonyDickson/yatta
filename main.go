package main

import (
	"log"
	"net/http"
	"os"
)

// TODO: Replace dummy user store with real one
type DummyUserStore struct{}

func (d *DummyUserStore) CreateUser(email, password string) error {
	return nil
}

const dbFileName = "todos.db.json"

func main() {
	database, err := os.OpenFile(dbFileName, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		log.Fatalf("could not open file %s: %v", dbFileName, err)
	}

	store := NewFileTaskStore(database)
	renderer, err := NewHTMLRenderer()

	if err != nil {
		log.Fatalf("an error occurred while creating the HTML renderer: %v", err)
	}

	server, err := NewServer(store, new(DummyUserStore), renderer)

	if err != nil {
		log.Fatalf("an error occurred while creating the server: %v", err)
	}

	handler := http.Handler(server)
	log.Fatal(http.ListenAndServe(":8000", handler))
}
