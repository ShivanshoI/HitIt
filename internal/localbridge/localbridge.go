// Package localbridge routes HTTP requests through a browser extension
// to reach localhost APIs from a deployed backend.
//
// Drop this file into internal/localbridge/localbridge.go
// No external dependencies — uses only stdlib.
package localbridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client sends requests to the bridge server, which forwards them
// through the browser extension to localhost.
type Client struct {
	serverURL  string
	authToken  string
	timeout    time.Duration
	httpClient *http.Client
}

// Config for the bridge client.
type Config struct {
	// ServerURL is the HTTP base URL of the deployed bridge server.
	// e.g. "http://localhost:3100" or "https://bridge.yourapp.com"
	ServerURL string

	// AuthToken must match AUTH_TOKEN on the bridge server (optional).
	AuthToken string

	// Timeout for each proxied request. Defaults to 30s.
	Timeout time.Duration
}

// New creates a LocalBridge client. Call once at startup and reuse.
func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &Client{
		serverURL: cfg.ServerURL,
		authToken: cfg.AuthToken,
		timeout:   cfg.Timeout,
		httpClient: &http.Client{
			// Give a little extra over the proxied timeout for network overhead
			Timeout: cfg.Timeout + 5*time.Second,
		},
	}
}

// ─── Request / Response types ─────────────────────────────────────────────────

// BridgeRequest is what we send to the bridge server's POST /request.
type BridgeRequest struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers,omitempty"`
	Body      string            `json:"body,omitempty"`
	TimeoutMs int64             `json:"timeoutMs"`
}

// BridgeResponse is what the bridge server returns.
type BridgeResponse struct {
	RequestID string            `json:"requestId"` // UUID from bridge, store alongside your Mongo _id
	Status    int               `json:"status"`
	Headers   map[string]string `json:"headers"`
	Body      json.RawMessage   `json:"body"`
	Duration  int64             `json:"duration"` // ms, as measured by the extension
	Error     string            `json:"error,omitempty"`
}

// BodyString returns the response body as a plain string.
// Works whether the upstream returned JSON or plain text.
func (r *BridgeResponse) BodyString() string {
	if len(r.Body) == 0 {
		return ""
	}
	// If it's a JSON string (quoted), unwrap it
	var s string
	if err := json.Unmarshal(r.Body, &s); err == nil {
		return s
	}
	// Otherwise return raw JSON bytes as string
	return string(r.Body)
}

// ─── Status check ─────────────────────────────────────────────────────────────

// StatusResponse is returned by GET /status on the bridge server.
type StatusResponse struct {
	OK           bool `json:"ok"`
	Extensions   int  `json:"extensions"`
	HasExtension bool `json:"hasExtension"`
}

// Status checks whether the bridge server is reachable and has an extension connected.
// Call this on startup or before routing to the bridge to give clear errors.
func (c *Client) Status(ctx context.Context) (*StatusResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.serverURL+"/status", nil)
	if err != nil {
		return nil, fmt.Errorf("localbridge: build status request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("localbridge: bridge server unreachable at %s: %w", c.serverURL, err)
	}
	defer resp.Body.Close()

	var s StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, fmt.Errorf("localbridge: decode status: %w", err)
	}
	return &s, nil
}

// ─── Core Do ──────────────────────────────────────────────────────────────────

// Do sends a request through the bridge. headers may be nil.
func (c *Client) Do(ctx context.Context, method, url string, headers map[string]string, body string) (*BridgeResponse, error) {
	payload := BridgeRequest{
		Method:    method,
		URL:       url,
		Headers:   headers,
		Body:      body,
		TimeoutMs: c.timeout.Milliseconds(),
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("localbridge: marshal payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverURL+"/request", bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("localbridge: build http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.authToken != "" {
		httpReq.Header.Set("X-Bridge-Token", c.authToken)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("localbridge: call bridge server: %w", err)
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("localbridge: read bridge response: %w", err)
	}

	// Bridge-level errors (503 = no extension, 504 = timeout, 400 = bad request)
	if httpResp.StatusCode != http.StatusOK {
		var e struct {
			Error string `json:"error"`
			Code  string `json:"code"`
		}
		_ = json.Unmarshal(respBytes, &e)

		// Give a clear error when no extension is connected — common during dev
		if httpResp.StatusCode == http.StatusServiceUnavailable {
			return nil, fmt.Errorf("localbridge: no extension connected (open Chrome with LocalBridge extension): %s", e.Error)
		}
		return nil, fmt.Errorf("localbridge: bridge error %d: %s", httpResp.StatusCode, e.Error)
	}

	var result BridgeResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, fmt.Errorf("localbridge: unmarshal bridge response: %w", err)
	}

	return &result, nil
}
