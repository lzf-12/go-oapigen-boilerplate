package errlib

import (
	"github.com/gin-gonic/gin"
)

func ErrorHandlerGinMiddleware(eh ErrorHandler) gin.HandlerFunc {
	return func(c *gin.Context) {

		// process request
		c.Next()

		// after request processed, check if there is error
		var err *gin.Error
		if len(c.Errors) > 0 {
			err = c.Errors.Last()
		}

		if err == nil {
			// continue
			return
		} else {
			// override error response if error is not nil
			eh.HandleAndSendErrorResponse(c.Writer, c.Request, err.Err)
		}
	}
}
