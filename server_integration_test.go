package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	yatta "github.com/AnthonyDickson/yatta"
)

func TestCreateAndGetTodo(t *testing.T) {
	store := yatta.NewInMemoryTodoStore()
	server := yatta.NewServer(store)

	user := "Pierre"
	tasks := []string{"write a book", "philosophise"}

	for _, task := range tasks {
		server.ServeHTTP(httptest.NewRecorder(), newCreateTodosRequest(t, user, task))
	}

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newGetTodosRequest(t, user))

	assertStatus(t, response, http.StatusOK)
	assertResponseBody(t, response, tasks)
}
