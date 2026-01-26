// Package services implements core business workflows.
package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/google/uuid"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/internal/platform"
	"github.com/poyrazk/thecloud/pkg/crypto"
)

const (
	errMultipartNotFound = "multipart upload not found"
	partPathFormat       = ".uploads/%s/part-%d"
	versionQueryFormat   = "%s?versionId=%s"
)

// StorageService manages object storage metadata and files.
type StorageService struct {
	repo       ports.StorageRepository
	store      ports.FileStore
	auditSvc   ports.AuditService
	encryptSvc ports.EncryptionService
	cfg        *platform.Config
}

// NewStorageService constructs a StorageService with its dependencies.
func NewStorageService(repo ports.StorageRepository, store ports.FileStore, auditSvc ports.AuditService, encryptSvc ports.EncryptionService, cfg *platform.Config) *StorageService {
	return &StorageService{
		repo:       repo,
		store:      store,
		auditSvc:   auditSvc,
		encryptSvc: encryptSvc,
		cfg:        cfg,
	}
}

func (s *StorageService) Upload(ctx context.Context, bucketName, key string, r io.Reader) (*domain.Object, error) {
	// 1. Check bucket versioning status
	bucket, err := s.repo.GetBucket(ctx, bucketName)
	if err != nil {
		return nil, err
	}

	versionID := "null" // Default version ID when versioning is disabled
	if bucket.VersioningEnabled {
		// Generate a timestamp-based version ID (reverse chronological)
		// 1<<62 is large enough to keep the result positive for a long time
		versionID = fmt.Sprintf("%d", (1<<62)-time.Now().UnixNano())
	}

	// 2. Write file to store
	// In the store, we'll prefix versions with versionID to avoid overwrites
	storeKey := key
	if bucket.VersioningEnabled {
		storeKey = fmt.Sprintf(versionQueryFormat, key, versionID)
	}

	// Encryption
	var finalReader io.Reader = r
	if bucket.EncryptionEnabled && s.encryptSvc != nil {
		// Read entire content to encrypt (streaming encryption is better but complex for now)
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		encryptedData, err := s.encryptSvc.Encrypt(ctx, bucketName, data)
		if err != nil {
			return nil, err
		}
		finalReader = bytes.NewReader(encryptedData)
	}

	size, err := s.store.Write(ctx, bucketName, storeKey, finalReader)
	if err != nil {
		return nil, err
	}

	// 3. Prepare metadata
	obj := &domain.Object{
		ID:          uuid.New(),
		UserID:      appcontext.UserIDFromContext(ctx),
		Bucket:      bucketName,
		Key:         key,
		VersionID:   versionID,
		IsLatest:    true,
		SizeBytes:   size,
		ContentType: "application/octet-stream", // In a real system we'd detect Content-Type
		CreatedAt:   time.Now(),
	}

	// Generate ARN
	// arn:thecloud:storage:local:default:object/<bucket>/<key>?versionId=<versionID>
	obj.ARN = fmt.Sprintf("arn:thecloud:storage:local:default:object/%s/%s", bucketName, key)
	if bucket.VersioningEnabled {
		obj.ARN += fmt.Sprintf("?versionId=%s", versionID)
	}

	// 4. Save metadata
	if err := s.repo.SaveMeta(ctx, obj); err != nil {
		// Cleanup file if DB save fails
		_ = s.store.Delete(ctx, bucketName, storeKey)
		return nil, err
	}

	_ = s.auditSvc.Log(ctx, obj.UserID, "storage.object_upload", "storage", obj.ID.String(), map[string]interface{}{
		"bucket":     obj.Bucket,
		"key":        obj.Key,
		"version_id": obj.VersionID,
	})

	// Metrics
	platform.StorageOperations.WithLabelValues("upload", bucketName, "success").Inc()
	platform.StorageBytesTransferred.WithLabelValues("upload").Add(float64(size))

	return obj, nil
}

