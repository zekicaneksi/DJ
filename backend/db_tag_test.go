package main

import "testing"

func TestCreateTag(t *testing.T) {
	// Setup
	setUpTest(t)

	// Initialize database
	if err := InitDatabase(testDirectoryPath); err != nil {
		t.Fatal(err)
	}
	defer CloseDB()

	// Test
	createTag := func(tagName string, expectingError bool) {
		_, err := CreateTag(tagName)
		if expectingError {
			if err == nil {
				t.Fatalf("Was expecting error with: %s", tagName)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	// Valid
	createTag("techno", false)
	createTag("instrumental elitist", false)

	// Tag with an existing name
	createTag("techno", true)

	// Invalid ones
	for _, name := range testInvalidTagNames {
		createTag(name, true)
	}
}

func TestRenameTag(t *testing.T) {
	// Setup
	setUpTest(t)

	// Initialize database
	if err := InitDatabase(testDirectoryPath); err != nil {
		t.Fatal(err)
	}
	defer CloseDB()

	// Create tags to rename
	tagID_1, err := CreateTag("hippie")
	if err != nil {
		t.Fatal(err)
	}

	_, err = CreateTag("melody")
	if err != nil {
		t.Fatal(err)
	}

	// Test
	renameTag := func(newName string, expectingError bool) {
		err := RenameTag(tagID_1, newName)
		if expectingError {
			if err == nil {
				t.Fatalf("Was expecting error with: %s", newName)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	// Valid
	renameTag("techno", false)

	// Same one twice
	for range 2 {
		renameTag("instrumental elitist", false)
	}

	// Tag name that already exists
	renameTag("melody", true)

	// Invalid ones
	for _, name := range testInvalidTagNames {
		renameTag(name, true)
	}
}

func TestDeleteTag(t *testing.T) {
	// Setup
	setUpTest(t)

	// Initialize database
	if err := InitDatabase(testDirectoryPath); err != nil {
		t.Fatal(err)
	}
	defer CloseDB()

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

	// Test
	deleteTag := func(tagID int64, expectingError bool) {
		err := DeleteTag(tagID)
		if expectingError {
			if err == nil {
				t.Fatalf("Was expecting error with tag id: %d", tagID)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	deleteTag(tagIDs[0], false)
	deleteTag(tagIDs[3], false)
	deleteTag(300, false) // Tag ID that does not exist
	deleteTag(0, false)   // Tag ID = 0
	deleteTag(-2, false)  // Tag ID < 0

}

func TestListTags(t *testing.T) {
	// Setup
	setUpTest(t)

	// Initialize database
	if err := InitDatabase(testDirectoryPath); err != nil {
		t.Fatal(err)
	}
	defer CloseDB()

	// Create Tags
	tagsToCreate := []string{"old", "new", "good", "bad"}
	for _, tag := range tagsToCreate {
		if _, err := CreateTag(tag); err != nil {
			t.Fatal(err)
		}
	}

	// Test
	tags, err := ListTagsAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 4 {
		t.Fatalf("Was expecting: 4 --- got: %d", len(tags))
	}
}

func TestListTagsByFileID(t *testing.T) {
	// Setup
	setUpTest(t)

	// Setup
	setUpAndFillDB(t)
	defer CloseDB()

	// Test
	listTagsByFileID := func(fileID int64, expectingError bool, expectedTagIDs []int64) {
		tags, err := ListTagsByFileID(fileID)
		if expectingError {
			if err == nil {
				t.Fatalf("With file ids: %v, was expecting error. Instead got: %v", fileID, tags)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			if len(tags) != len(expectedTagIDs) {
				t.Fatalf("expected ids: %v --- instead got: %v", expectedTagIDs, tags)
			}
			for i := range tags {
				if tags[i].ID != expectedTagIDs[i] {
					t.Fatalf("expected ids: %v --- instead got: %v", expectedTagIDs, tags)
				}
			}
		}
	}

	listTagsByFileID(1, false, []int64{1})
	listTagsByFileID(2, false, []int64{2})
	listTagsByFileID(3, false, []int64{1, 2})
	listTagsByFileID(300, false, []int64{}) // File ID does not exist
}

func TestUpdateTags(t *testing.T) {
	// Setup
	setUpTest(t)

	// Setup
	setUpAndFillDB(t)
	defer CloseDB()

	// Test
	updateTags := func(fileID int64, expectingError bool, newTagIDs []int64) {
		err := UpdateTags(fileID, newTagIDs)
		if expectingError {
			if err == nil {
				t.Fatalf("Was expecting error with file id: %d, and tags: %v", fileID, newTagIDs)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			tags, err := ListTagsByFileID(fileID)
			if err != nil {
				t.Fatalf("error while listing tags: %v", err)
			}
			if len(tags) != len(newTagIDs) {
				t.Fatalf("expected ids: %v --- instead got: %v", newTagIDs, tags)
			}
			for i := range tags {
				if tags[i].ID != newTagIDs[i] {
					t.Fatalf("expected ids: %v --- instead got: %v", newTagIDs, tags)
				}
			}
		}
	}

	updateTags(2, false, []int64{3})
	updateTags(1, false, []int64{1, 2})
	updateTags(1, false, []int64{1, 2}) // Giving it the same tags

	updateTags(1, true, []int64{0, 2})   // Tag ID = 0
	updateTags(1, true, []int64{-4, 2})  // Tag ID < 0
	updateTags(1, true, []int64{1, 200}) // Tag ID does not exist
	updateTags(1, true, []int64{1, 1})   // Duplicate Tag IDs
}
