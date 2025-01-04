package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"

	"github.com/AnthonyDickson/yatta/models"
	"github.com/AnthonyDickson/yatta/stores"
	"golang.org/x/crypto/bcrypt"
)

const (
	contentTypeHeader = "Content-Type"
	formContentType   = "application/x-www-form-urlencoded"
	htmlContentType   = "text/html"
)

type Server struct {
	userStore stores.UserStore
	taskStore stores.TaskStore
	renderer  Renderer
	http.Handler
}

func NewServer(taskStore stores.TaskStore, userStore stores.UserStore, renderer Renderer) (*Server, error) {
	server := new(Server)
	server.taskStore = taskStore
	server.userStore = userStore

	// TODO: Prefix API routes with /api. Only view routes should be at the root.
	router := http.NewServeMux()
	router.Handle("GET /coffee", http.HandlerFunc(server.getCoffee))
	router.Handle("GET /{$}", http.HandlerFunc(server.getRoot))
	router.Handle("GET /tasks/{id}", http.HandlerFunc(server.getTask))
	router.Handle("GET /users/{user}/tasks", http.HandlerFunc(server.getTasks))
	router.Handle("POST /users/{user}/tasks", http.HandlerFunc(server.addTask))
	router.Handle("POST /users", http.HandlerFunc(server.createUser))
	router.Handle("GET /register", http.HandlerFunc(server.getRegisterPage))

	server.Handler = router

	server.renderer = renderer

	return server, nil
}

func (s *Server) getCoffee(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

func (s *Server) getRoot(w http.ResponseWriter, r *http.Request) {
	users, err := s.userStore.GetUsers()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("could not get users: %v", err))
		return
	}

	body, err := s.renderer.RenderIndex(users)
	writeResponse(w, body, err, r.URL)
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

	w.Header().Add(contentTypeHeader, htmlContentType)

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

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Split up this function into smaller functions to improve readability.
	if r.Header.Get(contentTypeHeader) != formContentType {
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

	raw_email := r.Form.Get("email")
	email, err := mail.ParseAddress(raw_email)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, write_err := w.Write([]byte(fmt.Sprintf("invalid e%v", err)))

		if write_err != nil {
			slog.Error(fmt.Sprintf("could not write response: %v", write_err))
		}

		return
	}

	password := r.Form.Get("password")
	hash, err := models.NewPasswordHash(password, bcrypt.DefaultCost)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("could not create password hash: %v", err))
		return
	}

	if s.userStore.EmailInUse(email.Address) {
		w.Header().Add(contentTypeHeader, htmlContentType)
		w.WriteHeader(http.StatusConflict)
		slog.Warn(fmt.Sprintf("email address %q is already in use", email.Address))
		return
	}

	// TODO: Change User to use email.Address instead of string?
	err = s.userStore.AddUser(email.Address, hash)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error(fmt.Sprintf("could not create user: %v", err))
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) getRegisterPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(contentTypeHeader, htmlContentType)

	body, err := s.renderer.RenderRegistrationPage()
	writeResponse(w, body, err, r.URL)
}
