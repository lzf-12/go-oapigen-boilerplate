package spec_validator

import (
	"encoding/json"
	"net/http"

	validatorError "github.com/pb33f/libopenapi-validator/errors"
)

type MultiValidationErrorResponse struct {
	Error                  string                            `json:"error"`
	Message                string                            `json:"message"`
	ValidationErrorDetails []*validatorError.ValidationError `json:"validation_error_details"`
	Path                   string                            `json:"path"`
	Method                 string                            `json:"method"`
	Count                  int                               `json:"error_count"`
}

// write response that used in request validation
func (msv *MultiSpecValidator) WriteMultiValidationError(w http.ResponseWriter, r *http.Request, errs []*validatorError.ValidationError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	var errCount int
	for _, e := range errs {
		for range e.SchemaValidationErrors {
			errCount++
		}
	}

	// create response
	response := MultiValidationErrorResponse{
		Error:                  "validation_failed",
		Message:                "request validation failed",
		ValidationErrorDetails: errs,
		Path:                   r.URL.Path,
		Method:                 r.Method,
		Count:                  int(errCount),
	}

	// encode and write response
	if err := json.NewEncoder(w).Encode(response); err != nil {

		// fallback to simple error response if JSON encoding fails
		http.Error(w, "request validation failed", http.StatusBadRequest)
	}
}
