package main

// ── HTTP Methods ────────────────────────────────────────────────────

type HTTPMethod string

const (
	MethodGet    HTTPMethod = "GET"
	MethodPost   HTTPMethod = "POST"
	MethodPut    HTTPMethod = "PUT"
	MethodPatch  HTTPMethod = "PATCH"
	MethodDelete HTTPMethod = "DELETE"
)

// ── Environment ─────────────────────────────────────────────────────

type Environment string

const (
	EnvLocal      Environment = "LOCAL"
	EnvStaging    Environment = "STAGING"
	EnvProduction Environment = "PRODUCTION"
)

// ── Example Status (service-specific enum kept globally for reference) ──

type ExampleStatus string

const (
	StatusActive   ExampleStatus = "active"
	StatusInactive ExampleStatus = "inactive"
)
