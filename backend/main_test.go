package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
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

func TestDatabase(t *testing.T) {
	testDirectoryPath := "test"
	// Clean the test directory
	files, err := os.ReadDir(testDirectoryPath)
	if err != nil {
		t.Error(err)
	}

	filesToKeep := []string{"README.md", ".gitignore"}

	for _, file := range files {
		fileName := file.Name()
		if slices.Contains(filesToKeep, fileName) {
			continue
		}
		err := os.Remove(testDirectoryPath + "/" + fileName)
		if err != nil {
			t.Error(err)
		}
	}

	// Start Testing

	// --- Initialize Database ---
	for range 2 { // Twice to test initializing already existing DB
		if err := InitDatabase(testDirectoryPath + "/dj.sqlite"); err != nil {
			t.Error(err)
		}
	}

	// --- Tag Operations ---
	tagNames := []string{"techno", "rock", "instrumental"}

	// --- Create Tag
	for _, tagName := range tagNames {
		if _, err := CreateTag(tagName); err != nil {
			t.Error("Error creating tag", err)
		}
	}
	// Try to create the same tags once again, it is supposed to return an error
	for _, tagName := range tagNames {
		if _, err := CreateTag(tagName); err == nil {
			t.Error("This was supposed to return an error")
		}
	}

	// --- List Tags
	tags, err := ListTags()
	if err != nil {
		t.Error(err)
	}
	// Are all tags listed and matching?
	for i, tag := range tags {
		if tag.Name != tagNames[i] {
			t.Error("Tag mismatch")
		}
	}

	// --- Delete a tag
	if err := DeleteTag(3); err != nil {
		t.Error(err)
	}

	// Checking if the tag is really deleted
	tags, err = ListTags()
	if err != nil || len(tags) != 2 {
		t.Error(err)
	}
}
