package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPlaceholderHandlers(t *testing.T) {

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
			t.Fatalf(
				"expected status 200, got %d",
				response.StatusCode,
			)
		}

		if recorder.Body.String() != expected {
			t.Fatalf(
				"expected body %q, got %q",
				expected,
				recorder.Body.String(),
			)
		}
	}

	testHandler("/", "Hello, World!\n", HomeHandler)
	testHandler("/about", "About page\n", AboutHandler)
}

func TestMediaHandler(t *testing.T) {
	// Cleanup
	cleanTestDirectory(t)
	defer cleanTestDirectory(t)

	// Create a file
	createFile(t, "techno.mp3")

	// Initialize database
	if err := InitDatabase(testDirectoryPath); err != nil {
		t.Fatal(err)
	}
	defer CloseDB()

	// Test
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/media/1",
		nil,
	)

	// PathValue() only works when the request has been routed.
	// Set the path value manually for the test.
	req.SetPathValue("file_id", "1")

	rec := httptest.NewRecorder()

	MediaHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != testFileContents {
		t.Fatalf("expected %q, got %q", testFileContents, string(body))
	}
}
