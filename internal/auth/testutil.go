package auth

import (
	"github.com/google/uuid"
)

// MustParseUUID parses a UUID string or panics
// This is a helper function for tests
func MustParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}
