// Package httphandlers provides HTTP handlers for the API.
package httphandlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/poyrazk/thecloud/pkg/crypto"
	"github.com/poyrazk/thecloud/pkg/httputil"
)

// StorageHandler handles object storage HTTP endpoints.
type StorageHandler struct {
	svc ports.StorageService
	cfg *platform.Config
}

// NewStorageHandler constructs a StorageHandler.
func NewStorageHandler(svc ports.StorageService, cfg *platform.Config) *StorageHandler {
	return &StorageHandler{
		svc: svc,
		cfg: cfg,
	}
}

const (
	errInvalidUploadID = "invalid upload id"
)

// Upload uploads an object to a bucket
// @Summary Upload an object
// @Description Uploads a file/object to the specified bucket and key
// @Tags storage
// @Accept octet-stream
// @Produce json
// @Security APIKeyAuth
// @Param bucket path string true "Bucket name"
// @Param key path string true "Object key"
// @Param file formData file true "File to upload"
// @Success 201 {object} domain.Object
// @Failure 400 {object} httputil.Response
// @Router /storage/{bucket}/{key} [put]
func (h *StorageHandler) Upload(c *gin.Context) {
	bucket, key, ok := getBucketAndKeyRequired(c)
	if !ok {
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

// Download downloads an object from a bucket
// @Summary Download an object
// @Description Streams the specified object as an attachment
// @Tags storage
// @Produce octet-stream
// @Security APIKeyAuth
// @Param bucket path string true "Bucket name"
// @Param key path string true "Object key"
// @Success 200 {file} file "Object content"
// @Failure 404 {object} httputil.Response
// @Router /storage/{bucket}/{key} [get]
func (h *StorageHandler) Download(c *gin.Context) {
	bucket, key, ok := getBucketAndKeyRequired(c)
	if !ok {
		return
	}
	versionID := c.Query("versionId")

	var reader io.ReadCloser
	var obj *domain.Object
	var err error

	if versionID != "" {
		reader, obj, err = h.svc.DownloadVersion(c.Request.Context(), bucket, key, versionID)
	} else {
		reader, obj, err = h.svc.Download(c.Request.Context(), bucket, key)
	}

	if err != nil {
		httputil.Error(c, err)
		return
	}
	defer func() { _ = reader.Close() }()

	// Set headers
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", key))
	c.Header("Content-Type", obj.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", obj.SizeBytes))

	// Stream file to client
	_, _ = io.Copy(c.Writer, reader)
}

// List returns objects in a bucket
// @Summary List objects in a bucket
// @Description Gets a list of all objects within a specific bucket
// @Tags storage
// @Produce json
// @Security APIKeyAuth
// @Param bucket path string true "Bucket name"
// @Success 200 {array} domain.Object
// @Failure 404 {object} httputil.Response
// @Router /storage/{bucket} [get]
func (h *StorageHandler) List(c *gin.Context) {
	bucket, ok := getBucket(c)
	if !ok {
		return
	}

	objects, err := h.svc.ListObjects(c.Request.Context(), bucket)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, objects)
}

// Delete deletes an object from a bucket
// @Summary Delete an object
// @Description Removes an object from the specified bucket
// @Tags storage
// @Produce json
// @Security APIKeyAuth
// @Param bucket path string true "Bucket name"
// @Param key path string true "Object key"
// @Success 204
// @Failure 404 {object} httputil.Response
// @Router /storage/{bucket}/{key} [delete]
func (h *StorageHandler) Delete(c *gin.Context) {
	bucket, key, ok := getBucketAndKeyRequired(c)
	if !ok {
		return
	}
	versionID := c.Query("versionId")

	var err error
	if versionID != "" {
		err = h.svc.DeleteVersion(c.Request.Context(), bucket, key, versionID)
	} else {
		err = h.svc.DeleteObject(c.Request.Context(), bucket, key)
	}

	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusNoContent, nil)
}

// CreateBucket creates a new bucket
// @Summary Create a bucket
// @Description Creates a new storage bucket
// @Tags storage
// @Accept json
// @Produce json
// @Security APIKeyAuth
// @Param request body domain.Bucket true "Bucket creation request"
// @Success 201 {object} domain.Bucket
// @Failure 400 {object} httputil.Response
// @Router /storage/buckets [post]
func (h *StorageHandler) CreateBucket(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		IsPublic bool   `json:"is_public"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidRequestBody))
		return
	}

	bucket, err := h.svc.CreateBucket(c.Request.Context(), req.Name, req.IsPublic)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, bucket)
}

// ListBuckets lists all buckets for the user
// @Summary List buckets
// @Description Lists all storage buckets owned by the user
// @Tags storage
// @Produce json
// @Security APIKeyAuth
// @Success 200 {array} domain.Bucket
// @Failure 500 {object} httputil.Response
// @Router /storage/buckets [get]
func (h *StorageHandler) ListBuckets(c *gin.Context) {
	buckets, err := h.svc.ListBuckets(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, buckets)
}

// DeleteBucket deletes a bucket
// @Summary Delete a bucket
// @Description Deletes a storage bucket
// @Tags storage
// @Produce json
// @Security APIKeyAuth
// @Param bucket path string true "Bucket name"
// @Success 204
// @Failure 400 {object} httputil.Response
// @Router /storage/buckets/{bucket} [delete]
func (h *StorageHandler) DeleteBucket(c *gin.Context) {
	bucket := c.Param("bucket")
	if bucket == "" {
		httputil.Error(c, errors.New(errors.InvalidInput, "bucket name required"))
		return
	}

	if err := h.svc.DeleteBucket(c.Request.Context(), bucket); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusNoContent, nil)
}

// GetClusterStatus returns the current state of the storage cluster
// @Summary Get storage cluster status
// @Description Returns the status of all nodes in the distributed storage cluster
// @Tags storage
// @Produce json
// @Security APIKeyAuth
// @Success 200 {object} domain.StorageCluster
// @Router /storage/cluster/status [get]
func (h *StorageHandler) GetClusterStatus(c *gin.Context) {
	status, err := h.svc.GetClusterStatus(c.Request.Context())
	if err != nil {
		httputil.Error(c, err)
		return
	}
	httputil.Success(c, http.StatusOK, status)
}

// InitiateMultipartUpload initiates a new multipart upload
// @Summary Initiate multipart upload
// @Description Creates a new multipart upload session for a bucket and key
// @Tags storage
// @Produce json
// @Security APIKeyAuth
// @Param bucket path string true "Bucket name"
// @Param key path string true "Object key"
// @Success 201 {object} domain.MultipartUpload
// @Router /storage/{bucket}/{key}/multipart [post]
func (h *StorageHandler) InitiateMultipartUpload(c *gin.Context) {
	bucket, key, ok := getBucketAndKeyRequired(c)
	if !ok {
		return
	}

	upload, err := h.svc.CreateMultipartUpload(c.Request.Context(), bucket, key)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, upload)
}

// UploadPart uploads a single part of an ongoing multipart upload
// @Summary Upload a part
// @Description Uploads a data chunk (part) for the specified multipart upload
// @Tags storage
// @Accept octet-stream
// @Produce json
// @Security APIKeyAuth
// @Param id path string true "Upload ID"
// @Param part query int true "Part number"
// @Success 200 {object} domain.Part
// @Router /storage/multipart/:id/parts [put]
func (h *StorageHandler) UploadPart(c *gin.Context) {
	idStr := c.Param("id")
	uploadID, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidUploadID))
		return
	}

	partNumStr := c.Query("part")
	partNumber, err := strconv.Atoi(partNumStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid part number"))
		return
	}

	part, err := h.svc.UploadPart(c.Request.Context(), uploadID, partNumber, c.Request.Body)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, part)
}

// CompleteMultipartUpload completes a multipart upload
// @Summary Complete multipart upload
// @Description Assembles all uploaded parts into the final object
// @Tags storage
// @Produce json
// @Security APIKeyAuth
// @Param id path string true "Upload ID"
// @Success 200 {object} domain.Object
// @Router /storage/multipart/:id/complete [post]
func (h *StorageHandler) CompleteMultipartUpload(c *gin.Context) {
	idStr := c.Param("id")
	uploadID, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidUploadID))
		return
	}

	obj, err := h.svc.CompleteMultipartUpload(c.Request.Context(), uploadID)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, obj)
}

// AbortMultipartUpload aborts a multipart upload
// @Summary Abort multipart upload
// @Description Cancels the upload and deletes all uploaded parts
// @Tags storage
// @Produce json
// @Security APIKeyAuth
// @Param id path string true "Upload ID"
// @Success 204
// @Router /storage/multipart/:id [delete]
func (h *StorageHandler) AbortMultipartUpload(c *gin.Context) {
	idStr := c.Param("id")
	uploadID, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidUploadID))
		return
	}

	if err := h.svc.AbortMultipartUpload(c.Request.Context(), uploadID); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusNoContent, nil)
}

// GeneratePresignedURL creates a time-limited signed URL
func (h *StorageHandler) GeneratePresignedURL(c *gin.Context) {
	bucket, key, ok := getBucketAndKeyRequired(c)
	if !ok {
		return
	}

	var req struct {
		Method    string `json:"method"` // GET or PUT
		ExpirySec int    `json:"expiry_seconds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, errInvalidRequestBody))
		return
	}

	method := req.Method
	if method == "" {
		method = http.MethodGet
	}
	if method != http.MethodGet && method != http.MethodPut {
		httputil.Error(c, errors.New(errors.InvalidInput, "only GET and PUT methods supported"))
		return
	}

	presigned, err := h.svc.GeneratePresignedURL(c.Request.Context(), bucket, key, method, 0)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, presigned)
}

