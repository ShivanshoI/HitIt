package execution

type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ExecutionResult struct {
	StatusCode        int            `json:"status_code"`
	StatusText        string         `json:"status_text"`
	ResponseTimeMs    int64          `json:"response_time_ms"`
	ResponseSizeBytes int            `json:"response_size_bytes"`
	Headers           []KeyValuePair `json:"headers"`
	Body              string         `json:"body"`
}
