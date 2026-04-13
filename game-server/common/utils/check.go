package commonhelpers

import "github.com/google/uuid"

func UuidPtrToString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}