func (s *StorageService) Download(ctx context.Context, bucket, key string) (io.ReadCloser, *domain.Object, error) {
	// 1. Get metadata
	obj, err := s.repo.GetMeta(ctx, bucket, key)
	if err != nil {
		return nil, nil, err
	}

	// 2. Open file
	reader, err := s.store.Read(ctx, bucket, key)
	if err != nil {
		platform.StorageOperations.WithLabelValues("download", bucket, "error").Inc()
		return nil, nil, err
	}

	// Decryption
	// We need to check if object is encrypted?
	// Currently we check bucket setting but valid approach is: if bucket has encryption enabled, we try to decrypt.
	// OR we assume everything in an encrypted bucket is encrypted.
	// Ideally Object metadata should flag "IsEncrypted".
	// For now, let's rely on checking the bucket config again or just if we can decrypt.
	// BUT wait, Read returns io.ReadCloser (stream). We implemented full read for encryption.
	// So we must read all, decrypt, wrap in Reader.

	// Check bucket status for efficiency (though key lookup is fast)
	b, err := s.repo.GetBucket(ctx, bucket)
	if err == nil && b.EncryptionEnabled && s.encryptSvc != nil {
		data, err := io.ReadAll(reader)
		reader.Close() // Close underlying file stream
		if err != nil {
			return nil, nil, err
		}

		decryptedData, err := s.encryptSvc.Decrypt(ctx, bucket, data)
		if err != nil {
			// Fallback: maybe it wasn't encrypted (legacy objects before encryption enabled)
			// In a real system we'd check metadata.
			// Re-wrap original data if decrypt failed? No, AES-GCM fails clearly.
			return nil, nil, err
		}
		reader = io.NopCloser(bytes.NewReader(decryptedData))
	}

	platform.StorageOperations.WithLabelValues("download", bucket, "success").Inc()
	if obj != nil {
		platform.StorageBytesTransferred.WithLabelValues("download").Add(float64(obj.SizeBytes))
	}

	return reader, obj, nil
}

func (s *StorageService) ListObjects(ctx context.Context, bucket string) ([]*domain.Object, error) {
	return s.repo.List(ctx, bucket)
}

func (s *StorageService) DownloadVersion(ctx context.Context, bucket, key, versionID string) (io.ReadCloser, *domain.Object, error) {
	// 1. Get metadata
	obj, err := s.repo.GetMetaByVersion(ctx, bucket, key, versionID)
	if err != nil {
		return nil, nil, err
	}

	// 2. Open file
	storeKey := key
	if obj.VersionID != "null" {
		storeKey = fmt.Sprintf(versionQueryFormat, key, obj.VersionID)
	}

	reader, err := s.store.Read(ctx, bucket, storeKey)
	if err != nil {
		return nil, nil, err
	}

	return reader, obj, nil
}

func (s *StorageService) ListVersions(ctx context.Context, bucket, key string) ([]*domain.Object, error) {
	return s.repo.ListVersions(ctx, bucket, key)
}

func (s *StorageService) DeleteVersion(ctx context.Context, bucket, key, versionID string) error {
	// 1. Get meta to verify ownership/existence
	if _, err := s.repo.GetMetaByVersion(ctx, bucket, key, versionID); err != nil {
		return err
	}

	// 2. Delete from store
	storeKey := key
	if versionID != "null" {
		storeKey = fmt.Sprintf(versionQueryFormat, key, versionID)
	}

	if err := s.store.Delete(ctx, bucket, storeKey); err != nil {
		return err
	}

	// 3. Delete meta (hard delete for specific version)
	return s.repo.DeleteVersion(ctx, bucket, key, versionID)
}

func (s *StorageService) DeleteObject(ctx context.Context, bucket, key string) error {
	// 1. Soft delete in DB
	if err := s.repo.SoftDelete(ctx, bucket, key); err != nil {
		return err
	}

	// Note: We don't delete from FileStore yet because it's a "soft delete".
	// A background job could clean up Filesystem objects with deleted_at set.

	_ = s.auditSvc.Log(ctx, appcontext.UserIDFromContext(ctx), "storage.object_delete", "storage", bucket+"/"+key, map[string]interface{}{
		"bucket": bucket,
		"key":    key,
	})

	platform.StorageOperations.WithLabelValues("delete", bucket, "success").Inc()

	return nil
}

