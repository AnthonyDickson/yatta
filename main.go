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

func (i *InMemoryTodoStore) GetTodos(user string) []string {
	return i.todos[user]
}

func (i *InMemoryTodoStore) AddTodo(user string, task string) {
	i.todos[user] = append(i.todos[user], task)
}

func main() {
	store := NewInMemoryTodoStore()
	server := NewServer(store)
	handler := http.Handler(server)
	log.Fatal(http.ListenAndServe(":8000", handler))
}
