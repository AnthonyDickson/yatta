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
	"github.com/AnthonyDickson/yatta/models"
	"github.com/AnthonyDickson/yatta/stores"
)

const htmlContentType = "text/html"

type addTaskCall struct {
	user string
	task string
}

type getTasksCall struct {
	user  string
	tasks []models.Task
}

type getTaskCall struct {
	id   uint64
	task *models.Task
}

type StubTaskStore struct {
	store         map[string][]models.Task
	addCalls      []addTaskCall
	getTasksCalls []getTasksCall
	getTaskCalls  []getTaskCall
}

func (s *StubTaskStore) GetTasks(user string) ([]models.Task, error) {
	tasks := s.store[user]

	s.getTasksCalls = append(s.getTasksCalls, getTasksCall{user, tasks})

	return tasks, nil
}

func (s *StubTaskStore) GetTask(id uint64) (*models.Task, error) {
	for _, tasks := range s.store {
		for _, task := range tasks {
			if task.ID == id {
				s.getTaskCalls = append(s.getTaskCalls, getTaskCall{id: id, task: &task})
				return &task, nil
			}
		}
	}

	s.getTaskCalls = append(s.getTaskCalls, getTaskCall{id: id, task: nil})
	return nil, nil
}

type SpyRenderer struct {
	renderTasksCalls [][]models.Task
	renderTaskCalls  []models.Task
}

func (s *SpyRenderer) RenderTaskList(tasks []models.Task) ([]byte, error) {
	s.renderTasksCalls = append(s.renderTasksCalls, tasks)

	return nil, nil
}

func (s *SpyRenderer) RenderTask(task models.Task) ([]byte, error) {
	s.renderTaskCalls = append(s.renderTaskCalls, task)

	return nil, nil
}

func (s *StubTaskStore) AddTask(user string, task string) error {
	s.addCalls = append(s.addCalls, addTaskCall{user, task})

	return nil
}

type DummyUserStore struct{}

func (d *DummyUserStore) CreateUser(email, password string) error {
	return nil
}

func TestGetCoffee(t *testing.T) {
	t.Run("returns status 418", func(t *testing.T) {
		server := mustCreateServer(t, new(StubTaskStore), new(DummyUserStore), new(SpyRenderer))

		response := httptest.NewRecorder()
		request := newGetCoffeeRequest(t)

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusTeapot)
	})
}

func TestGetTasks(t *testing.T) {
	getStoreRendererAndServer := func() (*StubTaskStore, *SpyRenderer, *yatta.Server) {
		store := &StubTaskStore{
			store: map[string][]models.Task{
				"Alice": {{ID: 0, Description: "send message to Bob"}},
				"thor":  {{ID: 1, Description: "write more code"}},
			},
		}

		renderer := new(SpyRenderer)

		server := mustCreateServer(t, store, new(DummyUserStore), renderer)
		return store, renderer, server
	}

	t.Run("returns tasks for Alice", func(t *testing.T) {
		store, renderer, server := getStoreRendererAndServer()
		want := getTasksCall{"Alice", []models.Task{{ID: 0, Description: "send message to Bob"}}}

		request := newGetTasksRequest(t, want.user)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusOK)
		assertContentType(t, response, htmlContentType)
		assertGetTasksCall(t, store, want)
		assertRenderTasksCall(t, renderer, want.tasks)
	})

	t.Run("returns tasks for thor", func(t *testing.T) {
		store, renderer, server := getStoreRendererAndServer()
		want := getTasksCall{"thor", []models.Task{{ID: 1, Description: "write more code"}}}

		request := newGetTasksRequest(t, want.user)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusOK)
		assertContentType(t, response, htmlContentType)
		assertGetTasksCall(t, store, want)
		assertRenderTasksCall(t, renderer, want.tasks)
	})

	t.Run("returns 404 for nonexistent user", func(t *testing.T) {
		_, _, server := getStoreRendererAndServer()
		request := newGetTasksRequest(t, "Bob")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusNotFound)
	})
}

