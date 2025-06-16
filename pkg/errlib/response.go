package errlib

// based on RFC 7807
type ErrorResponse struct {
	Type      string                 `json:"type"`               // URI reference for error type
	Title     string                 `json:"title"`              // Human-readable summary
	Status    int                    `json:"status"`             // HTTP status code
	Detail    string                 `json:"detail,omitempty"`   // Human-readable explanation
	Instance  string                 `json:"instance,omitempty"` // URI reference for specific occurrence
	Timestamp string                 `json:"timestamp"`          // RFC3339 timestamp
	TraceID   string                 `json:"trace_id,omitempty"` // distributed tracing
	Errors    map[string]interface{} `json:"errors,omitempty"`   // Field-specific validation errors
}
