package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

var (
	testDirectoryPath = "test"
	testFileContents  = "Hello from DJ!\n"
)

// Delete all the test files from the test directory
func cleanTestDirectory(t *testing.T) {
	files, err := os.ReadDir(testDirectoryPath)
	if err != nil {
		t.Fatalf("error when reading directory: %v", err)
	}

	filesToKeep := []string{"README.md", ".gitignore"}

	for _, file := range files {
		fileName := file.Name()
		if slices.Contains(filesToKeep, fileName) {
			continue
		}
		// File might be busy (especially db file), waiting a little
		for i := range 5 {
			err := os.RemoveAll(filepath.Join(testDirectoryPath, fileName))
			if err != nil {
				if i == 4 {
					t.Fatalf("error when deleting the file: %s --- %v", fileName, err)
				} else {
					time.Sleep(500 * time.Millisecond)
				}
			} else {
				break
			}
		}
	}
}

// Create a file in the test directory
func createFile(t *testing.T, fileName string) {
	file, err := os.Create(filepath.Join(testDirectoryPath, fileName))
	if err != nil {
		t.Fatalf("error when creating a file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(testFileContents)
	if err != nil {
		t.Fatalf("error when creating a file: %v", err)
	}
}

// Delete a file from the test directory
func deleteFile(t *testing.T, fileName string) {
	err := os.Remove(filepath.Join(testDirectoryPath, fileName))
	if err != nil {
		t.Fatalf("error when deleting a file: %v", err)
	}
}

// Sets the database up for testing
func setUpAndFillDB(t *testing.T) {
	// Create files
	filesToCreate := []string{"techno.mp3", "best slow.mP4", "best_of_best.WAV", "house.avi"}
	for _, n := range filesToCreate {
		createFile(t, n)
	}

	// Initialize database
	if err := InitDatabase(testDirectoryPath); err != nil {
		t.Fatal(err)
	}

	// Create Tags
	tagsToCreate := []string{"old", "new", "good", "bad"}
	tagIDs := make([]int64, 0, len(tagsToCreate))

	for _, tag := range tagsToCreate {
		tagID, err := CreateTag(tag)
		if err != nil {
			t.Fatal(err)
		}
		tagIDs = append(tagIDs, tagID)
	}

	// Attach tags to files
	if err := UpdateTags(1, []int64{tagIDs[0]}); err != nil {
		t.Fatal(err)
	}

	if err := UpdateTags(2, []int64{tagIDs[1]}); err != nil {
		t.Fatal(err)
	}

	if err := UpdateTags(3, []int64{tagIDs[0], tagIDs[1]}); err != nil {
		t.Fatal(err)
	}
}
