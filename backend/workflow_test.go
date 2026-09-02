package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
)

// This test simulates intended user behavior
func TestWorkflow(t *testing.T) {
	// Setup
	setUpTest(t)

	// Create files for the directory
	filesToCreate := []string{"techno.mp3", "best slow.mP4", "best_of_best.WAV", "house.avi"}
	for _, n := range filesToCreate {
		createFile(t, n)
	}

	// Initialize database
	if err := InitDatabase(testDirectoryPath); err != nil {
		t.Fatal(err)
	}
	defer CloseDB()

	// List all files
	files, err := ListFilesAll()
	if err != nil {
		t.Fatalf("error when listing all files: %v", err)
	}
	if len(files) != 4 {
		t.Fatalf("Expected 4 items in files. Instead got: %v", files)
	}

	// List untagged files
	files, err = ListFilesUntagged()
	if err != nil {
		t.Fatalf("error when listing untagged files: %v", err)
	}
	if len(files) != 4 {
		t.Fatalf("Expected 4 items in untagged files. Instead got: %v", files)
	}

	// List all tags
	tags, err := ListTagsAll()
	if err != nil {
		t.Fatalf("error when listing all tags: %v", err)
	}
	if len(tags) != 0 {
		t.Fatalf("Expected 0 tags. Instead got: %v", tags)
	}

	// Create tags
	tagsToCreate := []string{"old", "new", "good", "bad", "so loud"}

	for _, tag := range tagsToCreate {
		_, err := CreateTag(tag)
		if err != nil {
			t.Fatalf("error when creating tag %s : %v", tag, err)
		}
	}

	// Rename a tag
	if err := RenameTag(1, "very old"); err != nil {
		t.Fatalf("Error when renaming tag: %v", err)
	}

	// Update tags on files
	updateTag := func(fileID int64, tagIDs []int64) {
		if err := UpdateTags(fileID, tagIDs); err != nil {
			t.Fatalf("Error when updating tags on file: %d -- %v", fileID, tagIDs)
		}
	}

	updateTag(1, []int64{1, 2})
	updateTag(2, []int64{2, 3})
	updateTag(1, []int64{1, 3})
	updateTag(3, []int64{1, 4})
	updateTag(4, []int64{1, 2, 3, 5})

	// Rename a tag
	if err := RenameTag(2, "very new"); err != nil {
		t.Fatalf("Error when renaming tag: %v", err)
	}

	// Delete a tag
	if err := DeleteTag(5); err != nil {
		t.Fatalf("Error when deleting tag: %v", err)
	}

	// Create some more files and update the database with new files
	filesToCreate = []string{"bravery.mp3", "slow song.mp4"}
	for _, n := range filesToCreate {
		createFile(t, n)
	}

	if err := UpdateFiles(); err != nil {
		t.Fatalf("Error when updating files: %v", err)
	}

	// List all files
	files, err = ListFilesAll()
	if err != nil {
		t.Fatalf("error when listing all files: %v", err)
	}
	if len(files) != 6 {
		t.Fatalf("Expected 6 items in files. Instead got: %v", files)
	}

	// List untagged files
	files, err = ListFilesUntagged()
	if err != nil {
		t.Fatalf("error when listing untagged files: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("Expected 2 items in untagged files. Instead got: %v", files)
	}

	// List tags by file ID
	tags, err = ListTagsByFileID(4)
	if err != nil {
		t.Fatalf("error when listing tags by file id 4: %v", err)
	}
	if len(tags) != 3 {
		t.Fatalf("Expected 3 tags. Instead got: %v", tags)
	}

	// List files by tag
	files, err = ListFilesByTagIDs([]int64{2, 3})
	if err != nil {
		t.Fatalf("Error when listing files by tag ids 2,3: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("Expected 2 files, instead got: %v", files)
	}

	// Remove a file and update database
	deleteFile(t, "best slow.mP4")

	if err := UpdateFiles(); err != nil {
		t.Fatalf("Error when updating files after deletion: %v", err)
	}

	// List all tags
	tags, err = ListTagsAll()
	if err != nil {
		t.Fatalf("error when listing all tags: %v", err)
	}
	if len(tags) != 4 {
		t.Fatalf("Expected 4 tags. Instead got: %v", tags)
	}

	// Create a playlist
	_, err = CreatePlaylist([]TagGroup{
		{
			TagIDs: []int64{1},
			Amount: 2,
		},
		{
			TagIDs: []int64{2, 3},
			Amount: 2,
		},
		{
			TagIDs: []int64{4},
			Amount: 1,
		},
	})
	if err != nil {
		t.Fatalf("Error when creating the playlist: %v", err)
	}
}

// This test simulates intended user behavior using the API routes
func TestWorkflowHandlers(t *testing.T) {
	// Setup
	setUpTest(t)

	// Create files for the directory
	filesToCreate := []string{"techno.mp3", "best slow.mP4", "best_of_best.WAV", "house.avi"}
	for _, n := range filesToCreate {
		createFile(t, n)
	}

	// Helper function for testing FilesByTagHandler
	// Makes the request and checks the amount of files in the response
	listFilesByTagIDAndCheck := func(tagIDs []int64, expectedAmount int) {
		requestBody, err := json.Marshal(map[string]any{
			"tagIDs": tagIDs,
		})

		if err != nil {
			t.Fatalf("Error when marshalling %v: %v", tagIDs, err)
		}

		_, responseBody := makeRequest(t, "POST", "/search-files-by-tag", string(requestBody), FilesByTagHandler, http.StatusOK)
		var responseVals struct {
			Files []File `json:"files"`
		}

		err = json.Unmarshal([]byte(responseBody), &responseVals)
		if err != nil {
			t.Fatalf("cannot unmarshal %s: %v", string(responseBody), err)
		}

		if len(responseVals.Files) != expectedAmount {
			t.Fatalf("Should have returned %d elements, instead got: %v", expectedAmount, responseVals.Files)
		}

	}

	// Helper function for testing ListTagsHandler
	// Makes the request and checks the amount of files in the response
	listAllTagsAndCheck := func(expectedAmount int) {
		_, responseBody := makeRequest(t, "GET", "/tags", "", ListTagsHandler, http.StatusOK)
		var responseVals struct {
			Tags []Tag `json:"tags"`
		}

		err := json.Unmarshal([]byte(responseBody), &responseVals)
		if err != nil {
			t.Fatalf("cannot unmarshal %s: %v", string(responseBody), err)
		}

		if len(responseVals.Tags) != expectedAmount {
			t.Fatalf("Should have returned %d elements, instead got: %v", expectedAmount, responseVals.Tags)
		}
	}

	// Helper function for testing TagsByFileIDHandler
	// Makes the request and checks the amount of files in the response
	listTagsByFileIDAndCheck := func(fileID int64, expectedAmount int) {
		_, responseBody := makePathValueRequest(t, "GET", "/tags/{file_id}", "file_id", strconv.FormatInt(fileID, 10), TagsByFileIDHandler, http.StatusOK)
		var responseVals struct {
			Tags []Tag `json:"tags"`
		}

		err := json.Unmarshal([]byte(responseBody), &responseVals)
		if err != nil {
			t.Fatalf("cannot unmarshal %s: %v", string(responseBody), err)
		}

		if len(responseVals.Tags) != expectedAmount {
			t.Fatalf("Should have returned %d elements, instead got: %v", expectedAmount, responseVals.Tags)
		}
	}

	// Helper function for UpdateTagHandler
	updateTag := func(fileID int64, tagIDs []int64) {
		requestBody, err := json.Marshal(map[string]any{
			"fileID": fileID,
			"tagIDs": tagIDs,
		})

		if err != nil {
			t.Fatalf("Error when marshalling %v: %v", tagIDs, err)
		}

		makeRequest(t, "POST", "/update-tag", string(requestBody), UpdateTagHandler, http.StatusNoContent)
	}

	// Initialize database
	makeRequest(t, "POST", "/choose-dir", fmt.Sprintf(`{"dirPath": "%s"}`, testDirectoryPath), ChooseDirHandler, http.StatusNoContent)
	defer CloseDB()

	/*
		// List all files
		files, err = ListFilesAll()
		if err != nil {
			t.Fatalf("error when listing all files: %v", err)
		}
		if len(files) != 4 {
			t.Fatalf("Expected 4 items in files. Instead got: %v", files)
		}
	*/

	/*
		// List untagged files
		files, err = ListFilesUntagged()
		if err != nil {
			t.Fatalf("error when listing untagged files: %v", err)
		}
		if len(files) != 4 {
			t.Fatalf("Expected 4 items in untagged files. Instead got: %v", files)
		}
	*/

	// List all tags
	listAllTagsAndCheck(0)

	// Create tags
	tagsToCreate := []string{"old", "new", "good", "bad", "so loud"}
	for _, tag := range tagsToCreate {
		makeRequest(t, "POST", "/create-tag", fmt.Sprintf(`{"name": "%s"}`, tag), CreateTagHandler, http.StatusCreated)
	}

	// Rename a tag
	makeRequest(t, "POST", "/rename-tag", `{"tagID": 1, "newName": "very old"}`, RenameTagHandler, http.StatusNoContent)

	// Update tags on files
	updateTag(1, []int64{1, 2})
	updateTag(2, []int64{2, 3})
	updateTag(1, []int64{1, 3})
	updateTag(3, []int64{1, 4})
	updateTag(4, []int64{1, 2, 3, 5})

	// Rename a tag
	makeRequest(t, "POST", "/rename-tag", `{"tagID": 2, "newName": "very new"}`, RenameTagHandler, http.StatusNoContent)

	// Delete a tag
	makeRequest(t, "POST", "/delete-tag", `{"tagID": 5}`, DeleteTagHandler, http.StatusNoContent)

	// Create some more files and update the database with new files
	filesToCreate = []string{"bravery.mp3", "slow song.mp4"}
	for _, n := range filesToCreate {
		createFile(t, n)
	}

	makeRequest(t, "POST", "/choose-dir", fmt.Sprintf(`{"dirPath": "%s"}`, testDirectoryPath), ChooseDirHandler, http.StatusNoContent)
	defer CloseDB()

	/*
		// List all files
		files, err = ListFilesAll()
		if err != nil {
			t.Fatalf("error when listing all files: %v", err)
		}
		if len(files) != 6 {
			t.Fatalf("Expected 6 items in files. Instead got: %v", files)
		}
	*/

	/*
		// List untagged files
		files, err = ListFilesUntagged()
		if err != nil {
			t.Fatalf("error when listing untagged files: %v", err)
		}
		if len(files) != 2 {
			t.Fatalf("Expected 2 items in untagged files. Instead got: %v", files)
		}
	*/

	// List tags by file ID
	listTagsByFileIDAndCheck(4, 3)

	// List files by tag
	listFilesByTagIDAndCheck([]int64{2, 3}, 2)

	// Remove a file and update database
	deleteFile(t, "best slow.mP4")

	makeRequest(t, "POST", "/choose-dir", fmt.Sprintf(`{"dirPath": "%s"}`, testDirectoryPath), ChooseDirHandler, http.StatusNoContent)
	defer CloseDB()

	// List all tags
	listAllTagsAndCheck(4)

	// Stream a media file
	{
		_, responseBody := makePathValueRequest(t, http.MethodGet, "/media/3", "file_id", "3", MediaHandler, http.StatusOK)
		if responseBody != testFileContents {
			t.Fatalf("expected %q, got %q", testFileContents, responseBody)
		}
	}

	// Create a playlist
	makeRequest(t, "POST", "/create-playlist", `{
		"tagGroups": [
			{
				"TagIDs": [1],
				"Amount": 2
			},
			{
				"TagIDs": [2,3],
				"Amount": 2
			},
			{
				"TagIDs": [4],
				"Amount": 1
			}
		]
	}`, CreatePlaylistHandler, http.StatusCreated)
}
