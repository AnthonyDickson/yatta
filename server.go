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
	store    TodoStore
	renderer Renderer
	http.Handler
}

func NewServer(store TodoStore) (*Server, error) {
	server := new(Server)
	server.store = store

	router := http.NewServeMux()
	router.Handle("GET /coffee", http.HandlerFunc(server.getCoffee))
	router.Handle("GET /users/{user}/todos", http.HandlerFunc(server.getTodos))
	router.Handle("POST /users/{user}/todos", http.HandlerFunc(server.addTodo))

	server.Handler = router

	renderer, err := NewRenderer()

	if err != nil {
		return nil, fmt.Errorf("an error occurred while creating the template renderer: %v", err)
	}

	server.renderer = *renderer

	return server, nil
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

	body, err := s.renderer.RenderTodosList(todos)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("an error occurred while rendering the template for %s: %v", r.URL, err))
		return
	}

	w.Header().Add("content-type", htmlContentType)

	if _, err := w.Write(body); err != nil {
		slog.Error(fmt.Sprintf("an error occurred while writing the response body: %v", err))
	}
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
