package main

// Creates a tag and returns the created tag's id
func CreateTag(tagName string) (int64, error) {
	if err := ValidateTagName(tagName); err != nil {
		return 0, err
	}

	result, err := dbHandle.Exec(
		"INSERT INTO tag (name) VALUES (?)",
		tagName,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Renames a tag
func RenameTag(tagID int64, newName string) error {
	if err := ValidateTagName(newName); err != nil {
		return err
	}

	_, err := dbHandle.Exec(`
		UPDATE tag
		SET name = ?
		WHERE id = ?
	`, newName, tagID)
	if err != nil {
		return err
	}

	return nil
}

// Deletes a tag
func DeleteTag(tagID int64) error {
	_, err := dbHandle.Exec(
		"DELETE FROM tag WHERE id = ?",
		tagID,
	)
	if err != nil {
		return err
	}

	return nil
}

// Returns all of the tags at once
func ListTags() ([]Tag, error) {
	rows, err := dbHandle.Query("SELECT id, name FROM tag ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag

	for rows.Next() {
		var tag Tag

		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// Lists tags of a file by its ID
func ListTagsByFileID(fileID int64) ([]Tag, error) {
	rows, err := dbHandle.Query(`
		SELECT t.id, t.name FROM tag AS t
		JOIN file_tag AS ft ON ft.tag_id = t.id
		WHERE ft.file_id=? ORDER BY t.id`, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag

	for rows.Next() {
		var tag Tag

		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// Clears the tags on a file, then attaches given tags
func UpdateTags(fileID int64, tagIDs []int64) error {

	// Begin transaction
	tx, err := dbHandle.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear tags
	_, err = tx.Exec(`
		DELETE FROM file_tag
		WHERE file_id = ?
	`, fileID)
	if err != nil {
		return err
	}

	// Insert new tags
	stmt, err := tx.Prepare(`
		INSERT INTO file_tag (file_id, tag_id)
		VALUES(?,?);
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, tagID := range tagIDs {
		if _, err := stmt.Exec(fileID, tagID); err != nil {
			return err
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
