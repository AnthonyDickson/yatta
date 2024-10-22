package main

import (
	"fmt"
	"net/http"
)

type Server struct {
	http.Handler
}

func NewServer() *Server {
	server := new(Server)

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

	if user == "Alice" {
		fmt.Fprint(w, "send message to Bob")
		return
	}

	fmt.Fprint(w, "write more code")
}
