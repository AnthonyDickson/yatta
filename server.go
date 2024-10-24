package main

import (
	"fmt"
	"io"
	"net/http"
)

type TodoStore interface {
	GetTodos(user string) string
	AddTodo(user string, task string)
}

type Server struct {
	store TodoStore
	http.Handler
}

func NewServer(store TodoStore) *Server {
	server := new(Server)
	server.store = store

	router := http.NewServeMux()
	router.Handle("GET /coffee", http.HandlerFunc(server.getCoffee))
	router.Handle("GET /users/{user}/todos", http.HandlerFunc(server.getTodos))
	router.Handle("POST /users/{user}/todos", http.HandlerFunc(server.addTodo))

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

func (s *Server) addTodo(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	bodyBytes, _ := io.ReadAll(r.Body)
	task := string(bodyBytes)

	s.store.AddTodo(user, task)
	w.WriteHeader(http.StatusAccepted)
}
