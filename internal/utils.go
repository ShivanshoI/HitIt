package internal

import (
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MustObjectID panics if the hex string is not a valid ObjectID.
// Use only when you are sure the ID is valid (e.g., from a trusted context).
func MustObjectID(hex string) primitive.ObjectID {
	id, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		panic("invalid object id: " + hex)
	}
	return id
}

// PtrObjectID returns a pointer to an ObjectID from hex, or nil if invalid.
func PtrObjectID(hex string) *primitive.ObjectID {
	id, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return nil
	}
	return &id
}

// LogInfo logs a message with an [INFO] prefix.
func LogInfo(format string, v ...interface{}) {
	log.Printf("[INFO] "+format, v...)
}
