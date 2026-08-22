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

// Makes an http request to a given handler
// Throws error if expected http status code isn't returned
// Returns the response and its body
func makeRequest(
	t *testing.T,
	method string,
	target string,
	body string,
	handler http.HandlerFunc,
	expectedStatusCode int,
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

	if response.StatusCode != expectedStatusCode {
		t.Fatalf(`Should have returned %d.
		Returned %d instead.
		Sent body: %v,
		Response: %v`,
			expectedStatusCode, response.StatusCode, body, string(responseBody))
	}

	return response, string(responseBody)
}

// Same thing with makeRequest function but for path-value routes (ie: /tag/{tag_id})
func makePathValueRequest(
	t *testing.T,
	method string,
	target string,
	pathName string,
	pathValue string,
	handler http.HandlerFunc,
	expectedStatusCode int,
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

	if response.StatusCode != expectedStatusCode {
		t.Fatalf(`Should have returned %d.
		Returned %d instead.
		Sent pathname: %s,
		pathvalue: %s,
		Response: %v`,
			expectedStatusCode, response.StatusCode, pathName, pathValue, string(responseBody))
	}

	return response, string(responseBody)
}

func TestChooseDirHandler(t *testing.T) {
	// Setup
	setUpTest(t)

	// Request
	doRequest := func(body string, expectedStatusCode int) (*http.Response, string) {
		return makeRequest(
			t,
			http.MethodPost,
			"/choose-dir",
			body,
			ChooseDirHandler,
			expectedStatusCode,
		)
	}

	// Valid Path
	// First run initializes a new database, second one opens the existing one
	for range 2 {
		doRequest(fmt.Sprintf(`{
			"dirPath": "%s"
		}`, testDirectoryPath), http.StatusNoContent)
	}

	// Invalid Path
	doRequest((`{"dirPath": "123123"}`), http.StatusInternalServerError)

	// Invalid JSON
	doRequest(`ascascascasc`, http.StatusBadRequest)

	// dirPath Missing
	doRequest(`{"message": "Hello"}`, http.StatusBadRequest)
}

