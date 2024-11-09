package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const htmlContentType = "text/html"

type Server struct {
	store    TaskStore
	renderer Renderer
	http.Handler
}

func NewServer(store TaskStore, renderer Renderer) (*Server, error) {
	server := new(Server)
	server.store = store

	router := http.NewServeMux()
	router.Handle("GET /coffee", http.HandlerFunc(server.getCoffee))
	router.Handle("GET /users/{user}/tasks", http.HandlerFunc(server.getTasks))
	router.Handle("POST /users/{user}/tasks", http.HandlerFunc(server.addTask))

	server.Handler = router

	server.renderer = renderer

	return server, nil
}

func (s *Server) getCoffee(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

func (s *Server) getTasks(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	tasks, err := s.store.GetTasks(user)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("an error occurred while getting the tasks for %s: %v", r.URL, err))
		return
	}

	if tasks == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body, err := s.renderer.RenderTaskList(tasks)

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

func (s *Server) addTask(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	bodyBytes, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Warn(fmt.Sprintf("an error occurred while reading the request body %v: %v", r.Body, err))
		return
	}

	task := string(bodyBytes)

	err = s.store.AddTask(user, task)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("could not add task %q for user %q: %v", user, task, err))
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
