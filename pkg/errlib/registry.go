package errlib

import "net/http"

var errorRegistry = map[string]AppError{
	// auth/user
	ErrCodeUserNotFound: {
		Code:    ErrCodeUserNotFound,
		Message: "User not found",
		Status:  http.StatusNotFound,
	},
	ErrCodeInvalidEmailOrPassword: {
		Code:    ErrCodeInvalidEmailOrPassword,
		Message: "Invalid email or password",
		Status:  http.StatusUnauthorized,
	},
	ErrCodeTokenExpired: {
		Code:    ErrCodeTokenExpired,
		Message: "Token has expired",
		Status:  http.StatusUnauthorized,
	},
	ErrCodeUnauthorized: {
		Code:    ErrCodeUnauthorized,
		Message: "Authentication required",
		Status:  http.StatusUnauthorized,
	},
	ErrCodeAccessDenied: {
		Code:    ErrCodeAccessDenied,
		Message: "Access denied",
		Status:  http.StatusForbidden,
	},
	ErrCodeForbidden: {
		Code:    ErrCodeForbidden,
		Message: "Insufficient permissions",
		Status:  http.StatusForbidden,
	},

	// validation/parsing
	ErrCodeInvalidInput: {
		Code:    ErrCodeInvalidInput,
		Message: "Invalid input provided",
		Status:  http.StatusBadRequest,
	},
	ErrCodeValidation: {
		Code:    ErrCodeValidation,
		Message: "Validation error",
		Status:  http.StatusBadRequest,
	},
	ErrCodeJSONUnmarshal: {
		Code:    ErrCodeJSONUnmarshal,
		Message: "Failed to parse JSON",
		Status:  http.StatusBadRequest,
	},
	ErrCodeJSONSyntax: {
		Code:    ErrCodeJSONSyntax,
		Message: "Invalid JSON syntax",
		Status:  http.StatusBadRequest,
	},

	// general system errors
	ErrCodeInternalServer: {
		Code:    ErrCodeInternalServer,
		Message: "An internal server error occurred",
		Status:  http.StatusInternalServerError,
	},
	ErrCodeRateLimited: {
		Code:    ErrCodeRateLimited,
		Message: "Too many requests",
		Status:  http.StatusTooManyRequests,
	},

	// db errors
	ErrCodeDBConnection: {
		Code:    ErrCodeDBConnection,
		Message: "Database connection failed",
		Status:  http.StatusServiceUnavailable,
	},
	ErrCodeDBQuery: {
		Code:    ErrCodeDBQuery,
		Message: "Database query failed",
		Status:  http.StatusInternalServerError,
	},
	ErrCodeDBTransaction: {
		Code:    ErrCodeDBTransaction,
		Message: "Database transaction failed",
		Status:  http.StatusInternalServerError,
	},
	ErrCodeDBConstraint: {
		Code:    ErrCodeDBConstraint,
		Message: "Database constraint violation",
		Status:  http.StatusConflict,
	},
	ErrCodeDBDuplicate: {
		Code:    ErrCodeDBDuplicate,
		Message: "Duplicate entry",
		Status:  http.StatusConflict,
	},
	ErrCodeDBTimeout: {
		Code:    ErrCodeDBTimeout,
		Message: "Database operation timed out",
		Status:  http.StatusGatewayTimeout,
	},

	// storage errors
	ErrCodeStorageNotFound: {
		Code:    ErrCodeStorageNotFound,
		Message: "Storage resource not found",
		Status:  http.StatusNotFound,
	},
	ErrCodeStorageAccess: {
		Code:    ErrCodeStorageAccess,
		Message: "Storage access denied",
		Status:  http.StatusForbidden,
	},

	// generic data error
	ErrCodeDataNotFound: {
		Code:    ErrCodeDataNotFound,
		Message: "Requested data not found",
		Status:  http.StatusNotFound,
	},
}
