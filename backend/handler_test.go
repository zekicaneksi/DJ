package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChooseDirHandler(t *testing.T) {
	// Setup
	setUpTest(t)

	// Request
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
	// First run initializes a new database, second one opens the existing one
	for range 2 {
		response, responseBody := testHandler(fmt.Sprintf(`{
		"dirPath": "%s"
	}`, testDirectoryPath))

		if response.StatusCode != http.StatusNoContent {
			t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusNoContent, response.StatusCode, responseBody)
		}
	}

	// Invalid Path
	response, responseBody := testHandler(fmt.Sprintf(`{
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

func TestListTagsHandler(t *testing.T) {
	// Setup
	setUpTest(t)

	// Set up DB
	setUpAndFillDB(t)
	defer CloseDB()

	// Request
	req := httptest.NewRequest(
		http.MethodGet,
		"/tags",
		nil,
	)

	recorder := httptest.NewRecorder()
	ListTagsHandler(recorder, req)

	response := recorder.Result()
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode == http.StatusOK {
		var responseVals struct {
			Tags []Tag `json:"tags"`
		}

		err = json.Unmarshal([]byte(string(responseBody)), &responseVals)
		if err != nil {
			t.Fatalf("cannot unmarshal %s: %v", string(responseBody), err)
		}

		if len(responseVals.Tags) != 4 {
			t.Fatalf("Should have returned 4 elements, instead got: %v", responseVals.Tags)
		}
	} else {
		t.Fatalf("Should've returned %d, instead got: %d", http.StatusOK, response.StatusCode)
	}
}

func TestTagsByFileIDHandler(t *testing.T) {
	// Setup
	setUpTest(t)

	// Set up DB
	setUpAndFillDB(t)
	defer CloseDB()

	// Request
	makeRequest := func(file_id string) (*http.Response, string) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/tags/"+file_id,
			nil,
		)

		// PathValue() only works when the request has been routed.
		// Set the path value manually for the test.
		req.SetPathValue("file_id", file_id)

		rec := httptest.NewRecorder()
		TagsByFileIDHandler(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}

		return res, string(body)
	}

	marshalResponse := func(stringResponse string) []Tag {
		var responseVals struct {
			Tags []Tag `json:"tags"`
		}

		err := json.Unmarshal([]byte(stringResponse), &responseVals)
		if err != nil {
			t.Fatalf("cannot unmarshal %s: %v", stringResponse, err)
		}

		return responseVals.Tags
	}

	var responseTags []Tag
	// Valid Request
	response, responseBody := makeRequest("3")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Should've returned %d, instead got: %d", http.StatusOK, response.StatusCode)
	}
	responseTags = marshalResponse(responseBody)
	if len(responseTags) != 2 {
		t.Fatalf("Should have returned 2 elements, instead got: %v", responseTags)
	}

	// Invalid file ID
	for _, param := range invalidDbIDs {
		response, responseBody = makeRequest(param)
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("Should've returned with %d with %s, instead got: %d", http.StatusBadRequest, param, response.StatusCode)
		}
	}

	// Missing file ID
	response, responseBody = makeRequest("123123")
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("Should've returned with %d, instead got: %d", http.StatusNotFound, response.StatusCode)
	}
}

func TestMediaHandler(t *testing.T) {
	// Setup
	setUpTest(t)

	// Create a file
	createFile(t, "techno.mp3")

	// Initialize database
	if err := InitDatabase(testDirectoryPath); err != nil {
		t.Fatal(err)
	}
	defer CloseDB()

	// Request
	makeRequest := func(file_id string) (*http.Response, string) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/media/"+file_id,
			nil,
		)

		// PathValue() only works when the request has been routed.
		// Set the path value manually for the test.
		req.SetPathValue("file_id", file_id)

		rec := httptest.NewRecorder()

		MediaHandler(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}

		return res, string(body)
	}

	// Valid Request
	response, responseBody := makeRequest("1")
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Should've returned %d, instead got: %d", http.StatusOK, response.StatusCode)
	}
	if responseBody != testFileContents {
		t.Fatalf("expected %q, got %q", testFileContents, responseBody)
	}

	// Invalid file id
	for _, param := range invalidDbIDs {
		response, responseBody = makeRequest(param)
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("Should've returned with %d with %s, instead got: %d", http.StatusBadRequest, param, response.StatusCode)
		}
	}

	// File not found
	response, responseBody = makeRequest("123123123")
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("Should've returned %d, instead got: %d", http.StatusNotFound, response.StatusCode)
	}
}
