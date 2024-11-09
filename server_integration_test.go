package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	yatta "github.com/AnthonyDickson/yatta"
)

func TestCreateAndGetTodo(t *testing.T) {
	database, cleanup := createTempFile(t, "")
	defer cleanup()

	store := yatta.NewFileTodoStore(database)
	renderer := mustCreateRenderer(t)
	server := mustCreateServer(t, store, renderer)
	user := "Pierre"
	tasks := []yatta.Task{{ID: 0, Description: "write a book"}, {ID: 0, Description: "philosophise"}}

	for _, task := range tasks {
		server.ServeHTTP(httptest.NewRecorder(), newCreateTodosRequest(t, user, task.Description))
	}

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newGetTodosRequest(t, user))

	assertStatus(t, response, http.StatusOK)
	assertHTMLContainsTodos(t, response.Body.String(), tasks)
}
