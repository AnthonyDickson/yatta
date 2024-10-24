package main_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	yatta "github.com/AnthonyDickson/yatta"
)

type StubTodoStore struct {
	store map[string]string
}

func (s *StubTodoStore) GetTodos(user string) string {
	return s.store[user]
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
	path := fmt.Sprintf("/users/%s/todos", user)
	method := http.MethodGet
	request, err := http.NewRequest(method, path, nil)

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
