package commonhelpers

/*
Commonly shared error helpers utilities.
*/

import (
	"database/sql"
	"errors"
	"strings"

	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
)

/**
* Analyzes which type of custom error an error is and returns the
* appropriate error type. If the error is a new type then return it directly.
**/
func AnalyzeDBErr(err error) error {
	if err == nil {
		return nil
	}
	// match custom error types
	if IsDuplicateError(err) {
		return commonconstants.ErrDuplicateResource
	}
	if IsConstraintViolation(err) {
		return commonconstants.ErrConstraintViolation
	}
	if errors.Is(err, sql.ErrNoRows) {
		return commonconstants.ErrNotFound
	}

	// unexpected errors
	return err
}

/**
* Helper function to determine if an error is a "duplicate item" error.
**/
func IsDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "duplicate key value")
}

/**
* Helper function to determine if an error is from an attempt to insert without
* following column constraints.
**/
func IsConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "violates check constraint")
}
