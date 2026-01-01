package httphandlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/poyraz/cloud/internal/core/ports"
	"github.com/poyraz/cloud/internal/errors"
	"github.com/poyraz/cloud/pkg/httputil"
)

type StorageHandler struct {
	svc ports.StorageService
}

func NewStorageHandler(svc ports.StorageService) *StorageHandler {
	return &StorageHandler{
		svc: svc,
	}
}

func (h *StorageHandler) Upload(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	if bucket == "" || key == "" {
		httputil.Error(c, errors.New(errors.InvalidInput, "bucket and key are required"))
		return
	}

	// Read from request body (stream)
	obj, err := h.svc.Upload(c.Request.Context(), bucket, key, c.Request.Body)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, obj)
}

func (h *StorageHandler) Download(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	reader, obj, err := h.svc.Download(c.Request.Context(), bucket, key)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	defer reader.Close()

	// Set headers
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", key))
	c.Header("Content-Type", obj.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", obj.SizeBytes))

	// Stream file to client
	_, _ = io.Copy(c.Writer, reader)
}

func (h *StorageHandler) List(c *gin.Context) {
	bucket := c.Param("bucket")

	objects, err := h.svc.ListObjects(c.Request.Context(), bucket)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, objects)
}

func (h *StorageHandler) Delete(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	if err := h.svc.DeleteObject(c.Request.Context(), bucket, key); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusNoContent, nil)
}
