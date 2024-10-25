package main_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	yatta "github.com/AnthonyDickson/yatta"
)

const jsonContentType = "application/json"

type addTodoCall struct {
	user string
	task string
}

type StubTodoStore struct {
	store    map[string][]string
	addCalls []addTodoCall
}

func (s *StubTodoStore) GetTodos(user string) []string {
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
		store: map[string][]string{
			"Alice": {"send message to Bob"},
			"thor":  {"write more code"},
		},
	}
	server := yatta.NewServer(store)

	t.Run("returns todos for Alice", func(t *testing.T) {
		request := newGetTodosRequest(t, "Alice")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusOK)
		assertContentType(t, response, jsonContentType)
		assertResponseBody(t, response, []string{"send message to Bob"})

	})

	t.Run("returns todos for thor", func(t *testing.T) {
		request := newGetTodosRequest(t, "thor")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusOK)
		assertContentType(t, response, jsonContentType)
		assertResponseBody(t, response, []string{"write more code"})
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
			store: map[string][]string{},
		}
		server := yatta.NewServer(store)

		want := addTodoCall{
			user: "Alice",
			task: "encrypt messages",
		}

		request := newCreateTodosRequest(t, want.user, want.task)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusAccepted)
		assertAddCalls(t, store, []addTodoCall{want})
	})

	t.Run("create multiple todos", func(t *testing.T) {
		store := &StubTodoStore{
			store: map[string][]string{},
		}
		server := yatta.NewServer(store)

		cases := []addTodoCall{
			{"Thor", "write code"},
			{"Thor", "debug code"},
			{"Thor", "fix code"},
		}

		for _, want := range cases {
			request := newCreateTodosRequest(t, want.user, want.task)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			assertStatus(t, response, http.StatusAccepted)
		}

		assertAddCalls(t, store, cases)
	})
}

func assertContentType(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()

	got := response.Result().Header.Get("content-type")

	if got != want {
		t.Errorf("got header content-type %q want %q", got, want)
	}
}

func assertAddCalls(t *testing.T, store *StubTodoStore, wantCalls []addTodoCall) {
	t.Helper()

	if len(store.addCalls) != len(wantCalls) {
		t.Fatalf("got %d calls to add todo, want %d", len(store.addCalls), len(wantCalls))
	}

	for i := 0; i < len(wantCalls); i++ {
		got := store.addCalls[i]
		want := wantCalls[i]

		if got.user != want.user {
			t.Errorf("got call to add with user %q, want %q", got.user, want.user)
		}

		if got.task != want.task {
			t.Errorf("got call to add with task %q, want %q", got.task, want.task)
		}
	}
}

func assertResponseBody(t *testing.T, response *httptest.ResponseRecorder, want []string) {
	t.Helper()

	var got []string
	err := json.NewDecoder(response.Body).Decode(&got)

	if err != nil {
		t.Fatalf("a problem ocurred while decoding the response body as JSON: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
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

func newCreateTodosRequest(t *testing.T, user string, task string) *http.Request {
	t.Helper()

	return newTodosRequest(t, http.MethodPost, user, strings.NewReader(task))
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
