package common

import "oapi-to-rest/pkg/errlib"

func ConvertToStandardErrorResponse(er errlib.ErrorResponse) StandardErrorResponse {
	return StandardErrorResponse{
		Type:      &er.Type,
		Details:   &er.Detail,
		Title:     &er.Title,
		Status:    &er.Status,
		Instance:  &er.Instance,
		Timestamp: &er.Timestamp,
		TraceId:   &er.TraceID,
		Errors:    &er.Errors,
	}
}
