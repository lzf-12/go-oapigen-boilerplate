package errlib

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"oapi-to-rest/pkg/errlib/trace"
	"time"

	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
)

const (
	postgres = "postgres"
	sqlite   = "sqlite"
)

// errorHandler centralizes error response handling
type ErrorHandler struct {
	debug         bool
	defaultErrRef string
}

func NewErrorHandler(debug bool) *ErrorHandler {
	return &ErrorHandler{debug: debug}
}

func (eh *ErrorHandler) HandleError(r *http.Request, err error) ErrorResponse {

	var appErr *AppError
	var status int
	var errResp ErrorResponse
	var dbErrRef string
	var isDBErr bool

	// type assertion to determine error type
	switch e := err.(type) {
	case *AppError:
		appErr = e
		status = e.Status
	case *json.UnmarshalTypeError:
		appErr = NewAppErrorWithDetails(ErrCodeJSONUnmarshal, map[string]interface{}{
			"field": e.Field,
			"type":  e.Type.String(),
		})
		status = http.StatusBadRequest
	case *json.SyntaxError:
		appErr = NewAppError(ErrCodeJSONSyntax)
		status = http.StatusBadRequest
	default:
		// check for database/storage specific errors
		if dbErr := eh.handleDatabaseError(err); dbErr != nil {

			isDBErr = true
			appErr = dbErr
			status = dbErr.Status

			if dbErr.Details != nil {
				if db, ok := dbErr.Details["database"]; ok {
					switch db {
					case "postgres":
						dbErrRef = "https://www.postgresql.org/docs/current/errcodes-appendix.html"
					case "sqlite":
						dbErrRef = "https://www.sqlite.org/rescode.html"
					default:
						dbErrRef = ""
					}
				}
			}

		} else {
			// log unexpected errors for debugging
			fmt.Printf("unexpected error: %v\n", err)
			appErr = ErrInternalServer()
			status = http.StatusInternalServerError
		}
	}

	// build error response
	errResp = ErrorResponse{
		Type:      eh.defaultErrRef,
		Title:     appErr.Message,
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if r != nil {
		errResp.Instance = r.URL.Path
		errResp.TraceID = trace.GetTraceIDFromContext(r.Context())
	}

	// replace error reference with default value if empty
	if isDBErr {
		errResp.Type = dbErrRef
	}

	// add detail to response if debug enabled
	if eh.debug && appErr.Details != nil {
		errResp.Errors = appErr.Details
	}

	// force include details if client 4xx errors
	if status >= 400 && status < 500 && appErr.Details != nil {
		errResp.Detail = fmt.Sprintf("%v", appErr.Details)
	}

	return errResp
}

func (eh *ErrorHandler) HandleAndSendErrorResponse(w http.ResponseWriter, r *http.Request, err error) {

	errResp := eh.HandleError(r, err)

	// response headers
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(errResp.Status)

	// encode and send response
	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		fmt.Printf("failed to encode error response: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (eh *ErrorHandler) handleDatabaseError(err error) *AppError {

	// handle sql.ErrNoRows specifically
	if err == sql.ErrNoRows {
		return NewAppError(ErrCodeStorageNotFound)
	}

	// postgres specific error
	if pqErr, ok := err.(*pq.Error); ok {
		return eh.handlePostgreSQLError(pqErr)
	}

	// sqlite specific error
	if sqliteErr, ok := err.(sqlite3.Error); ok {
		return eh.handleSQLiteError(sqliteErr)
	}

	// if no specific database error pattern matched, return nil
	// default error handler (not db handler) will takeover
	return nil
}

func (eh *ErrorHandler) handlePostgreSQLError(pqErr *pq.Error) *AppError {
	details := map[string]interface{}{
		"database":   postgres,
		"error_code": pqErr.Code,
		"error_name": pqErr.Code.Name(),
		"message":    pqErr.Message,
		"detail":     pqErr.Detail,
		"hint":       pqErr.Hint,
		"position":   pqErr.Position,
		"table":      pqErr.Table,
		"column":     pqErr.Column,
		"constraint": pqErr.Constraint,
	}

	// remove empty fields
	for key, value := range details {
		if value == "" || value == nil {
			delete(details, key)
		}
	}

	// reference: https://www.postgresql.org/docs/current/errcodes-appendix.html
	switch pqErr.Code.Class() {
	case "08": // connection
		return NewAppErrorWithDetails(ErrCodeDBConnection, details)

	case "23": // constraint
		switch pqErr.Code {
		case "23505": // unique_violation
			return NewAppErrorWithDetails(ErrCodeDBDuplicate, details)
		case "23503", "23502", "23514": // foreign_key_violation, not_null_violation, check_violation
			return NewAppErrorWithDetails(ErrCodeDBConstraint, details)
		default:
			return NewAppErrorWithDetails(ErrCodeDBConstraint, details)
		}

	case "25": // transaction state
		return NewAppErrorWithDetails(ErrCodeDBTransaction, details)

	case "57": // operator intervention (timeout, cancel, etc.)
		return NewAppErrorWithDetails(ErrCodeDBTimeout, details)

	case "53": // insufficient resources
		return NewAppErrorWithDetails(ErrCodeDBConnection, details)

	case "42":
		return NewAppErrorWithDetails(ErrCodeDBQuery, details)

	default: // unmapped
		return NewAppErrorWithDetails(ErrCodeDBQuery, details)
	}
}

func (eh *ErrorHandler) handleSQLiteError(sqliteErr sqlite3.Error) *AppError {
	details := map[string]interface{}{
		"database":      sqlite,
		"error_code":    int(sqliteErr.Code),
		"extended_code": int(sqliteErr.ExtendedCode),
		"message":       sqliteErr.Error(),
	}

	switch sqliteErr.Code {
	case sqlite3.ErrConstraint:
		switch sqliteErr.ExtendedCode {
		case sqlite3.ErrConstraintUnique, sqlite3.ErrConstraintPrimaryKey:
			return NewAppErrorWithDetails(ErrCodeDBDuplicate, details)
		case sqlite3.ErrConstraintForeignKey:
			return NewAppErrorWithDetails(ErrCodeDBConstraint, details)
		case sqlite3.ErrConstraintNotNull, sqlite3.ErrConstraintCheck:
			return NewAppErrorWithDetails(ErrCodeDBConstraint, details)
		default:
			return NewAppErrorWithDetails(ErrCodeDBConstraint, details)
		}

	case sqlite3.ErrBusy, sqlite3.ErrLocked:
		return NewAppErrorWithDetails(ErrCodeDBTimeout, details)

	case sqlite3.ErrCantOpen, sqlite3.ErrNotADB, sqlite3.ErrCorrupt:
		return NewAppErrorWithDetails(ErrCodeDBConnection, details)

	case sqlite3.ErrPerm, sqlite3.ErrAuth:
		return NewAppErrorWithDetails(ErrCodeStorageAccess, details)

	case sqlite3.ErrNotFound: // sqlite file not found
		return NewAppErrorWithDetails(ErrCodeStorageNotFound, details)

	case sqlite3.ErrRange, sqlite3.ErrMisuse, sqlite3.ErrNoLFS:
		return NewAppErrorWithDetails(ErrCodeDBQuery, details)

	default: // unmapped
		return NewAppErrorWithDetails(ErrCodeDBQuery, details)
	}
}
