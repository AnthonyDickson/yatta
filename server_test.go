package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	yatta "github.com/AnthonyDickson/yatta"
)

func TestGetTeapot(t *testing.T) {
	t.Run("returns status 418", func(t *testing.T) {
		server := yatta.NewServer()

		response := httptest.NewRecorder()
		request := newGetCoffeeRequest(t)

		server.ServeHTTP(response, request)

		assertStatus(t, response, http.StatusTeapot)
	})
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

func assertStatus(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()

	got := response.Code

	if got != want {
		t.Errorf("got status %d want %d", got, want)
	}
}