// ServePresignedDownload handles object download via signed URL (no auth needed)
func (h *StorageHandler) ServePresignedDownload(c *gin.Context) {
	bucket, key, ok := getBucketAndKeyRequired(c)
	if !ok {
		return
	}
	expires := c.Query("expires")
	signature := c.Query("signature")

	// Verify signature
	// Note: We need the secret key here to verify.
	// Ideally the service handles verification, but the domain model is cleaner if we verify here
	// OR delegate to a service method `VerifyPresignedAccess`.
	// For now, let's verify using the shared secret.
	secret := h.cfg.StorageSecret // Use configured secret
	if secret == "" {
		httputil.Error(c, errors.New(errors.Internal, "storage secret not configured"))
		return
	}
	path := fmt.Sprintf("/storage/presigned/%s/%s", bucket, key)

	if err := crypto.VerifyURL(secret, http.MethodGet, path, expires, signature); err != nil {
		httputil.Error(c, errors.New(errors.Forbidden, "invalid or expired signature"))
		return
	}

	// Bypass normal Download auth checks?
	// The problem is `svc.Download` might assume checking permissions via Context's UserID.
	// But `Download` implementation currently only checks repository metadata.
	// We need to bypass the 'AuditLog' or 'RBAC' that might check UserID in context.
	// Since `svc.Download` is mostly pure logic, it should work if we pass background context
	// but we lose audit of *who* generated the link.

	reader, obj, err := h.svc.Download(c.Request.Context(), bucket, key)
	if err != nil {
		httputil.Error(c, err)
		return
	}
	defer func() { _ = reader.Close() }()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", key))
	c.Header("Content-Type", obj.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", obj.SizeBytes))
	_, _ = io.Copy(c.Writer, reader)
}

// ServePresignedUpload handles object upload via signed URL (no auth needed)
func (h *StorageHandler) ServePresignedUpload(c *gin.Context) {
	bucket, key, ok := getBucketAndKeyRequired(c)
	if !ok {
		return
	}
	expires := c.Query("expires")
	signature := c.Query("signature")

	secret := h.cfg.StorageSecret
	if secret == "" {
		httputil.Error(c, errors.New(errors.Internal, "storage secret not configured"))
		return
	}
	path := fmt.Sprintf("/storage/presigned/%s/%s", bucket, key)

	if err := crypto.VerifyURL(secret, http.MethodPut, path, expires, signature); err != nil {
		httputil.Error(c, errors.New(errors.Forbidden, "invalid or expired signature"))
		return
	}

	// For anonymous uploads, we might ideally need a dummy user ID or the owner's ID embedded in the token.
	// Current Upload implementation grabs UserID from context.
	// We need to support "system" or "anonymous" upload.

	// FIX: The current `svc.Upload` derives UserID from context.
	// We should probably allow the context to carry a "PresignedUser" or similar.
	// or modify Upload to accept nil UserID.

	// For now, let's proceed and see:
	// `Upload` -> `appcontext.UserIDFromContext(ctx)` which returns uuid.Nil if not found.
	// The Repository `SaveMeta` saves this UserID. It's valid to have Nil (0000...) for system/anon uploads?
	// It's acceptable for this feature.

	obj, err := h.svc.Upload(c.Request.Context(), bucket, key, c.Request.Body)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusCreated, obj)
}

