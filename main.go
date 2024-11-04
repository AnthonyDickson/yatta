package main

import (
	"log"
	"net/http"
)

type InMemoryTodoStore struct {
	todos map[string][]string
}

func NewInMemoryTodoStore() *InMemoryTodoStore {
	store := new(InMemoryTodoStore)
	store.todos = make(map[string][]string)

	return store
}

func (i *InMemoryTodoStore) GetTodos(user string) ([]string, error) {
	return i.todos[user], nil
}

func (i *InMemoryTodoStore) AddTodo(user string, task string) error {
	i.todos[user] = append(i.todos[user], task)

	return nil
}

func main() {
	store := NewInMemoryTodoStore()
	renderer, err := NewHTMLRenderer()

	if err != nil {
		log.Fatalf("an error occurred while creating the HTML renderer: %v", err)
	}

	server, err := NewServer(store, renderer)

	if err != nil {
		log.Fatalf("an error occurred while creating the server: %v", err)
	}

	handler := http.Handler(server)
	log.Fatal(http.ListenAndServe(":8000", handler))
}
