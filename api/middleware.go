package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"oapi-to-rest/specs/spec_validator"
	"strings"

	"github.com/gin-gonic/gin"
)

// request validator middleware
func RequestValidationMiddleware(msv *spec_validator.MultiSpecValidator) gin.HandlerFunc {
	return func(c *gin.Context) {

		// skipped path, continue without validation
		if msv.Config.Validation.SkipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// add route mapping
		for _, spec := range msv.Config.Specs {

			msv.AddRouteMapping(spec.BasePath, spec.Name)

		}

		// validate the request based on spec
		if errs, _ := msv.ValidateRequest(c.Request); len(errs) > 0 {
			msv.WriteMultiValidationError(c.Writer, c.Request, errs)
			c.Abort()
			return
		}

		c.Next()
	}
}

// responseRecorder with gin
type ResponseRecorder struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func NewResponseRecorder(w gin.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{
		ResponseWriter: w,
		body:           bytes.NewBuffer([]byte{}),
		status:         http.StatusOK,
	}
}

func (r *ResponseRecorder) Write(data []byte) (int, error) {
	r.body.Write(data)
	return r.ResponseWriter.Write(data)
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseRecorder) Status() int {
	return r.status
}

func (r *ResponseRecorder) Body() []byte {
	return r.body.Bytes()
}

// response validation log any validation error based on spec definen, actual response won't be changed.
func ResponseValidationMiddleware(msv *spec_validator.MultiSpecValidator) gin.HandlerFunc {
	return func(c *gin.Context) {

		// skip if response validation is not enabled
		if !msv.Config.Validation.ValidateResponses {
			c.Next()
			return
		}

		// create a response recorder to capture the response
		recorder := NewResponseRecorder(c.Writer)
		c.Writer = recorder

		// execute the handler first
		c.Next()

		// validate the response after handler execution
		if recorder.Body() != nil && len(recorder.Body()) > 0 {

			// create an http.Response for validation
			resp := &http.Response{
				StatusCode: recorder.Status(),
				Header:     recorder.Header(),
				Body:       io.NopCloser(strings.NewReader(string(recorder.Body()))),
			}

			// validate the response
			errs, err := msv.ValidateResponse(c.Request, resp)

			// log validation errors but don't modify the response
			if len(errs) > 0 {
				for _, e := range errs {
					fmt.Printf("response validation failed: %s\n", e.Message)
				}
			} else if err != nil {
				fmt.Printf("response validation error is not nil: %s\n", err.Error())
			}
		}

	}
}