// SetBucketVersioning toggles versioning for a bucket
// @Summary Set bucket versioning
// @Description Enables or disables object versioning for the specified bucket
// @Tags storage
// @Accept json
// @Produce json
// @Security APIKeyAuth
// @Param bucket path string true "Bucket name"
// @Param request body object true "Versioning request"
// @Success 200
// @Router /storage/buckets/{bucket}/versioning [patch]
func (h *StorageHandler) SetBucketVersioning(c *gin.Context) {
	bucket, ok := getBucket(c)
	if !ok {
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, errors.New(errors.InvalidInput, "invalid request body"))
		return
	}

	if err := h.svc.SetBucketVersioning(c.Request.Context(), bucket, req.Enabled); err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, gin.H{"status": "updated"})
}

// ListVersions returns all versions of an object
// @Summary List object versions
// @Description Gets a list of all versions of a specific object
// @Tags storage
// @Produce json
// @Security APIKeyAuth
// @Param bucket path string true "Bucket name"
// @Param key path string true "Object key"
// @Success 200 {array} domain.Object
// @Router /storage/versions/{bucket}/{key} [get]
func (h *StorageHandler) ListVersions(c *gin.Context) {
	bucket, key, ok := getBucketAndKeyRequired(c)
	if !ok {
		return
	}

	versions, err := h.svc.ListVersions(c.Request.Context(), bucket, key)
	if err != nil {
		httputil.Error(c, err)
		return
	}

	httputil.Success(c, http.StatusOK, versions)
}
