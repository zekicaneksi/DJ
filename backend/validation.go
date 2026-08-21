package main

import (
	"strings"
)

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
