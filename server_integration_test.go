package main_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	yatta "github.com/AnthonyDickson/yatta"
)

func TestCreateAndGetTodo(t *testing.T) {
	store := yatta.NewInMemoryTodoStore()
	server := yatta.NewServer(store)

	server.ServeHTTP(httptest.NewRecorder(), newTodosRequest(t, http.MethodPost, "Pierre", strings.NewReader("write a book")))

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newGetTodosRequest(t, "Pierre"))

	assertStatus(t, response, http.StatusOK)
	assertResponseBody(t, response, "write a book")
}
