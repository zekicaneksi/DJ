package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlers(t *testing.T) {

	testHandler := func(route string, expected string, Handler func(http.ResponseWriter, *http.Request)) {
		req := httptest.NewRequest(
			http.MethodGet,
			route,
			nil,
		)

		recorder := httptest.NewRecorder()

		Handler(recorder, req)

		response := recorder.Result()

		if response.StatusCode != http.StatusOK {
			t.Errorf(
				"expected status 200, got %d",
				response.StatusCode,
			)
		}

		if recorder.Body.String() != expected {
			t.Errorf(
				"expected body %q, got %q",
				expected,
				recorder.Body.String(),
			)
		}
	}

	testHandler("/", "Hello, World!\n", HomeHandler)
	testHandler("/about", "About page\n", AboutHandler)
}