func TestGetTask(t *testing.T) {
	t.Run("get task by id", func(t *testing.T) {
		want := []models.Task{
			{ID: 0, Description: "write more tasks"},
			{ID: 1, Description: "stop writing too many tasks"},
		}

		store := &StubTaskStore{
			store: map[string][]models.Task{
				"Alice": want,
			},
		}
		renderer := new(SpyRenderer)
		server := mustCreateServer(t, store, new(DummyUserStore), renderer)

		request, _ := http.NewRequest(http.MethodGet, "/tasks/0", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusOK)
		assertContentType(t, response, htmlContentType)
		assertGetTaskCall(t, store, want[0])
		assertRenderTaskCall(t, renderer, want[0])
	})

	t.Run("get task by invalid ID returns 404 not found", func(t *testing.T) {
		store := &StubTaskStore{
			store: map[string][]models.Task{
				"Alice": {{ID: 0, Description: "find my todos list"}},
			},
		}

		renderer := new(SpyRenderer)
		server := mustCreateServer(t, store, new(DummyUserStore), renderer)

		request, _ := http.NewRequest(http.MethodGet, "/tasks/8", nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusNotFound)
	})
}

func TestCreateTasks(t *testing.T) {
	t.Run("creates task on POST", func(t *testing.T) {
		store := &StubTaskStore{
			store: map[string][]models.Task{},
		}
		server := mustCreateServer(t, store, new(DummyUserStore), new(SpyRenderer))

		want := addTaskCall{
			user: "Alice",
			task: "encrypt messages",
		}

		request := newCreateTasksRequest(t, want.user, want.task)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusAccepted)
		assertAddCalls(t, store, []addTaskCall{want})
	})

	t.Run("create multiple tasks", func(t *testing.T) {
		store := &StubTaskStore{
			store: map[string][]models.Task{},
		}
		server := mustCreateServer(t, store, new(DummyUserStore), new(SpyRenderer))

		cases := []addTaskCall{
			{"Thor", "write code"},
			{"Thor", "debug code"},
			{"Thor", "fix code"},
		}

		for _, want := range cases {
			request := newCreateTasksRequest(t, want.user, want.task)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			assertStatus(t, response, http.StatusAccepted)
		}

		assertAddCalls(t, store, cases)
	})
}

const formContentType = "application/x-www-form-urlencoded"

type DummyTaskStore struct{}

func (d *DummyTaskStore) GetTask(id uint64) (*models.Task, error) {
	return nil, nil
}

func (d *DummyTaskStore) GetTasks(user string) ([]models.Task, error) {
	return nil, nil
}

func (d *DummyTaskStore) AddTask(user string, description string) error {
	return nil
}

type DummyRenderer struct{}

func (d *DummyRenderer) RenderTask(task models.Task) ([]byte, error) {
	return nil, nil
}

func (d *DummyRenderer) RenderTaskList(tasks []models.Task) ([]byte, error) {
	return nil, nil
}

type createUserCall struct {
	Email    string
	Password string
}

type SpyUserStore struct {
	createUserCalls []createUserCall
}

func (s *SpyUserStore) CreateUser(email, password string) error {
	s.createUserCalls = append(s.createUserCalls, createUserCall{email, password})
	return nil
}

func TestCreateUser(t *testing.T) {
	t.Run("can create a new user", func(t *testing.T) {
		cases := []createUserCall{
			{"test@test.com", "hunter2"},
			{"foo@bar.com", "baz"},
		}

		for _, want := range cases {
			store := new(SpyUserStore)

			server := mustCreateServer(t, new(DummyTaskStore), store, new(DummyRenderer))
			request := newCreateUserRequest(t, want)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			assertStatus(t, response, http.StatusAccepted)
			assertCreateUserCalls(t, store, want)
		}
	})

	server := mustCreateServer(t, new(DummyTaskStore), new(DummyUserStore), new(DummyRenderer))

	t.Run("wrong content type returns HTTP status unsupported media type", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/users", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusUnsupportedMediaType)
	})

	t.Run("invalid form returns HTTP status bad request", func(t *testing.T) {
		cases := []io.Reader{
			nil,
			strings.NewReader(""),
			strings.NewReader("email="),
			strings.NewReader("password="),
		}

		for _, requestBody := range cases {
			request := httptest.NewRequest(http.MethodPost, "/users", requestBody)
			request.Header.Add("Content-Type", formContentType)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			assertStatus(t, response, http.StatusBadRequest)
		}
	})

	// TODO: validate emails for formatting and uniqueness, and passwords for strength.
}

