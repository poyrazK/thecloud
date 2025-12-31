package httputil

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/poyraz/cloud/internal/errors"
)

type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error interface{} `json:"error,omitempty"`
}

func Success(c *gin.Context, code int, data interface{}) {
	c.JSON(code, Response{Data: data})
}

func Error(c *gin.Context, err error) {
	var e errors.Error
	if apiErr, ok := err.(errors.Error); ok {
		e = apiErr
	} else {
		e = errors.Error{
			Type:    errors.Internal,
			Message: "An unexpected error occurred",
			Cause:   err,
		}
	}

	statusCode := http.StatusInternalServerError
	switch e.Type {
	case errors.NotFound:
		statusCode = http.StatusNotFound
	case errors.InvalidInput:
		statusCode = http.StatusBadRequest
	case errors.Unauthorized:
		statusCode = http.StatusUnauthorized
	case errors.Conflict:
		statusCode = http.StatusConflict
	}

	c.JSON(statusCode, Response{Error: e})
}
