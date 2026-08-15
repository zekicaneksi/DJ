package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Creates a shuffled playlist by the provided tag groups and amounts
// returns the created file's absolute path
func CreatePlaylist(tagGroups []TagGroup) (string, error) {

	if err := ValidateTagGroups(tagGroups); err != nil {
		return "", err
	}

	tx, err := dbHandle.Begin()
	if err != nil {
		return "", fmt.Errorf("error when beginning playlist transaction: %w", err)
	}
	defer tx.Rollback()

	var files []File

	// IDs selected by earlier groups to prevent multiple selection of same files
	selectedIDs := make(map[int64]struct{})

	for _, tagGroup := range tagGroups {

		// For the (?, ?, ?, ...)
		tagPlaceholders := make([]string, len(tagGroup.TagsIDs))
		// Arguments to provide to the query in the end
		args := make([]any, 0, len(tagGroup.TagsIDs)+len(selectedIDs)+1)

		for i, tagID := range tagGroup.TagsIDs {
			tagPlaceholders[i] = "?"
			args = append(args, tagID)
		}

		query := `
			SELECT f.id, f.name
			FROM file AS f
			JOIN file_tag AS ft ON ft.file_id = f.id
			WHERE ft.tag_id IN (` + strings.Join(tagPlaceholders, ",") + `)
		`

		// Exclude files already selected by previous groups.
		if len(selectedIDs) > 0 {
			excludePlaceholders := make([]string, 0, len(selectedIDs))

			for id := range selectedIDs {
				excludePlaceholders = append(excludePlaceholders, "?")
				args = append(args, id)
			}

			query += `
				AND f.id NOT IN (` + strings.Join(excludePlaceholders, ",") + `)
			`
		}

		query += `
			GROUP BY f.id, f.name
			HAVING COUNT(DISTINCT ft.tag_id) = ?
			ORDER BY RANDOM()
			LIMIT ?
		`

		args = append(args, len(tagGroup.TagsIDs), tagGroup.Amount)

		rows, err := tx.Query(query, args...)
		if err != nil {
			return "", fmt.Errorf("error when querying files for playlist: %w", err)
		}

		for rows.Next() {
			var f File

			if err := rows.Scan(&f.ID, &f.Name); err != nil {
				rows.Close()
				return "", fmt.Errorf("error when scanning file: %w", err)
			}

			files = append(files, f)
			selectedIDs[f.ID] = struct{}{}
		}

		if err := rows.Err(); err != nil {
			rows.Close()
			return "", fmt.Errorf("error when iterating files: %w", err)
		}

		rows.Close()

	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	// Shuffling the files
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	fileNames := make([]string, len(files))

	for i, file := range files {
		fileNames[i] = file.Name
	}

	r.Shuffle(len(fileNames), func(i, j int) {
		fileNames[i], fileNames[j] = fileNames[j], fileNames[i]
	})

	playlistPath, err := CreateM3U8(fileNames)
	if err != nil {
		return "", fmt.Errorf("error when committing playlist transaction: %w", err)
	}

	return playlistPath, nil
}
