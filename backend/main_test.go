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
	fullDBPath := testDirectoryPath + "/dj.sqlite"

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
	// Initialize with no media files
	if err := InitDatabase(fullDBPath); err != nil {
		t.Error(err)
	}

	// Initialize again with new media files
	createFile := func(fileName string) {
		file, err := os.Create(testDirectoryPath + "/" + fileName)
		if err != nil {
			t.Error(err)
		}
		defer file.Close()

		_, err = file.WriteString("Hello from DJ!\n")
		if err != nil {
			t.Error(err)
		}
	}

	filesToCreate := []string{"best_techno.mp3", "best_slow.mP4", "best_of_best.WAW"}
	for _, n := range filesToCreate {
		createFile(n)
	}

	if err := InitDatabase(fullDBPath); err != nil {
		t.Error(err)
	}
	// Initialize again with a missing media file
	removeMediaFile := func(index int) {
		err = os.Remove(testDirectoryPath + "/" + filesToCreate[index])
		if err != nil {
			t.Error(err)
		}
	}

	removeMediaFile(1)
	if err := InitDatabase(fullDBPath); err != nil {
		t.Error(err)
	}

	// Initialize again with all media files missing
	removeMediaFile(0)
	removeMediaFile(2)
	if err := InitDatabase(fullDBPath); err != nil {
		t.Error(err)
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
		t.Error("Error deleting tag", err)
	}

	// --- Rename a tag
	if err := RenameTag(2, "disco"); err != nil {
		t.Error(err)
	}

	// Check if it worked
	tags, err = ListTags()
	if err != nil || tags[1].Name != "disco" {
		t.Error("Error renaming tag", err)
	}
}
