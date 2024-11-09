package main_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	yatta "github.com/AnthonyDickson/yatta"
)

const htmlContentType = "text/html"

type addTodoCall struct {
	user string
	task string
}

type getTodosCall struct {
	user  string
	tasks []yatta.Task
}

type StubTodoStore struct {
	store         map[string][]yatta.Task
	addCalls      []addTodoCall
	getTodosCalls []getTodosCall
}

func (s *StubTodoStore) GetTodos(user string) ([]yatta.Task, error) {
	tasks := s.store[user]

	s.getTodosCalls = append(s.getTodosCalls, getTodosCall{user, tasks})

	return tasks, nil
}

type SpyRenderer struct {
	renderTodosCalls [][]yatta.Task
}

func (s *SpyRenderer) RenderTodosList(todos []yatta.Task) ([]byte, error) {
	s.renderTodosCalls = append(s.renderTodosCalls, todos)

	return nil, nil
}

func (s *StubTodoStore) AddTodo(user string, task string) error {
	s.addCalls = append(s.addCalls, addTodoCall{user, task})

	return nil
}

func TestGetCoffee(t *testing.T) {
	t.Run("returns status 418", func(t *testing.T) {
		server := mustCreateServer(t, new(StubTodoStore), new(SpyRenderer))

		response := httptest.NewRecorder()
		request := newGetCoffeeRequest(t)

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusTeapot)
	})
}

func TestGetTodos(t *testing.T) {
	getStoreRendererAndServer := func() (*StubTodoStore, *SpyRenderer, *yatta.Server) {
		store := &StubTodoStore{
			store: map[string][]yatta.Task{
				"Alice": {{ID: 0, Description: "send message to Bob"}},
				"thor":  {{ID: 1, Description: "write more code"}},
			},
		}

		renderer := new(SpyRenderer)

		server := mustCreateServer(t, store, renderer)
		return store, renderer, server
	}

	t.Run("returns todos for Alice", func(t *testing.T) {
		store, renderer, server := getStoreRendererAndServer()
		want := getTodosCall{"Alice", []yatta.Task{{ID: 0, Description: "send message to Bob"}}}

		request := newGetTodosRequest(t, want.user)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusOK)
		assertContentType(t, response, htmlContentType)
		assertGetTodosCall(t, store, want)
		assertRenderTodosCall(t, renderer, want.tasks)
	})

	t.Run("returns todos for thor", func(t *testing.T) {
		store, renderer, server := getStoreRendererAndServer()
		want := getTodosCall{"thor", []yatta.Task{{ID: 1, Description: "write more code"}}}

		request := newGetTodosRequest(t, want.user)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusOK)
		assertContentType(t, response, htmlContentType)
		assertGetTodosCall(t, store, want)
		assertRenderTodosCall(t, renderer, want.tasks)
	})

	t.Run("returns 404 for nonexistent user", func(t *testing.T) {
		_, _, server := getStoreRendererAndServer()
		request := newGetTodosRequest(t, "Bob")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusNotFound)
	})
}

func TestCreateTodos(t *testing.T) {
	t.Run("creates todo on POST", func(t *testing.T) {
		store := &StubTodoStore{
			store: map[string][]yatta.Task{},
		}
		server := mustCreateServer(t, store, new(SpyRenderer))

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
			store: map[string][]yatta.Task{},
		}
		server := mustCreateServer(t, store, new(SpyRenderer))

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

func mustCreateServer(t *testing.T, store yatta.TodoStore, renderer yatta.Renderer) *yatta.Server {
	t.Helper()

	server, err := yatta.NewServer(store, renderer)

	if err != nil {
		t.Errorf("an ocurred while creating the server: %v", err)
	}

	return server
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

func assertGetTodosCall(t *testing.T, store *StubTodoStore, want getTodosCall) {
	t.Helper()

	calls := store.getTodosCalls

	if len(calls) != 1 {
		t.Errorf("got %d call(s) to GetTodos want 1", len(calls))
	}

	got := calls[0]

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got calls to GetTodos %q, want %q", got, want)
	}
}

func assertRenderTodosCall(t *testing.T, renderer *SpyRenderer, want []yatta.Task) {
	t.Helper()

	if len(renderer.renderTodosCalls) != 1 {
		t.Fatalf("got %d calls to RenderTodosList, want 1", len(renderer.renderTodosCalls))
	}

	got := renderer.renderTodosCalls[0]

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got calls to RenderTodosList %q, want %q", got, want)
	}
}
