package execution

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"pog/database/history"
	"pog/database/requests"
	"pog/internal"
	"pog/internal/localbridge"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExecutionService struct {
	requestRepo  *requests.RequestRepository
	historyRepo  *history.HistoryRepository
	bridge       *localbridge.Client // nil = bridge not configured
}

func NewExecutionService(
	requestRepo *requests.RequestRepository,
	historyRepo *history.HistoryRepository,
	bridge *localbridge.Client, // pass nil if BRIDGE_URL is not set
) *ExecutionService {
	return &ExecutionService{
		requestRepo: requestRepo,
		historyRepo: historyRepo,
		bridge:      bridge,
	}
}

// isLocalhost returns true if the URL targets localhost / 127.0.0.1.
// These cannot be reached by the deployed backend directly — must go via bridge.
func isLocalhost(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := u.Hostname() // strips port
	return host == "localhost" || host == "127.0.0.1"
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

	finalURL := reqModel.URL

	// ── Build shared header map (used by both paths) ──────────────────────────
	headers := map[string]string{}
	hasUserAgent := false

	for _, h := range reqModel.Headers {
		if strings.ToLower(h.Key) == "host" {
			continue // handled separately in direct path
		}
		headers[h.Key] = h.Value
		if strings.ToLower(h.Key) == "user-agent" {
			hasUserAgent = true
		}
	}

	if reqModel.Auth != "" && headers["Authorization"] == "" {
		headers["Authorization"] = reqModel.Auth
	}

	if !hasUserAgent {
		headers["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36"
	}

	if headers["Content-Type"] == "" && reqModel.Body != "" {
		headers["Content-Type"] = "application/json"
	}

	// ── Route decision ────────────────────────────────────────────────────────
	//   localhost URL + bridge configured  →  bridge path
	//   everything else                    →  direct http.Client path (unchanged)

	var execResult *ExecutionResult
	var bridgeRequestID string // stored in history alongside Mongo _id

	if isLocalhost(finalURL) && s.bridge != nil {
		execResult, bridgeRequestID, err = s.executeThroughBridge(ctx, reqModel, finalURL, headers)
	} else {
		execResult, err = s.executeDirect(ctx, reqModel, finalURL, headers)
	}

	if err != nil {
		return nil, err
	}

	// ── Async history save (unchanged logic, bridge_request_id added) ─────────
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
			// BridgeRequestID stored if routed via extension (see history schema note below)
			BridgeRequestID: bridgeRequestID,
		}

		scope := internal.GetScope(ctx)
		if scope.TeamID != "" {
			if objTeamID, err := primitive.ObjectIDFromHex(scope.TeamID); err == nil {
				historyEntry.TeamID = &objTeamID
			}
		}

		if scope.OrgID != "" {
			if objOrgID, err := primitive.ObjectIDFromHex(scope.OrgID); err == nil {
				historyEntry.OrgID = &objOrgID
			}
		}

		_, hErr := s.historyRepo.Create(bgCtx, historyEntry)
		if hErr != nil {
			log.Printf("[SERVICE] Failed to save execution history: %v", hErr)
		}
	}()

	return execResult, nil
}

// executeThroughBridge routes the request via LocalBridge → extension → localhost.
// Returns the execution result and the bridge's requestId (a UUID string).
func (s *ExecutionService) executeThroughBridge(
	ctx context.Context,
	reqModel *requests.APIRequest, // matches database/requests/requests.schema.go
	finalURL string,
	headers map[string]string,
) (*ExecutionResult, string, error) {

	// Append query params to URL (same as direct path)
	if len(reqModel.Params) > 0 {
		u, err := url.Parse(finalURL)
		if err != nil {
			return nil, "", internal.NewInternalError("invalid URL: " + err.Error())
		}
		q := u.Query()
		for _, p := range reqModel.Params {
			q.Add(p.Key, p.Value)
		}
		u.RawQuery = q.Encode()
		finalURL = u.String()
	}

	bridgeResp, err := s.bridge.Do(ctx, reqModel.Method, finalURL, headers, reqModel.Body)
	if err != nil {
		// Distinguish "no extension" from other errors so frontend can show a helpful message
		if strings.Contains(err.Error(), "no extension connected") {
			return nil, "", internal.NewBadRequest("LocalBridge extension is not connected. Open Chrome with the LocalBridge extension to hit localhost APIs.")
		}
		return nil, "", internal.NewInternalError("Bridge request failed: " + err.Error())
	}

	// Build response headers slice (same shape as direct path)
	respHeaders := []KeyValuePair{}
	for k, v := range bridgeResp.Headers {
		respHeaders = append(respHeaders, KeyValuePair{Key: k, Value: v})
	}

	result := &ExecutionResult{
		StatusCode:        bridgeResp.Status,
		StatusText:        http.StatusText(bridgeResp.Status),
		ResponseTimeMs:    bridgeResp.Duration,
		ResponseSizeBytes: len(bridgeResp.BodyString()),
		Headers:           respHeaders,
		Body:              bridgeResp.BodyString(),
	}

	return result, bridgeResp.RequestID, nil
}

// executeDirect is your original http.Client logic, completely unchanged.
func (s *ExecutionService) executeDirect(
	ctx context.Context,
	reqModel *requests.APIRequest,
	finalURL string,
	headers map[string]string,
) (*ExecutionResult, error) {

	var bodyReader io.Reader
	if reqModel.Body != "" {
		bodyReader = strings.NewReader(reqModel.Body)
	}

	proxyReq, err := http.NewRequestWithContext(ctx, reqModel.Method, finalURL, bodyReader)
	if err != nil {
		return nil, internal.NewInternalError("Failed to construct proxy request: " + err.Error())
	}

	for _, h := range reqModel.Headers {
		if strings.ToLower(h.Key) == "host" {
			proxyReq.Host = h.Value
		} else {
			proxyReq.Header.Set(h.Key, h.Value)
		}
	}

	// Re-apply the auth / user-agent / content-type defaults using the shared map
	// (already computed above, just set them on the request)
	for k, v := range headers {
		if proxyReq.Header.Get(k) == "" {
			proxyReq.Header.Set(k, v)
		}
	}

	q := proxyReq.URL.Query()
	for _, p := range reqModel.Params {
		q.Add(p.Key, p.Value)
	}
	proxyReq.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 30 * time.Second}
	start := time.Now()
	resp, err := client.Do(proxyReq)
	duration := time.Since(start)

	if err != nil {
		return nil, internal.NewInternalError("Request execution failed: " + err.Error())
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil && err != io.EOF {
		return nil, internal.NewInternalError("Failed to read response body")
	}

	respHeaders := []KeyValuePair{}
	for k, v := range resp.Header {
		val := ""
		if len(v) > 0 {
			val = strings.Join(v, ", ")
		}
		respHeaders = append(respHeaders, KeyValuePair{Key: k, Value: val})
	}

	return &ExecutionResult{
		StatusCode:        resp.StatusCode,
		StatusText:        http.StatusText(resp.StatusCode),
		ResponseTimeMs:    duration.Milliseconds(),
		ResponseSizeBytes: len(respBodyBytes),
		Headers:           respHeaders,
		Body:              string(respBodyBytes),
	}, nil
}

// ── History + GetHistory + ClearHistory unchanged below ──────────────────────

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

func (s *ExecutionService) ClearHistory(ctx context.Context, userID string) error {
	err := s.historyRepo.DeleteAllByUserID(ctx, userID)
	if err != nil {
		return internal.NewInternalError("Failed to clear history")
	}
	return nil
}