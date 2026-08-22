package main

import (
	"fmt"
	"log"
	"slices"
	"strings"
)

// Checks if given string is a valid Tag name
func ValidateTagName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrOnlySpaces
	}

	if name != strings.TrimSpace(name) {
		return ErrSurroundingSpaces
	}

	if len(name) > 50 {
		return ErrTagNameTooLong
	}

	return nil
}

// Checks if given []TagGroup is valid to create a playlist with
func ValidateTagGroups(tagGroups []TagGroup) error {
	if len(tagGroups) == 0 {
		return ErrTagGroupEmptyArr
	}

	for _, tagGroup := range tagGroups {
		if len(tagGroup.TagsIDs) == 0 {
			return ErrTagGroupEmpty
		}

		if tagGroup.Amount <= 0 {
			return ErrTagGroupAmount
		}

		// Looking for duplicate ids, and returning error if found
		tagSet := make(map[int64]struct{}, len(tagGroup.TagsIDs))

		for _, tagID := range tagGroup.TagsIDs {
			if _, exists := tagSet[tagID]; exists {
				return ErrTagGroupDuplicateID
			}

			tagSet[tagID] = struct{}{}
		}
	}

	return nil
}

// Checks if given ID's are present in given table
// Returns missing IDs
func CheckIDsInDB(tableName string, IDs []int64) ([]int64, error) {
	// Validating table name
	validTableNames := []string{
		"file",
		"tag",
	}
	if !slices.Contains(validTableNames, tableName) {
		log.Fatalf("unknown table %s", tableName)
	}

	// Check
	if len(IDs) == 0 {
		return nil, nil
	}

	// Preparing Query
	placeholders := make([]string, len(IDs))
	args := make([]any, len(IDs))

	for i, id := range IDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		"SELECT id FROM %s WHERE id IN (%s)",
		tableName,
		strings.Join(placeholders, ", "),
	)

	// Executing Query
	rows, err := dbHandle.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Making a map from returned IDs to check against
	found := make(map[int64]struct{}, len(IDs))

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		found[id] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Matching given IDs with returned IDs
	var missing []int64

	for _, id := range IDs {
		if _, ok := found[id]; !ok {
			missing = append(missing, id)
		}
	}

	return missing, nil
}
