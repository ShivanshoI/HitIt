package execution

import (
	"context"
	"io"
	"log"
	"net/http"
	"pog/database/history"
	"pog/database/requests"
	"pog/internal"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExecutionService struct {
	requestRepo *requests.RequestRepository
	historyRepo *history.HistoryRepository
}

func NewExecutionService(requestRepo *requests.RequestRepository, historyRepo *history.HistoryRepository) *ExecutionService {
	return &ExecutionService{
		requestRepo: requestRepo,
		historyRepo: historyRepo,
	}
}

func (s *ExecutionService) ExecuteRequest(ctx context.Context, requestID string, userID string) (*ExecutionResult, error) {
	userIdObj, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, internal.NewBadRequest("invalid user id")
	}

	reqModel, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return nil, internal.NewNotFound("Request not found")
	}

	// Basic variable replacement could happen here in the future
	finalURL := reqModel.URL

	// 1. Prepare the proxy HTTP request
	var bodyReader io.Reader
	if reqModel.Body != "" {
		bodyReader = strings.NewReader(reqModel.Body)
	}

	proxyReq, err := http.NewRequestWithContext(ctx, reqModel.Method, finalURL, bodyReader)
	if err != nil {
		return nil, internal.NewInternalError("Failed to construct proxy request: " + err.Error())
	}

	// 2. Inject headers
	hasUserAgent := false
	for _, h := range reqModel.Headers {
		if strings.ToLower(h.Key) == "host" {
			proxyReq.Host = h.Value
		} else {
			proxyReq.Header.Set(h.Key, h.Value)
		}
		if strings.ToLower(h.Key) == "user-agent" {
			hasUserAgent = true
		}
	}

	// 2.1 Handle Auth field if present and Authorization header is not already set
	if reqModel.Auth != "" && proxyReq.Header.Get("Authorization") == "" {
		// Just a simple assumption that Auth might be the token or the full "Bearer <token>"
		// You might need to adjust this depending on how the frontend saves the Auth type
		proxyReq.Header.Set("Authorization", reqModel.Auth)
	}

	// 2.2 Default User-Agent to prevent WAFs from blocking Go-http-client
	if !hasUserAgent {
		proxyReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	}

	// Default Content-Type to application/json if not set, and body is provided
	if proxyReq.Header.Get("Content-Type") == "" && reqModel.Body != "" {
		proxyReq.Header.Set("Content-Type", "application/json")
	}

	// 3. Add Query Params (This will append to existing URL query string if any)
	q := proxyReq.URL.Query()
	for _, p := range reqModel.Params {
		q.Add(p.Key, p.Value)
	}
	proxyReq.URL.RawQuery = q.Encode()

	// 4. Execute the Hit
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	start := time.Now()
	resp, err := client.Do(proxyReq)
	duration := time.Since(start)

	if err != nil {
		// If the HTTP call fails completely (e.g., DNS error, timeout)
		return nil, internal.NewInternalError("Request execution failed: " + err.Error())
	}
	defer resp.Body.Close()

	// 5. Read response
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil && err != io.EOF {
		return nil, internal.NewInternalError("Failed to read response body")
	}
	respSizeBytes := len(respBodyBytes)

	// Build the response headers payload
	respHeaders := []KeyValuePair{}
	for k, v := range resp.Header {
		val := ""
		if len(v) > 0 {
			val = strings.Join(v, ", ") // join multiple values e.g. multiple Set-Cookie
		}
		respHeaders = append(respHeaders, KeyValuePair{Key: k, Value: val})
	}

	execResult := &ExecutionResult{
		StatusCode:        resp.StatusCode,
		StatusText:        http.StatusText(resp.StatusCode),
		ResponseTimeMs:    duration.Milliseconds(),
		ResponseSizeBytes: respSizeBytes,
		Headers:           respHeaders,
		Body:              string(respBodyBytes),
	}

	// 6. Asynchronously save history logging
	go func() {
		bgCtx := context.Background()
		historyEntry := &history.RequestHistory{
			UserID:            userIdObj,
			RequestID:         reqModel.ID,
			CollectionID:      reqModel.CollectionID,
			Name:              reqModel.Name,
			Method:            reqModel.Method,
			URL:               finalURL,
			StatusCode:        execResult.StatusCode,
			ResponseTimeMs:    execResult.ResponseTimeMs,
			ResponseSizeBytes: execResult.ResponseSizeBytes,
			// ExecutedAt is set by repo
		}

		// Also extract from the original request context because we're using a background context
		if teamID, ok := ctx.Value(internal.TeamIDKey).(string); ok && teamID != "" {
			if objTeamID, err := primitive.ObjectIDFromHex(teamID); err == nil {
				historyEntry.TeamID = &objTeamID
			}
		}

		_, hErr := s.historyRepo.Create(bgCtx, historyEntry)
		if hErr != nil {
			log.Printf("[SERVICE] Failed to save execution history: %v", hErr)
		}
	}()

	return execResult, nil
}

// GetHistory fetches the recent execution logs for a user with pagination
func (s *ExecutionService) GetHistory(ctx context.Context, userID string, page int, limit int) ([]history.RequestHistory, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	historyLogs, total, err := s.historyRepo.ListByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, internal.NewInternalError("Failed to fetch history")
	}
	return historyLogs, total, nil
}

// ClearHistory deletes all history for a user
func (s *ExecutionService) ClearHistory(ctx context.Context, userID string) error {
	err := s.historyRepo.DeleteAllByUserID(ctx, userID)
	if err != nil {
		return internal.NewInternalError("Failed to clear history")
	}
	return nil
}
