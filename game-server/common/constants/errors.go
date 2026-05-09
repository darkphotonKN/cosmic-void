package commonconstants

import "errors"

/**
* All custom error types in the application, allowing for consistent
* reference to the same types of errors.
**/
var (
	ErrNotFound             = errors.New("resource not found")
	ErrInvalidInput         = errors.New("invalid input")
	ErrDuplicateResource    = errors.New("resource already exists")
	ErrConstraintViolation  = errors.New("input does not follow column constraints")
	ErrForbidden            = errors.New("you do not have permission to access this resource")
	ErrUnauthorized         = errors.New("incorrect credentials entered during when attempting to authenticate")
	ErrTransient            = errors.New("transient error")
	ErrUUIDCouldNotBeParsed = errors.New("uuid could not be parsed")

	// game
	ErrGameDoesntExist = errors.New("game does not exist")

	// outbox
	ErrOutboxItemNotFound = errors.New("outbox item not found")

	// inbox (consumer-side idempotency)
	ErrAlreadyProcessed = errors.New("event already processed")
)
