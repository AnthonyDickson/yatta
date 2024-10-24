package main_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	yatta "github.com/AnthonyDickson/yatta"
)

type addTodoCall struct {
	user string
	task string
}

type StubTodoStore struct {
	store    map[string]string
	addCalls []addTodoCall
}

func (s *StubTodoStore) GetTodos(user string) string {
	return s.store[user]
}

func (s *StubTodoStore) AddTodo(user string, task string) {
	s.addCalls = append(s.addCalls, addTodoCall{user, task})
}

func TestGetTeapot(t *testing.T) {
	t.Run("returns status 418", func(t *testing.T) {
		server := yatta.NewServer(new(StubTodoStore))

		response := httptest.NewRecorder()
		request := newGetCoffeeRequest(t)

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusTeapot)
	})
}

func TestGetTodos(t *testing.T) {
	store := &StubTodoStore{
		store: map[string]string{
			"Alice": "send message to Bob",
			"thor":  "write more code",
		},
	}
	server := yatta.NewServer(store)

	t.Run("returns todos for Alice", func(t *testing.T) {
		request := newGetTodosRequest(t, "Alice")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusOK)

		want := "send message to Bob"
		assertResponseBody(t, response, want)

	})

	t.Run("returns todos for thor", func(t *testing.T) {
		request := newGetTodosRequest(t, "thor")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusOK)

		want := "write more code"
		assertResponseBody(t, response, want)
	})

	t.Run("returns 404 for nonexistent user", func(t *testing.T) {
		request := newGetTodosRequest(t, "Bob")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusNotFound)
	})
}

func TestCreateTodos(t *testing.T) {
	t.Run("creates todo on POST", func(t *testing.T) {
		store := &StubTodoStore{
			store: map[string]string{},
		}
		server := yatta.NewServer(store)

		want := addTodoCall{
			user: "Alice",
			task: "encrypt messages",
		}

		request := newTodosRequest(t, http.MethodPost, want.user, strings.NewReader(want.task))
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusAccepted)
		assertAddCalls(t, store, want)
	})
}

func assertAddCalls(t *testing.T, store *StubTodoStore, want addTodoCall) {
	t.Helper()

	if len(store.addCalls) != 1 {
		t.Fatalf("got %d calls to add todo, want %d", len(store.addCalls), 1)
	}

	got := store.addCalls[0]

	if got.user != want.user {
		t.Errorf("got call to add with user %q, want %q", got.user, want.user)
	}

	if got.task != want.task {
		t.Errorf("got call to add with task %q, want %q", got.task, want.task)
	}
}

func assertResponseBody(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	got := response.Body.String()

	if got != want {
		t.Errorf("got response body %q want %q", got, want)
	}
}

func newGetCoffeeRequest(t *testing.T) *http.Request {
	path := "/coffee"
	method := http.MethodGet
	request, err := http.NewRequest(method, path, nil)

	if err != nil {
		t.Fatalf("could not create request for path %s using method %s: %v", path, method, err)
	}

	return request
}

func newGetTodosRequest(t *testing.T, user string) *http.Request {
	t.Helper()

	return newTodosRequest(t, http.MethodGet, user, nil)
}

func newTodosRequest(t *testing.T, method string, user string, body io.Reader) *http.Request {
	t.Helper()

	path := fmt.Sprintf("/users/%s/todos", user)
	request, err := http.NewRequest(method, path, body)

	if err != nil {
		t.Fatalf("could not create request for path %s using method %s: %v", path, method, err)
	}

	return request
}

func assertStatus(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()

	got := response.Code

	if got != want {
		t.Errorf("got status %d want %d", got, want)
	}
}
