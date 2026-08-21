package main

import "errors"

var (
	// String Errors
	ErrOnlySpaces        = errors.New("consists of only spaces")
	ErrSurroundingSpaces = errors.New("contains trailing or leading spaces")

	ErrTagNameTooLong = errors.New("tag name is too long")

	// Database Errors
	ErrOpeningDatabase    = errors.New("could not create/open the database file")
	ErrDatabaseConnection = errors.New("could not connect to the database")
	ErrPlaylistDirCreate  = errors.New("could not create the playlist directory")
	ErrTableCreation      = errors.New("could not create the database tables")
	ErrUpdatingFiles      = errors.New("could not update the files for the database")

	// Tag Group Errors
	ErrTagGroupEmptyArr    = errors.New("empty tag group array provided")
	ErrTagGroupEmpty       = errors.New("a tag group is empty")
	ErrTagGroupAmount      = errors.New("tag group amount has to be bigger than 0")
	ErrTagGroupDuplicateID = errors.New("duplicate tag id in a tag group")

	// Database ID Errors
	ErrIDInvalid     = errors.New("invalid ID")
	ErrIDNotPositive = errors.New("ID has to be bigger than 0")
)

// Can be used when validating a JSON with many fields
type ValidationError struct {
	Field   string
	Message string
}

// Error implementation
func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

/* Example Usage
return 0, ValidationError{
    Field:   "name",
    Message: "must not be empty",
}
*/
