package main

import "errors"

var (
	ErrTagNameTooLong = errors.New("tag name is too long")

	ErrOnlySpaces        = errors.New("consists of only spaces")
	ErrSurroundingSpaces = errors.New("contains trailing or leading spaces")
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
