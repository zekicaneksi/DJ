package main

import (
	"testing"
)

// This test simulates intended user behavior
func TestWorkflow(t *testing.T) {
	// Cleanup
	cleanTestDirectory(t)
	defer cleanTestDirectory(t)

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
	tagIDs := make([]int64, 0, len(tagsToCreate))

	for _, tag := range tagsToCreate {
		tagID, err := CreateTag(tag)
		if err != nil {
			t.Fatalf("error when creating tag %s : %v", tag, err)
		}
		tagIDs = append(tagIDs, tagID)
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
			TagsIDs: []int64{1},
			Amount:  2,
		},
		{
			TagsIDs: []int64{2, 3},
			Amount:  2,
		},
		{
			TagsIDs: []int64{4},
			Amount:  1,
		},
	})
	if err != nil {
		t.Fatalf("Error when creating the playlist: %v", err)
	}

}
