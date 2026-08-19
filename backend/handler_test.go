package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChooseDirHandler(t *testing.T) {

	testHandler := func(body string) (*http.Response, string) {
		req := httptest.NewRequest(
			http.MethodPost,
			"/choose-dir",
			bytes.NewBufferString(body),
		)
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		ChooseDirHandler(recorder, req)

		response := recorder.Result()
		defer response.Body.Close()

		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}

		return response, string(responseBody)
	}

	// Valid Path
	response, responseBody := testHandler(fmt.Sprintf(`{
		"dirPath": "%s"
	}`, testDirectoryPath))

	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusNoContent, response.StatusCode, responseBody)
	}

	// Invalid Path
	response, responseBody = testHandler(fmt.Sprintf(`{
		"dirPath": "%s"
	}`, "123123"))

	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusInternalServerError, response.StatusCode, responseBody)
	}

	// Invalid JSON
	response, responseBody = testHandler(`ascascascasc`)

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusBadRequest, response.StatusCode, responseBody)
	}

	// dirPath Missing
	response, responseBody = testHandler(`{
		"message": "Hello"
	}`)

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusBadRequest, response.StatusCode, responseBody)
	}

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
