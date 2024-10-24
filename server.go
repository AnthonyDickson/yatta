package main

import (
	"fmt"
	"net/http"
)

type TodoStore interface {
	GetTodos(name string) string
}

type Server struct {
	store TodoStore
	http.Handler
}

func NewServer(store TodoStore) *Server {
	server := new(Server)
	server.store = store

	router := http.NewServeMux()
	router.Handle("/coffee", http.HandlerFunc(server.getCoffee))
	router.Handle("/users/{user}/todos", http.HandlerFunc(server.getTodos))

	server.Handler = router

	return server
}

func (s *Server) getCoffee(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

func (s *Server) getTodos(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	todos := s.store.GetTodos(user)

	if todos == "" {
		w.WriteHeader(http.StatusNotFound)
	}

	fmt.Fprint(w, todos)
}