func TestListTagsHandler(t *testing.T) {
	// Setup
	setUpTest(t)

	// Set up DB
	setUpAndFillDB(t)
	defer CloseDB()

	// Request
	response, responseBody := makeRequest(t, http.MethodGet, "/tags", "", ListTagsHandler, http.StatusOK)

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
	doRequest := func(fileID string, expectedStatusCode int) (*http.Response, string) {
		return makePathValueRequest(
			t,
			http.MethodGet,
			"/tags/"+fileID,
			"file_id",
			fileID,
			TagsByFileIDHandler,
			expectedStatusCode,
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
	_, responseBody := doRequest("3", http.StatusOK)

	responseTags := marshalResponse(responseBody)
	if len(responseTags) != 2 {
		t.Fatalf("Should have returned 2 elements, instead got: %v", responseTags)
	}

	// Empty string
	doRequest("", http.StatusBadRequest)

	// String
	doRequest("hello", http.StatusBadRequest)

	// Non-existent file ID
	doRequest("123123", http.StatusNotFound)
}

func TestCreateTagHandler(t *testing.T) {
	// Setup
	setUpTest(t)

	// Set up DB
	setUpAndFillDB(t)
	defer CloseDB()

	// Request
	doRequest := func(body string, expectedStatusCode int) (*http.Response, string) {
		return makeRequest(
			t,
			http.MethodPost,
			"/create-tag",
			body,
			CreateTagHandler,
			expectedStatusCode,
		)
	}

	// Valid Tag
	doRequest(`{"name": "folk"}`, http.StatusCreated)

	// Duplicate Tag
	doRequest(`{"name": "folk"}`, http.StatusConflict)

	// Invalid Names
	for _, name := range testInvalidTagNames {
		doRequest(fmt.Sprintf(`{"name": "%s"}`, name), http.StatusBadRequest)
	}

	// Missing name field
	doRequest(`{"hello": "folk"}`, http.StatusBadRequest)
}

func TestRenameTagHandler(t *testing.T) {
	// Setup
	setUpTest(t)

	// Set up DB
	setUpAndFillDB(t)
	defer CloseDB()

	// Request
	doRequest := func(body string, expectedStatusCode int) (*http.Response, string) {
		return makeRequest(
			t,
			http.MethodPost,
			"/rename-tag",
			body,
			RenameTagHandler,
			expectedStatusCode,
		)
	}

	// Valid
	// Twice to test updating the same tag to its own name
	for range 2 {
		doRequest(`{"tagID": 3,"newName": "folk"}`, http.StatusNoContent)
	}

	// Non-existent Tag ID
	doRequest(`{"tagID": 123123,"newName": "trance"}`, http.StatusNotFound)

	// Updating another tag to same name
	doRequest(`{"tagID": 2,"newName": "folk"}`, http.StatusConflict)

	// Checking all the invalid values
	for _, tagName := range testInvalidTagNames {
		doRequest(fmt.Sprintf(`{"tagID": "%d","newName": "%s"}`, 1, tagName), http.StatusBadRequest)
	}

	// Invalid tag id
	doRequest(fmt.Sprintf(`{"tagID": "%s","newName": "%s"}`, "hello", "fresh"), http.StatusBadRequest)

	// Missing tagID
	doRequest(`{"hello": 2, "newName": "folk"}`, http.StatusBadRequest)

	// Missing newName
	doRequest(`{"tagID": 2, "hello": "folk"}`, http.StatusBadRequest)
}

func TestDeleteTagHandler(t *testing.T) {
	// Setup
	setUpTest(t)

	// Set up DB
	setUpAndFillDB(t)
	defer CloseDB()

	// Request
	doRequest := func(body string, expectedStatusCode int) (*http.Response, string) {
		return makeRequest(
			t,
			http.MethodPost,
			"/delete-tag",
			body,
			DeleteTagHandler,
			expectedStatusCode,
		)
	}

	// Valid
	doRequest(`{"tagID": 2}`, http.StatusNoContent)

	// Non-existent Tag ID
	doRequest(`{"tagID": 123123}`, http.StatusNotFound)

	// Invalid Requests
	invalidRequests := []string{
		`{"hello": 2}`,     // Missing tagID
		`{"tagID": "abc"}`, // Invalid tagId
		`{"tagID": ""}`,    // Empty string
	}

	for _, body := range invalidRequests {
		doRequest(body, http.StatusBadRequest)
	}
}

func TestUpdateTagHandler(t *testing.T) {
	// Setup
	setUpTest(t)

	// Set up DB
	setUpAndFillDB(t)
	defer CloseDB()

	// Request
	doRequest := func(body string, expectedStatusCode int) (*http.Response, string) {
		return makeRequest(
			t,
			http.MethodPost,
			"/update-tag",
			body,
			UpdateTagHandler,
			expectedStatusCode,
		)
	}

	// Valid
	for range 2 {
		doRequest(`{"fileID": 2, "tagIDs": [1,2,3]}`, http.StatusNoContent)
	}

	// Valid, empty tagIDs
	doRequest(`{"fileID": 2, "tagIDs": []}`, http.StatusNoContent)

	// fileID missing
	doRequest(`{"hello": 2, "tagIDs": [1,2,3]}`, http.StatusBadRequest)

	// tagIDs missing
	doRequest(`{"fileID": 2, "hello": [1,2,3]}`, http.StatusBadRequest)

	// Invalid fileID
	doRequest(`{"fileID": "2", "tagIDs": [1,2,3]}`, http.StatusBadRequest)

	// Invalid tagIDs
	doRequest(`{"fileID": 2, "tagIDs": 35}`, http.StatusBadRequest)

	// Non-existent fileID
	doRequest(`{"fileID": 123, "tagIDs": [1,2,3]}`, http.StatusNotFound)

	// Non-existent tag ID
	doRequest(`{"fileID": 2, "tagIDs": [1,2,123,456]}`, http.StatusNotFound)
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
	doRequest := func(fileID string, expectedStatusCode int) (*http.Response, string) {
		return makePathValueRequest(
			t,
			http.MethodGet,
			"/media/"+fileID,
			"file_id",
			fileID,
			MediaHandler,
			expectedStatusCode,
		)
	}

	// Valid Request
	_, responseBody := doRequest("1", http.StatusOK)

	if responseBody != testFileContents {
		t.Fatalf("expected %q, got %q", testFileContents, responseBody)
	}

	// Empty string
	doRequest("", http.StatusBadRequest)

	// String
	doRequest("hello", http.StatusBadRequest)

	// File not found
	doRequest("123123123", http.StatusNotFound)
}