// CreateBucket creates a new storage bucket.
func (s *StorageService) CreateBucket(ctx context.Context, name string, isPublic bool) (*domain.Bucket, error) {
	if err := validateBucketName(name); err != nil {
		return nil, err
	}
	bucket := &domain.Bucket{
		ID:        uuid.New(),
		Name:      name,
		UserID:    appcontext.UserIDFromContext(ctx),
		IsPublic:  isPublic,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateBucket(ctx, bucket); err != nil {
		return nil, err
	}

	_ = s.auditSvc.Log(ctx, bucket.UserID, "storage.bucket_create", "bucket", bucket.ID.String(), map[string]interface{}{
		"name": name,
	})

	return bucket, nil
}

// GetBucket retrieves bucket metadata.
func (s *StorageService) GetBucket(ctx context.Context, name string) (*domain.Bucket, error) {
	return s.repo.GetBucket(ctx, name)
}

// DeleteBucket deletes a bucket.
func (s *StorageService) DeleteBucket(ctx context.Context, name string) error {
	// Check if bucket is empty? (Logic improvement for later)
	return s.repo.DeleteBucket(ctx, name)
}

// ListBuckets list buckets for the current user.
func (s *StorageService) SetBucketVersioning(ctx context.Context, name string, enabled bool) error {
	return s.repo.SetBucketVersioning(ctx, name, enabled)
}

func (s *StorageService) ListBuckets(ctx context.Context) ([]*domain.Bucket, error) {
	userID := appcontext.UserIDFromContext(ctx)
	return s.repo.ListBuckets(ctx, userID.String())
}

// GetClusterStatus returns the current state of the storage cluster.
func (s *StorageService) GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error) {
	return s.store.GetClusterStatus(ctx)
}

// CreateMultipartUpload initiates a new multipart upload session.
func (s *StorageService) CreateMultipartUpload(ctx context.Context, bucket, key string) (*domain.MultipartUpload, error) {
	// 1. Verify bucket exists
	if _, err := s.repo.GetBucket(ctx, bucket); err != nil {
		return nil, err
	}

	// 2. Create upload metadata
	upload := &domain.MultipartUpload{
		ID:        uuid.New(),
		UserID:    appcontext.UserIDFromContext(ctx),
		Bucket:    bucket,
		Key:       key,
		CreatedAt: time.Now(),
	}

	// 3. Save to repo
	if err := s.repo.SaveMultipartUpload(ctx, upload); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to initiate multipart upload", err)
	}

	_ = s.auditSvc.Log(ctx, upload.UserID, "storage.multipart_init", "storage", upload.ID.String(), map[string]interface{}{
		"bucket": bucket,
		"key":    key,
	})

	return upload, nil
}

// UploadPart uploads a single part of a multipart upload.
func (s *StorageService) UploadPart(ctx context.Context, uploadID uuid.UUID, partNumber int, r io.Reader) (*domain.Part, error) {
	// 1. Get upload metadata
	upload, err := s.repo.GetMultipartUpload(ctx, uploadID)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, errMultipartNotFound, err)
	}

	// 2. Generate unique key for the part
	partKey := fmt.Sprintf(partPathFormat, upload.ID.String(), partNumber)

	// 3. Write data to store
	size, err := s.store.Write(ctx, upload.Bucket, partKey, r)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to write part data", err)
	}

	// 4. Create part metadata
	part := &domain.Part{
		UploadID:   uploadID,
		PartNumber: partNumber,
		SizeBytes:  size,
		ETag:       uuid.New().String(), // In a real system we'd use MD5
	}

	// 5. Save part to repo
	if err := s.repo.SavePart(ctx, part); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to save part metadata", err)
	}

	return part, nil
}