func newCreateUserRequest(t *testing.T, user createUserCall) *http.Request {
	t.Helper()

	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(fmt.Sprintf("email=%s&password=%s", user.Email, user.Password)))
	request.Header.Add("Content-Type", formContentType)

	return request
}

func assertCreateUserCalls(t *testing.T, store *SpyUserStore, want createUserCall) {
	t.Helper()

	if len(store.createUserCalls) != 1 {
		t.Fatalf("got %d calls to CreateUser, want 1", len(store.createUserCalls))
	}

	got := store.createUserCalls[0]

	if got != want {
		t.Errorf("got call to CreateUser with arguments %v, want %v", got, want)
	}
}

func mustCreateServer(t *testing.T, taskStore stores.TaskStore, userStore stores.UserStore, renderer yatta.Renderer) *yatta.Server {
	t.Helper()

	server, err := yatta.NewServer(taskStore, userStore, renderer)

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

func newGetTasksRequest(t *testing.T, user string) *http.Request {
	t.Helper()

	return newTasksRequest(t, http.MethodGet, user, nil)
}

func newCreateTasksRequest(t *testing.T, user string, task string) *http.Request {
	t.Helper()

	return newTasksRequest(t, http.MethodPost, user, strings.NewReader(task))
}

func newTasksRequest(t *testing.T, method string, user string, body io.Reader) *http.Request {
	t.Helper()

	path := fmt.Sprintf("/users/%s/tasks", user)
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

func assertAddCalls(t *testing.T, store *StubTaskStore, wantCalls []addTaskCall) {
	t.Helper()

	if len(store.addCalls) != len(wantCalls) {
		t.Fatalf("got %d calls to add task, want %d", len(store.addCalls), len(wantCalls))
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

func assertGetTasksCall(t *testing.T, store *StubTaskStore, want getTasksCall) {
	t.Helper()

	calls := store.getTasksCalls

	if len(calls) != 1 {
		t.Errorf("got %d call(s) to GetTasks want 1", len(calls))
	}

	got := calls[0]

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got calls to GetTasks %q, want %q", got, want)
	}
}

func assertRenderTasksCall(t *testing.T, renderer *SpyRenderer, want []models.Task) {
	t.Helper()

	if len(renderer.renderTasksCalls) != 1 {
		t.Fatalf("got %d calls to RenderTasksList, want 1", len(renderer.renderTasksCalls))
	}

	got := renderer.renderTasksCalls[0]

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got calls to RenderTasksList %q, want %q", got, want)
	}
}

func assertGetTaskCall(t *testing.T, store *StubTaskStore, want models.Task) {
	t.Helper()

	if len(store.getTaskCalls) != 1 {
		t.Fatalf("got %d calls to GetTask, want 1", len(store.getTaskCalls))
	}

	got := store.getTaskCalls[0].task

	if got == nil {
		t.Errorf("got nil task, want %v", want)
	} else if *got != want {
		t.Errorf("got task %v want %v", *got, want)
	}
}

func assertRenderTaskCall(t *testing.T, renderer *SpyRenderer, want models.Task) {
	t.Helper()

	if len(renderer.renderTaskCalls) != 1 {
		t.Fatalf("got %d calls to RenderTask, want 1", len(renderer.renderTaskCalls))
	}

	got := renderer.renderTaskCalls[0]

	if got != want {
		t.Errorf("got call to RenderTask with task %q, want call with task %q", got, want)
	}
}
