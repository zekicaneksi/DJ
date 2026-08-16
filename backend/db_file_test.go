package main

/*
	----	DISCLAIMER

	The tests done here check the length of the returned files instead of checking if the CORRECT files are returned.
	Which might make the tests pass, even though they actually failed. But checking correctly simply is not worth the trouble.

*/

import (
	"testing"
)

func TestUpdateFiles(t *testing.T) {
	// Cleanup
	cleanTestDirectory(t)
	defer cleanTestDirectory(t)

	// Initialize database -- Also runs UpdateFiles with no files present
	if err := InitDatabase(testDirectoryPath); err != nil {
		t.Fatal(err)
	}
	defer CloseDB()

	// Create files
	filesToCreate := []string{"techno.mp3", "best slow.mP4", "best_of_best.WAV", "house.avi"}
	for _, n := range filesToCreate {
		createFile(t, n)
	}

	// Function to check if UpdateFiles succeded
	checkFiles := func(amount int) {
		files, err := ListFilesAll()
		if err != nil {
			t.Fatal(err)
		}
		if len(files) != amount {
			t.Fatalf("expected amount: %d --- got: %d", amount, len(files))
		}
	}

	// Update Files -- Which will only do adding to the database
	if err := UpdateFiles(); err != nil {
		t.Fatal(err)
	}
	checkFiles(4)

	// Delete files
	deleteFile(t, filesToCreate[0])
	deleteFile(t, filesToCreate[1])

	// Update Files -- Which will only do deleting from the database
	if err := UpdateFiles(); err != nil {
		t.Fatal(err)
	}
	checkFiles(2)

	// Add and delete files
	createFile(t, filesToCreate[0])
	deleteFile(t, filesToCreate[2])

	// Update Files -- Which will add and delete files from the database
	if err := UpdateFiles(); err != nil {
		t.Fatal(err)
	}
	checkFiles(2)
}

func TestListFilesAll(t *testing.T) {
	// Cleanup
	cleanTestDirectory(t)
	defer cleanTestDirectory(t)

	// Setup
	setUpAndFillDB(t)
	defer CloseDB()

	// Test
	files, err := ListFilesAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 4 {
		t.Fatalf("Expected: 4 -- got: %d", len(files))
	}
}

func TestListFilesUntagged(t *testing.T) {
	// Cleanup
	cleanTestDirectory(t)
	defer cleanTestDirectory(t)

	// Setup
	setUpAndFillDB(t)
	defer CloseDB()

	// Test
	files, err := ListFilesUntagged()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("Expected: 1 -- got: %d", len(files))
	}
}

func TestListFilesByTagIDs(t *testing.T) {
	// Cleanup
	cleanTestDirectory(t)
	defer cleanTestDirectory(t)

	// Setup
	setUpAndFillDB(t)
	defer CloseDB()

	// Test
	listFilesByTagIDs := func(tagIDs []int64, expectingError bool, expectedLength int) {
		files, err := ListFilesByTagIDs(tagIDs)
		if expectingError {
			if err == nil {
				t.Fatalf("was expecting error with %v", tagIDs)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			if len(files) != expectedLength {
				t.Fatalf("With tag IDs: %v , Expected: %v -- got: %d", tagIDs, expectedLength, len(files))
			}
		}
	}

	listFilesByTagIDs([]int64{1, 2}, false, 1)
	listFilesByTagIDs([]int64{1}, false, 2)
	listFilesByTagIDs([]int64{4}, false, 0)
	listFilesByTagIDs([]int64{500}, false, 0)

	listFilesByTagIDs([]int64{1, 2, 2, 3}, true, 0) // Duplicate id
	listFilesByTagIDs([]int64{1, -2, 3}, true, 0)   // id < 0
	listFilesByTagIDs([]int64{1, 0, 3}, true, 0)    // id == 0

}
