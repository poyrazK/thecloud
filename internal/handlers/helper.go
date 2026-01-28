// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// parseUUID extracts a UUID from a path parameter and handles error reporting.
func parseUUID(c *gin.Context, paramName string) (*uuid.UUID, bool) {
	val := c.Param(paramName)
	if val == "" {
		httputil.Error(c, errors.New(errors.InvalidInput, paramName+" is required"))
		return nil, false
	}

	id, err := uuid.Parse(val)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid "+paramName+" format"))
		return nil, false
	}

	return &id, true
}

// getBucketAndKey extracts bucket and key from path params.
func getBucketAndKey(c *gin.Context) (bucket, key string, ok bool) {
	bucket = c.Param("bucket")
	key = c.Param("key")

	if bucket == "" {
		httputil.Error(c, errors.New(errors.InvalidInput, "bucket is required"))
		return "", "", false
	}

	// key can be empty for listing bucket
	return bucket, key, true
}
