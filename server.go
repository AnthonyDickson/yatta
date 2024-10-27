package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
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

var (
	//go:embed "templates/*"
	todosListTemplate embed.FS
)

func (s *Server) getTodos(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	todos := s.store.GetTodos(user)

	if todos == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// TODO: Parse templates once in constructor.
	tmpl, err := template.ParseFS(todosListTemplate, "templates/*.html")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("an error occurred while parsing the template for the route %q: %v", r.URL, err))
		return
	}

	// TODO: Separate out rendering logic and error handling.

	body := new(bytes.Buffer)

	if err := tmpl.ExecuteTemplate(body, "todo_list.html", todos); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("an error occurred while rendering the template for the route %q with data %q: %v", r.URL, todos, err))
		return
	}

	w.Header().Add("content-type", htmlContentType)
	_, err = w.Write(body.Bytes())

	if err != nil {
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
