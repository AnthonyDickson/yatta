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
	raw_password := "propertyistheft"

	want_user := createUserRequestData{
		Email:    email,
		Password: raw_password,
	}
	want_tasks := []models.Task{{ID: 0, Description: "write a book"}, {ID: 1, Description: "Become ungovernable."}}

	server.ServeHTTP(httptest.NewRecorder(), newCreateUserRequest(t, want_user))

	for _, task := range want_tasks {
		server.ServeHTTP(httptest.NewRecorder(), newCreateTasksRequest(t, email, task.Description))
	}

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newGetTasksRequest(t, email))

	assertStatus(t, response, http.StatusOK)
	assertHTMLContainsTasks(t, response.Body.String(), want_tasks, "li")
	assertStoreHasUser(t, userStore, 1, want_user)
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

func assertStoreHasUser(t *testing.T, store *stores.FileUserStore, want_id uint64, want createUserRequestData) {
	t.Helper()

	got_user, err := store.GetUser(want_id)
	yattatest.AssertNoError(t, err)

	if got_user == nil {
		t.Fatalf("got nil user, want %v", want)
	}

	if got_user.Email != want.Email {
		t.Errorf("got email %v, want %v", got_user.Email, want.Email)
	}

	if got_user.Password.Compare(want.Password) != nil {
		t.Errorf("hash %s does not verify password %q", got_user.Password, want.Password)
	}
}
