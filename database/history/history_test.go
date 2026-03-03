package history

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// This is a minimal spec file for testing the history struct logic.
// Database connection tests would require a live or mocked mongo instance.
func TestRequestHistoryStruct(t *testing.T) {
	userId := primitive.NewObjectID()
	reqId := primitive.NewObjectID()
	colId := primitive.NewObjectID()
	
	rh := RequestHistory{
		UserID:            userId,
		RequestID:         reqId,
		CollectionID:      colId,
		Name:              "Test Request",
		Method:            "GET",
		URL:               "http://localhost/test",
		StatusCode:        200,
		ResponseTimeMs:    150,
		ResponseSizeBytes: 1024,
		ExecutedAt:        time.Now(),
	}

	if rh.Method != "GET" {
		t.Errorf("Expected method GET, got %s", rh.Method)
	}

	if rh.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", rh.StatusCode)
	}

	if rh.ResponseTimeMs != 150 {
		t.Errorf("Expected response time 150, got %d", rh.ResponseTimeMs)
	}
}
