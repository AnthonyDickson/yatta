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
	database, cleanup := yattatest.CreateTempFile(t, "")
	defer cleanup()

	taskStore := stores.NewFileTaskStore(database)
	// userStore := yatta.NewFileUserStore(database)
	renderer := mustCreateRenderer(t)
	// TODO: Replace dummy user store with real one
	server := mustCreateServer(t, taskStore, new(DummyUserStore), renderer)
	user := "Pierre"
	tasks := []models.Task{{ID: 0, Description: "write a book"}, {ID: 0, Description: "philosophise"}}

	for _, task := range tasks {
		server.ServeHTTP(httptest.NewRecorder(), newCreateTasksRequest(t, user, task.Description))
	}

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newGetTasksRequest(t, user))

	assertStatus(t, response, http.StatusOK)
	assertHTMLContainsTasks(t, response.Body.String(), tasks, "li")
}
