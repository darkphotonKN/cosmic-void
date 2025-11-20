package commonconstants

import "errors"

/**
* All custom error types in the application, allowing for consistent
* reference to the same types of errors.
**/
var (
	ErrNotFound            = errors.New("Resource not found.")
	ErrInvalidInput        = errors.New("Invalid input.")
	ErrDuplicateResource   = errors.New("Resource already exists.")
	ErrConstraintViolation = errors.New("Input does not follow column constraints.")
	ErrForbidden           = errors.New("You do not have permission to access this resource.")
	ErrUnauthorized        = errors.New("Incorrect credentials entered during when attempting to authenticate.")
)
