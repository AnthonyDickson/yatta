package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AnthonyDickson/yatta/models"
	"github.com/AnthonyDickson/yatta/stores"
	"github.com/AnthonyDickson/yatta/yattatest"
)

func TestCreateAndGetTasks(t *testing.T) {
	taskStore, cleanupTaskDatabase := mustCreateFileTaskStore(t, "")
	defer cleanupTaskDatabase()

	userStore, cleanupUserDatabase := mustCreateFileUserStore(t, "")
	defer cleanupUserDatabase()

	renderer := mustCreateRenderer(t)
	server := mustCreateServer(t, taskStore, userStore, renderer)

	email := "Pierre.Joseph@Proudhorn.fr"
	want_user := models.User{ID: 1, Email: email, Password: "propertyistheft"}
	want_tasks := []models.Task{{ID: 0, Description: "write a book"}, {ID: 0, Description: "Become ungovernable."}}

	server.ServeHTTP(httptest.NewRecorder(), newCreateUserRequest(t, createUserCall{
		Email:    want_user.Email,
		Password: want_user.Password,
	}))

	for _, task := range want_tasks {
		server.ServeHTTP(httptest.NewRecorder(), newCreateTasksRequest(t, email, task.Description))
	}

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newGetTasksRequest(t, email))

	assertStatus(t, response, http.StatusOK)
	assertHTMLContainsTasks(t, response.Body.String(), want_tasks, "li")
	assertStoreHasUser(t, userStore, want_user)
}

func mustCreateFileUserStore(t *testing.T, initialData string) (*stores.FileUserStore, func()) {
	t.Helper()

	database, cleanup := yattatest.CreateTempFile(t, initialData)

	store, err := stores.NewFileUserStore(database)

	if err != nil {
		t.Fatalf("could not load user store: %v", err)
	}

	return store, cleanup
}

func mustCreateFileTaskStore(t *testing.T, initialData string) (*stores.FileTaskStore, func()) {
	t.Helper()

	database, cleanup := yattatest.CreateTempFile(t, initialData)

	store, err := stores.NewFileTaskStore(database)

	if err != nil {
		t.Fatalf("could not load task store: %v", err)
	}

	return store, cleanup
}

func assertStoreHasUser(t *testing.T, store *stores.FileUserStore, want models.User) {
	t.Helper()

	got_user, err := store.GetUser(want.ID)
	yattatest.AssertNoError(t, err)

	if got_user == nil {
		t.Fatalf("got nil user, want %v", want)
	}

	if *got_user != want {
		t.Errorf("got user %v, want %v", *got_user, want)
	}
}
