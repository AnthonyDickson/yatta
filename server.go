package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
)

const htmlContentType = "text/html"

type Server struct {
	userStore UserStore
	taskStore TaskStore
	renderer  Renderer
	http.Handler
}

func NewServer(taskStore TaskStore, userStore UserStore, renderer Renderer) (*Server, error) {
	server := new(Server)
	server.taskStore = taskStore
	server.userStore = userStore

	router := http.NewServeMux()
	router.Handle("GET /coffee", http.HandlerFunc(server.getCoffee))
	router.Handle("GET /tasks/{id}", http.HandlerFunc(server.getTask))
	router.Handle("GET /users/{user}/tasks", http.HandlerFunc(server.getTasks))
	router.Handle("POST /users/{user}/tasks", http.HandlerFunc(server.addTask))
	router.Handle("POST /users", http.HandlerFunc(server.createUser))

	server.Handler = router

	server.renderer = renderer

	return server, nil
}

func (s *Server) getCoffee(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	id_string := r.PathValue("id")
	id, err := strconv.ParseUint(id_string, 10, 64)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	task, err := s.taskStore.GetTask(id)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("could not get task with ID %q with URL %q: %v", id, r.URL, err))
		return
	}

	if task == nil {
		http.NotFound(w, r)
		return
	}

	body, err := s.renderer.RenderTask(*task)
	writeResponse(w, body, err, r.URL)
}

func (s *Server) getTasks(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	tasks, err := s.taskStore.GetTasks(user)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("an error occurred while getting the tasks for %s: %v", r.URL, err))
		return
	}

	if tasks == nil {
		http.NotFound(w, r)
		return
	}

	body, err := s.renderer.RenderTaskList(tasks)
	writeResponse(w, body, err, r.URL)
}

func writeResponse(w http.ResponseWriter, body []byte, err error, requestURL *url.URL) {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("an error occurred while rendering the template for %s: %v", requestURL, err))
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

	err = s.taskStore.AddTask(user, task)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("could not add task %q for user %q: %v", user, task, err))
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

const formContentType = "application/x-www-form-urlencoded"

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != formContentType {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	err := r.ParseForm()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("could not parse form: %v", err))
		return
	}

	if !r.Form.Has("email") || !r.Form.Has("password") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	err = s.userStore.CreateUser(email, password)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("could not create user: %v", err))
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