// CompleteMultipartUpload assembles all parts and completes the upload.
func (s *StorageService) CompleteMultipartUpload(ctx context.Context, uploadID uuid.UUID) (*domain.Object, error) {
	// 1. Get upload metadata
	upload, err := s.repo.GetMultipartUpload(ctx, uploadID)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, errMultipartNotFound, err)
	}

	// 2. List all parts
	parts, err := s.repo.ListParts(ctx, uploadID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list parts", err)
	}

	if len(parts) == 0 {
		return nil, errors.New(errors.InvalidInput, "no parts found for upload")
	}

	// 3. Prepare part keys for assembly
	partKeys := make([]string, len(parts))
	for i, p := range parts {
		partKeys[i] = fmt.Sprintf(partPathFormat, upload.ID.String(), p.PartNumber)
	}

	// 4. Assemble in store
	actualSize, err := s.store.Assemble(ctx, upload.Bucket, upload.Key, partKeys)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to assemble object", err)
	}

	// 5. Create final object metadata
	obj := &domain.Object{
		ID:          uuid.New(),
		UserID:      upload.UserID,
		Bucket:      upload.Bucket,
		Key:         upload.Key,
		SizeBytes:   actualSize,
		ContentType: "application/octet-stream",
		CreatedAt:   time.Now(),
		ARN:         fmt.Sprintf("arn:thecloud:storage:local:default:object/%s/%s", upload.Bucket, upload.Key),
	}

	// 6. Save metadata
	if err := s.repo.SaveMeta(ctx, obj); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to save object metadata", err)
	}

	// 7. Cleanup multipart records
	_ = s.repo.DeleteMultipartUpload(ctx, uploadID)

	_ = s.auditSvc.Log(ctx, obj.UserID, "storage.multipart_complete", "storage", obj.ID.String(), map[string]interface{}{
		"bucket": obj.Bucket,
		"key":    obj.Key,
		"size":   obj.SizeBytes,
	})

	return obj, nil
}

// AbortMultipartUpload cancels a multipart upload and cleans up parts.
func (s *StorageService) AbortMultipartUpload(ctx context.Context, uploadID uuid.UUID) error {
	// 1. Get upload
	upload, err := s.repo.GetMultipartUpload(ctx, uploadID)
	if err != nil {
		return errors.Wrap(errors.NotFound, errMultipartNotFound, err)
	}

	// 2. List parts
	parts, err := s.repo.ListParts(ctx, uploadID)
	if err == nil {
		// 3. Delete each part from store
		for _, p := range parts {
			partKey := fmt.Sprintf(partPathFormat, upload.ID.String(), p.PartNumber)
			_ = s.store.Delete(ctx, upload.Bucket, partKey)
		}
	}

	// 4. Delete from repo
	if err := s.repo.DeleteMultipartUpload(ctx, uploadID); err != nil {
		return errors.Wrap(errors.Internal, "failed to delete multipart upload", err)
	}

	_ = s.auditSvc.Log(ctx, appcontext.UserIDFromContext(ctx), "storage.multipart_abort", "storage", uploadID.String(), nil)

	return nil
}

// GeneratePresignedURL generates a temporary signed URL for an object.
func (s *StorageService) GeneratePresignedURL(ctx context.Context, bucket, key, method string, expiry time.Duration) (*domain.PresignedURL, error) {
	// 1. Verify bucket exists
	if _, err := s.repo.GetBucket(ctx, bucket); err != nil {
		return nil, errors.Wrap(errors.NotFound, "bucket not found", err)
	}

	if expiry == 0 {
		expiry = 15 * time.Minute
	}

	expiresAt := time.Now().Add(expiry)

	// Use dependency injected config secret if available
	secret := "storage-secret-key"
	if s.cfg != nil && s.cfg.SecretsEncryptionKey != "" {
		secret = s.cfg.SecretsEncryptionKey
	}

	// Dynamic base URL derived from config or default
	baseURL := "http://localhost:8080"
	if s.cfg != nil && s.cfg.Port != "" {
		// In production this might need a full public URL config
		baseURL = fmt.Sprintf("http://localhost:%s", s.cfg.Port)
	}

	urlStr, err := crypto.SignURL(secret, baseURL, method, bucket, key, expiresAt)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to sign URL", err)
	}

	return &domain.PresignedURL{
		URL:       urlStr,
		Method:    method,
		ExpiresAt: expiresAt,
	}, nil
}

func validateBucketName(name string) error {
	if len(name) == 0 || len(name) > 63 {
		return errors.New(errors.InvalidInput, "bucket name must be 1-63 characters")
	}
	// Only allow alphanumeric, hyphens, and periods.
	// Must start and end with alphanumeric.
	var bucketRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9.\-]*[a-z0-9]$`)
	if !bucketRegex.MatchString(name) {
		return errors.New(errors.InvalidInput, "bucket name must contain only lowercase letters, numbers, hyphens, and periods")
	}
	return nil
}
