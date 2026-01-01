package httputil

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/poyraz/cloud/internal/errors"
)

type Meta struct {
	RequestID string `json:"request_id,omitempty"`
	Timestamp string `json:"timestamp"`
}

type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error interface{} `json:"error,omitempty"`
	Meta  *Meta       `json:"meta,omitempty"`
}

func Success(c *gin.Context, code int, data interface{}) {
	requestID, _ := c.Get("requestID")
	reqIDStr, _ := requestID.(string)

	c.JSON(code, Response{
		Data: data,
		Meta: &Meta{
			RequestID: reqIDStr,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func Error(c *gin.Context, err error) {
	var e errors.Error
	if apiErr, ok := err.(errors.Error); ok {
		e = apiErr
	} else {
		e = errors.Error{
			Type:    errors.Internal,
			Message: "An unexpected error occurred",
			Code:    string(errors.Internal),
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
	case errors.BucketNotFound, errors.ObjectNotFound:
		statusCode = http.StatusNotFound
	case errors.ObjectTooLarge:
		statusCode = http.StatusRequestEntityTooLarge
	case errors.InstanceNotRunning, errors.PortConflict, errors.TooManyPorts:
		statusCode = http.StatusConflict
	}

	requestID, _ := c.Get("requestID")
	reqIDStr, _ := requestID.(string)

	c.JSON(statusCode, Response{
		Error: e,
		Meta: &Meta{
			RequestID: reqIDStr,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	})
}
