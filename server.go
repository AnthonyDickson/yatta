package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const htmlContentType = "text/html"

type TodoStore interface {
	GetTodos(user string) []string
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

	if todos == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Add("content-type", htmlContentType)
	// TODO: Render HTML with templates
	fmt.Fprint(w, "<ul>")
	for _, todo := range todos {
		fmt.Fprintf(w, "<li>%s</li>", todo)
	}
	fmt.Fprint(w, "</ul>")
}

func (s *Server) addTodo(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	bodyBytes, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Warn(fmt.Sprintf("an error occurred while reading the request body %v: %v", r.Body, err))
		return
	}

	task := string(bodyBytes)

	s.store.AddTodo(user, task)
	w.WriteHeader(http.StatusAccepted)
}
