package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func makeRequest(
	t *testing.T,
	method string,
	target string,
	body string,
	handler http.HandlerFunc,
) (*http.Response, string) {
	t.Helper()

	req := httptest.NewRequest(
		method,
		target,
		bytes.NewBufferString(body),
	)

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	handler(recorder, req)

	response := recorder.Result()
	t.Cleanup(func() {
		response.Body.Close()
	})

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	return response, string(responseBody)
}

func makePathValueRequest(
	t *testing.T,
	method string,
	target string,
	pathName string,
	pathValue string,
	handler http.HandlerFunc,
) (*http.Response, string) {
	t.Helper()

	req := httptest.NewRequest(
		method,
		target,
		nil,
	)

	// PathValue() only works when the request has been routed.
	// Set the path value manually for the test.
	req.SetPathValue(pathName, pathValue)

	recorder := httptest.NewRecorder()
	handler(recorder, req)

	response := recorder.Result()
	t.Cleanup(func() {
		response.Body.Close()
	})

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	return response, string(responseBody)
}

func TestChooseDirHandler(t *testing.T) {
	// Setup
	setUpTest(t)

	// Request
	doRequest := func(body string) (*http.Response, string) {
		return makeRequest(
			t,
			http.MethodPost,
			"/choose-dir",
			body,
			ChooseDirHandler,
		)
	}

	// Valid Path
	// First run initializes a new database, second one opens the existing one
	for range 2 {
		response, responseBody := doRequest(fmt.Sprintf(`{
		"dirPath": "%s"
	}`, testDirectoryPath))

		if response.StatusCode != http.StatusNoContent {
			t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusNoContent, response.StatusCode, responseBody)
		}
	}

	// Invalid Path
	response, responseBody := doRequest((`{"dirPath": "123123"}`))

	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusInternalServerError, response.StatusCode, responseBody)
	}

	// Invalid JSON
	response, responseBody = doRequest(`ascascascasc`)

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusBadRequest, response.StatusCode, responseBody)
	}

	// dirPath Missing
	response, responseBody = doRequest(`{
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
	response, responseBody := makeRequest(t, http.MethodGet, "/tags", "", ListTagsHandler)

	if response.StatusCode == http.StatusOK {
		var responseVals struct {
			Tags []Tag `json:"tags"`
		}

		err := json.Unmarshal([]byte(responseBody), &responseVals)
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
	doRequest := func(fileID string) (*http.Response, string) {
		return makePathValueRequest(
			t,
			http.MethodGet,
			"/tags/"+fileID,
			"file_id",
			fileID,
			TagsByFileIDHandler,
		)
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

	// Valid Request
	response, responseBody := doRequest("3")

	if response.StatusCode != http.StatusOK {
		t.Fatalf("Should've returned %d, instead got: %d", http.StatusOK, response.StatusCode)
	}

	responseTags := marshalResponse(responseBody)
	if len(responseTags) != 2 {
		t.Fatalf("Should have returned 2 elements, instead got: %v", responseTags)
	}

	// Invalid file ID
	for _, param := range testInvalidDbIDs {
		response, responseBody = doRequest(param)

		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("Should've returned with %d with %s, instead got: %d", http.StatusBadRequest, param, response.StatusCode)
		}
	}

	// Non-existent file ID
	response, responseBody = doRequest("123123")

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("Should've returned with %d, instead got: %d", http.StatusNotFound, response.StatusCode)
	}
}

func TestCreateTagHandler(t *testing.T) {
	// Setup
	setUpTest(t)

	// Set up DB
	setUpAndFillDB(t)
	defer CloseDB()

	// Request
	doRequest := func(body string) (*http.Response, string) {
		return makeRequest(
			t,
			http.MethodPost,
			"/create-tag",
			body,
			CreateTagHandler,
		)
	}

	// Valid Tag
	response, responseBody := doRequest(`{"name": "folk"}`)

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusCreated, response.StatusCode, responseBody)
	}

	// Duplicate Tag
	response, responseBody = doRequest(`{"name": "folk"}`)

	if response.StatusCode != http.StatusConflict {
		t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusConflict, response.StatusCode, responseBody)
	}

	// Invalid Names
	for _, name := range testInvalidTagNames {
		response, responseBody = doRequest(fmt.Sprintf(`{
		"name": "%s"
	}`, name))

		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusBadRequest, response.StatusCode, responseBody)
		}
	}

	// Missing name field
	response, responseBody = doRequest(`{"hello": "folk"}`)

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusBadRequest, response.StatusCode, responseBody)
	}
}

func TestRenameTagHandler(t *testing.T) {
	// Setup
	setUpTest(t)

	// Set up DB
	setUpAndFillDB(t)
	defer CloseDB()

	// Request
	doRequest := func(body string) (*http.Response, string) {
		return makeRequest(
			t,
			http.MethodPost,
			"/rename-tag",
			body,
			RenameTagHandler,
		)
	}

	// Valid
	// Twice to test updating the same tag to its own name
	for range 2 {
		response, responseBody := doRequest(`{"tagID": "3","newName": "folk"}`)

		if response.StatusCode != http.StatusNoContent {
			t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusNoContent, response.StatusCode, responseBody)
		}
	}

	// Non-existent Tag ID
	response, responseBody := doRequest(`{"tagID": "123123","newName": "trance"}`)

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusNotFound, response.StatusCode, responseBody)
	}

	// Updating another tag to same name
	response, responseBody = doRequest(`{"tagID": "2","newName": "folk"}`)

	if response.StatusCode != http.StatusConflict {
		t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusConflict, response.StatusCode, responseBody)
	}

	// Checking all the invalid values
	for _, dbID := range append(slices.Clone(testInvalidDbIDs), "2") {
		for _, tagName := range append(slices.Clone(testInvalidTagNames), "guitar") {
			// This is a valid one, skip it.
			if dbID == "2" && tagName == "guitar" {
				continue
			}

			response, responseBody = doRequest(fmt.Sprintf(`{
				"tagID": "%s",
				"newName": "%s"
			}`, dbID, tagName))

			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("Should have returned %d. Returned %d instead. tagID: %s name: %s Response: %v", http.StatusBadRequest, response.StatusCode, dbID, tagName, responseBody)
			}
		}
	}

	// Missing tagID
	response, responseBody = doRequest(`{"hello": "2","newName": "folk"}`)

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusBadRequest, response.StatusCode, responseBody)
	}

	// Missing newName
	response, responseBody = doRequest(`{"tagID": "2","hello": "folk"}`)

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("Should have returned %d. Returned %d instead. Response: %v", http.StatusBadRequest, response.StatusCode, responseBody)
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
	doRequest := func(fileID string) (*http.Response, string) {
		return makePathValueRequest(
			t,
			http.MethodGet,
			"/media/"+fileID,
			"file_id",
			fileID,
			MediaHandler,
		)
	}

	// Valid Request
	response, responseBody := doRequest("1")

	if response.StatusCode != http.StatusOK {
		t.Fatalf("Should've returned %d, instead got: %d", http.StatusOK, response.StatusCode)
	}

	if responseBody != testFileContents {
		t.Fatalf("expected %q, got %q", testFileContents, responseBody)
	}

	// Invalid file id
	for _, param := range testInvalidDbIDs {
		response, responseBody = doRequest(param)

		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("Should've returned with %d with %s, instead got: %d", http.StatusBadRequest, param, response.StatusCode)
		}
	}

	// File not found
	response, responseBody = doRequest("123123123")

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("Should've returned %d, instead got: %d", http.StatusNotFound, response.StatusCode)
	}
}
