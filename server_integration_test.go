package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	yatta "github.com/AnthonyDickson/yatta"
)

func TestCreateAndGetTasks(t *testing.T) {
	database, cleanup := createTempFile(t, "")
	defer cleanup()

	store := yatta.NewFileTaskStore(database)
	renderer := mustCreateRenderer(t)
	server := mustCreateServer(t, store, renderer)
	user := "Pierre"
	tasks := []yatta.Task{{ID: 0, Description: "write a book"}, {ID: 0, Description: "philosophise"}}

	for _, task := range tasks {
		server.ServeHTTP(httptest.NewRecorder(), newCreateTasksRequest(t, user, task.Description))
	}

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newGetTasksRequest(t, user))

	assertStatus(t, response, http.StatusOK)
	assertHTMLContainsTasks(t, response.Body.String(), tasks, "li")
}
