package main

import (
	"log"
	"net/http"
)

type InMemoryTodoStore struct{}

func (i *InMemoryTodoStore) GetTodos(user string) string {
	return ""
}

func main() {
	store := new(InMemoryTodoStore)
	server := NewServer(store)
	handler := http.Handler(server)
	log.Fatal(http.ListenAndServe(":8000", handler))
}
