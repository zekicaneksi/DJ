package main

import (
	"strconv"
	"strings"
)

// Checks if given string is a valid database ID
func ValidateDbId(param_id string) (int64, error) {
	id, err := strconv.ParseInt(param_id, 10, 64)
	if err != nil {
		return 0, ErrIDInvalid
	}
	if id <= 0 {
		return 0, ErrIDNotPositive
	}

	return id, nil
}

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
