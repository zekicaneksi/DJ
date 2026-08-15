package main

import (
	"fmt"
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
		return fmt.Errorf("empty tag group array provided")
	}

	for _, tagGroup := range tagGroups {
		if len(tagGroup.TagsIDs) == 0 {
			return fmt.Errorf("tag group is empty")
		}

		if tagGroup.Amount <= 0 {
			return fmt.Errorf("tag group amount has to be bigger than 0")
		}

		// Looking for duplicate ids, and returning error if found
		tagSet := make(map[int64]struct{}, len(tagGroup.TagsIDs))

		for _, tagID := range tagGroup.TagsIDs {
			if _, exists := tagSet[tagID]; exists {
				return fmt.Errorf("duplicate tag ID %d in tag group", tagID)
			}

			tagSet[tagID] = struct{}{}
		}
	}

	return nil
}
