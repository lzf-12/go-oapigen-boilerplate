package errlib

import (
	"fmt"
	"log"
	"net/http"
)

type AppError struct {
	Code    string
	Message string
	Status  int
	Details map[string]interface{}
}

func NewAppErrorWithLog(err error, code string) *AppError {
	log.Println(fmt.Errorf("error: %w", err))
	return NewAppError(code)
}

// helper to create errors from registry
func NewAppError(code string) *AppError {
	if template, exists := errorRegistry[code]; exists {
		return &AppError{
			Code:    template.Code,
			Message: template.Message,
			Status:  template.Status,
			Details: nil,
		}
	}

	// fallback if not defined in error registry
	return &AppError{
		Code:    ErrCodeInternalServer,
		Message: "An internal server error occurred",
		Status:  http.StatusInternalServerError,
	}
}

// create error with additional details
func NewAppErrorWithDetails(code string, details map[string]interface{}) *AppError {
	err := NewAppError(code)
	err.Details = details
	return err
}

// get error message
func (e *AppError) Error() string {
	return e.Message
}

// common error without detail function
func ErrUserNotFound() *AppError            { return NewAppError(ErrCodeUserNotFound) }
func ErrInvalidEmailrOrPassword() *AppError { return NewAppError(ErrCodeInvalidEmailOrPassword) }
func ErrInvalidInput() *AppError            { return NewAppError(ErrCodeInvalidInput) }
func ErrUnauthorized() *AppError            { return NewAppError(ErrCodeUnauthorized) }
func ErrForbidden() *AppError               { return NewAppError(ErrCodeForbidden) }
func ErrInternalServer() *AppError          { return NewAppError(ErrCodeInternalServer) }
func ErrRateLimited() *AppError             { return NewAppError(ErrCodeRateLimited) }
func ErrDBConnection() *AppError            { return NewAppError(ErrCodeDBConnection) }
func ErrDBQuery() *AppError                 { return NewAppError(ErrCodeDBQuery) }
func ErrDBTransaction() *AppError           { return NewAppError(ErrCodeDBTransaction) }
func ErrDBConstraint() *AppError            { return NewAppError(ErrCodeDBConstraint) }
func ErrDBDuplicate() *AppError             { return NewAppError(ErrCodeDBDuplicate) }
func ErrDBTimeout() *AppError               { return NewAppError(ErrCodeDBTimeout) }
func ErrStorageNotFound() *AppError         { return NewAppError(ErrCodeStorageNotFound) }
func ErrStorageAccess() *AppError           { return NewAppError(ErrCodeStorageAccess) }
