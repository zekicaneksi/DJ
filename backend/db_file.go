package main

import (
	"fmt"
	"strings"
)

// Checks the directory the database is in
// Adds new files into the database and removes the missing files from the database
func UpdateFiles() error {

	fileNames, err := GetDirMediaFiles(dbPath)
	if err != nil {
		return err
	}

	// -- Update the database
	tx, err := dbHandle.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Add new files
	stmt, err := tx.Prepare(`
		INSERT INTO file(name)
		VALUES(?)
		ON CONFLICT(name) DO NOTHING;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, name := range fileNames {
		if _, err := stmt.Exec(name); err != nil {
			return err
		}
	}

	// Remove missing files
	if len(fileNames) == 0 {
		_, err := tx.Exec("DELETE FROM file")
		if err != nil {
			return err
		}
	} else {
		placeholders := strings.TrimRight(strings.Repeat("?,", len(fileNames)), ",")

		query := fmt.Sprintf(
			"DELETE FROM file WHERE name NOT IN (%s)",
			placeholders,
		)

		args := make([]any, len(fileNames))
		for i, n := range fileNames {
			args[i] = n
		}

		if _, err := tx.Exec(query, args...); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil

}

// List files by tag IDs
// If the first tag ID is "0", all untagged files will be returned
// If tagIDs is empty, all files will be returned
func ListFiles(tagIDs []int64) ([]File, error) {
	var (
		files       []File
		queryString string
		queryArgs   []any
	)

	if len(tagIDs) == 0 {
		// Return all files
		queryString = `
			SELECT id, name
			FROM file
			ORDER BY id;
		`
	} else if tagIDs[0] == 0 {
		// Return untagged files
		queryString = `
			SELECT f.id, f.name
			FROM file AS f
			LEFT JOIN file_tag AS ft ON ft.file_id = f.id
			WHERE ft.file_id IS NULL
			ORDER BY f.id;
		`
	} else {
		// Return files by tagIDs
		placeholders := strings.TrimRight(strings.Repeat("?,", len(tagIDs)), ",")

		queryString = fmt.Sprintf(`
			SELECT f.id, f.name
			FROM file AS f
			JOIN file_tag AS ft ON ft.file_id = f.id
			WHERE ft.tag_id IN (%s)
			GROUP BY f.id, f.name
			HAVING COUNT(*) = ?
			ORDER BY f.id;
		`, placeholders)

		queryArgs = make([]any, len(tagIDs))
		for i, tagID := range tagIDs {
			queryArgs[i] = tagID
		}
		queryArgs = append(queryArgs, len(tagIDs))
	}

	// Execute the query
	rows, err := dbHandle.Query(queryString, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var file File

		if err := rows.Scan(&file.ID, &file.Name); err != nil {
			return nil, err
		}

		files = append(files, file)
	}

	return files, rows.Err()
}
