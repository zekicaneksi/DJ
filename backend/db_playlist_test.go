package main

import "testing"

/*
	----	DISCLAIMER
	This test only tests if the creation succeeds or not, but doesn't check if the playlist is created by correct files.
	Checking that is a ton of work and is not worth it.
*/
func TestCreatePlaylist(t *testing.T) {
	// Cleanup
	cleanTestDirectory(t)
	defer cleanTestDirectory(t)

	// Setup
	setUpAndFillDB(t)
	defer CloseDB()

	// Test
	tagGroups := []TagGroup{
		{
			TagsIDs: []int64{1, 2},
			Amount:  3,
		},
		{
			TagsIDs: []int64{3},
			Amount:  1,
		},
		{
			TagsIDs: []int64{},
			Amount:  1,
		},
		{
			TagsIDs: []int64{3},
			Amount:  -2,
		},
		{
			TagsIDs: []int64{1, 2, 2},
			Amount:  1,
		},
	}

	if _, err := CreatePlaylist([]TagGroup{
		tagGroups[0],
		tagGroups[1],
	}); err != nil {
		t.Fatal(err)
	}

	// Empty TagIDs
	if _, err := CreatePlaylist([]TagGroup{
		tagGroups[2],
	}); err == nil {
		t.Fatalf("Expected error with: %v", tagGroups[2])
	}

	// Amount < 0
	if _, err := CreatePlaylist([]TagGroup{
		tagGroups[3],
	}); err == nil {
		t.Fatalf("Expected error with: %v", tagGroups[3])
	}

	// Duplicates in TagIDs
	if _, err := CreatePlaylist([]TagGroup{
		tagGroups[4],
	}); err == nil {
		t.Fatalf("Expected error with: %v", tagGroups[4])
	}

	// All together with valid and invalid TagGroups
	if _, err := CreatePlaylist(tagGroups); err == nil {
		t.Fatalf("Expected error with: %v", tagGroups)
	}

}
