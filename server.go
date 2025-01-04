package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/mail"
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

type handlerError struct {
	status int
	err    error
}

func (h handlerError) Error() string {
	return fmt.Sprintf("HTTP status code %d: %v", h.status, h.err)
}

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
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("could not get users: %v", err))
		return
	}

	body, err := s.renderer.RenderIndex(users)
	writeRendererResponse(w, r, body, err)
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
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("could not get task with ID %q: %v", id, err))
		return
	}

	if task == nil {
		http.NotFound(w, r)
		return
	}

	body, err := s.renderer.RenderTask(*task)
	writeRendererResponse(w, r, body, err)
}

func (s *Server) getTasks(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	tasks, err := s.taskStore.GetTasks(user)

	if err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("could not get tasks for user %q: %v", user, err))
		return
	}

	if tasks == nil {
		http.NotFound(w, r)
		return
	}

	body, err := s.renderer.RenderTaskList(tasks)
	writeRendererResponse(w, r, body, err)
}

func (s *Server) addTask(w http.ResponseWriter, r *http.Request) {
	user := r.PathValue("user")
	bodyBytes, err := io.ReadAll(r.Body)

	if err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("could not read the request body: %v", err))
		return
	}

	task := string(bodyBytes)
	err = s.taskStore.AddTask(user, task)

	if err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("could not add task %q for user %q: %v", user, task, err))
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	raw_email, password, e := parseRegistrationFrom(r)

	if e != nil {
		writeError(w, r, e.status, fmt.Sprintf("could not parse registration form: %v", e.err))
		return
	}

	email, err := mail.ParseAddress(raw_email)

	if err != nil {
		writeReponse(w, http.StatusUnprocessableEntity, fmt.Sprintf("invalid email: %v", err))
		return
	}

	hash, err := models.NewPasswordHash(password, bcrypt.DefaultCost)

	if err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("could not create password hash: %v", err))
		return
	}

	if s.userStore.EmailInUse(email.Address) {
		writeReponse(w, http.StatusConflict, fmt.Sprintf("email %q is already in use", email.Address))
		return
	}

	// TODO: Change User to use email.Address instead of string?
	err = s.userStore.AddUser(email.Address, hash)

	if err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("could not create user: %v", err))
		return
	}

	writeReponse(w, http.StatusAccepted, fmt.Sprintf("user %q created", email.Address))
}

func (s *Server) getRegisterPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(contentTypeHeader, htmlContentType)
	body, err := s.renderer.RenderRegistrationPage()
	writeRendererResponse(w, r, body, err)
}

func parseRegistrationFrom(r *http.Request) (email string, password string, err *handlerError) {
	if r.Header.Get(contentTypeHeader) != formContentType {
		return "", "", &handlerError{
			status: http.StatusUnsupportedMediaType,
			err:    fmt.Errorf("content type %q not supported", r.Header.Get(contentTypeHeader)),
		}
	}

	e := r.ParseForm()

	if e != nil {
		return "", "", &handlerError{status: http.StatusBadRequest, err: e}
	}

	if !r.Form.Has("email") || !r.Form.Has("password") {
		return "", "", &handlerError{status: http.StatusBadRequest, err: errors.New("email or password not set in form")}
	}

	email = r.Form.Get("email")
	password = r.Form.Get("password")

	return email, password, nil
}

func writeReponse(w http.ResponseWriter, status int, body string) {
	w.Header().Add(contentTypeHeader, htmlContentType)
	w.WriteHeader(status)

	if _, err := w.Write([]byte(body)); err != nil {
		slog.Error(fmt.Sprintf("could not write response: %v", err))
	}
}

func writeError(w http.ResponseWriter, r *http.Request, status int, message string) {
	w.WriteHeader(status)
	slog.Error(fmt.Sprintf("[%s] %d: %s", r.URL, status, message))
}

func writeRendererResponse(w http.ResponseWriter, r *http.Request, body []byte, err error) {
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("an error occurred while rendering the template: %v", err))
		return
	}

	w.Header().Add(contentTypeHeader, htmlContentType)

	if _, err := w.Write(body); err != nil {
		slog.Error(fmt.Sprintf("an error occurred while writing the response body: %v", err))
	}
}
