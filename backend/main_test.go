package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
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

func TestEverything(t *testing.T) {

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
		err := os.RemoveAll(filepath.Join(testDirectoryPath, fileName))
		if err != nil {
			t.Error(err)
		}
	}

	// Start Testing

	// --- Initialize Database ---
	// Initialize with no media files
	if err := InitDatabase(testDirectoryPath); err != nil {
		t.Error(err)
	}

	// Initialize with new media files
	fileContents := "Hello from DJ!\n"
	createFile := func(fileName string) {
		file, err := os.Create(filepath.Join(testDirectoryPath, fileName))
		if err != nil {
			t.Error(err)
		}
		defer file.Close()

		_, err = file.WriteString(fileContents)
		if err != nil {
			t.Error(err)
		}
	}

	filesToCreate := []string{"techno.mp3", "best slow.mP4", "best_of_best.WAV", "house.avi"}
	for _, n := range filesToCreate {
		createFile(n)
	}

	if err := InitDatabase(testDirectoryPath); err != nil {
		t.Error(err)
	}

	// --- Create Tags
	tagNames := []string{"techno", "rock", "instrumental"}

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

	// --- Attach tags
	if err := AttachTag(1, []int{1, 2}); err != nil {
		t.Error(err)
	}

	if err := AttachTag(2, []int{2, 3}); err != nil {
		t.Error(err)
	}

	if err := AttachTag(3, []int{1, 3}); err != nil {
		t.Error(err)
	}

	if err := AttachTag(1, []int{2, 3}); err == nil {
		t.Error("The function should have given an error for duplicate rows")
	}

	if err := AttachTag(1, []int{3}); err != nil {
		t.Error(err)
	}

	// --- Rename tag
	if err := RenameTag(2, "disco"); err != nil {
		t.Error(err)
	}

	// --- List Files

	// List all files
	listFiles, err := ListFiles(nil)
	if err != nil || len(listFiles) != 4 {
		t.Error("Error when listing all files", err)
	}

	// List untagged files
	listFiles, err = ListFiles([]int64{0})
	if err != nil || len(listFiles) != 1 {
		t.Error("Error when listing untagged files", err)
	}

	// List by tags
	listFiles, err = ListFiles([]int64{2, 3})
	if err != nil || len(listFiles) != 2 {
		t.Error("Error when listing by tag", err)
	}

	// Create a playlist
	if _, err := CreatePlaylist([]struct {
		TagsIDs []int64
		Amount  int
	}{
		{TagsIDs: []int64{1, 2, 3}, Amount: 1},
		{TagsIDs: []int64{2, 3}, Amount: 3},
	}); err != nil {
		t.Error(err)
	}

	// --- Detach tag
	if err := DetachTag(3, []int{1, 3}); err != nil {
		t.Error(err)
	}

	// --- Delete tag
	if err := DeleteTag(2); err != nil {
		t.Error(err)
	}

	// --- Testing Media Handler (streaming media file over http)
	testMediaHandler := func() {
		req := httptest.NewRequest(
			http.MethodGet,
			"/api/media/2",
			nil,
		)

		// PathValue() only works when the request has been routed.
		// Set the path value manually for the test.
		req.SetPathValue("file_id", "2")

		rec := httptest.NewRecorder()

		MediaHandler(rec, req)

		res := rec.Result()
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Error(err)
		}

		if string(body) != fileContents {
			t.Errorf("expected %q, got %q", fileContents, string(body))
		}
	}

	testMediaHandler()

	// Initialize with a missing media file
	removeMediaFile := func(index int) {
		err = os.Remove(filepath.Join(testDirectoryPath, filesToCreate[index]))
		if err != nil {
			t.Error(err)
		}
	}

	removeMediaFile(1)
	if err := InitDatabase(testDirectoryPath); err != nil {
		t.Error(err)
	}

	// Initialize with all media files missing
	removeMediaFile(0)
	removeMediaFile(2)
	removeMediaFile(3)
	if err := InitDatabase(testDirectoryPath); err != nil {
		t.Error(err)
	}
}
