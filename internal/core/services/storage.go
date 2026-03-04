// Package services implements core business workflows.
package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
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
	versionEpochBit      = 1 << 62
)

// generateVersionID generates a timestamp-based version ID (reverse chronological).
func generateVersionID() string {
	return fmt.Sprintf("%d", versionEpochBit-time.Now().UnixNano())
}

// versionedStoreKey returns the store key with version ID suffix.
func versionedStoreKey(key, versionID string) string {
	return fmt.Sprintf(versionQueryFormat, key, versionID)
}

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
		versionID = generateVersionID()
	}

	// 2. Sniff content-type (first 512 bytes)
	sniffBuf := make([]byte, 512)
	n, _ := io.ReadFull(r, sniffBuf)
	sniffBuf = sniffBuf[:n]
	contentType := http.DetectContentType(sniffBuf)

	// Combine back for full streaming
	fullReader := io.MultiReader(bytes.NewReader(sniffBuf), r)

	// 3. Prepare initial metadata (PENDING status)
	obj := &domain.Object{
		ID:           uuid.New(),
		UserID:       appcontext.UserIDFromContext(ctx),
		Bucket:       bucketName,
		Key:          key,
		VersionID:    versionID,
		IsLatest:     true,
		SizeBytes:    0,
		ContentType:  contentType,
		UploadStatus: domain.UploadStatusPending,
		CreatedAt:    time.Now(),
	}

	// Generate ARN
	obj.ARN = fmt.Sprintf("arn:thecloud:storage:local:default:object/%s/%s", bucketName, key)
	if bucket.VersioningEnabled {
		obj.ARN += fmt.Sprintf("?versionId=%s", versionID)
	}

	// Save preliminary metadata
	if err := s.repo.SaveMeta(ctx, obj); err != nil {
		return nil, err
	}

	// 4. Setup checksum calculation while streaming
	hash := sha256.New()
	teeReader := io.TeeReader(fullReader, hash)

	// 5. Write file to store
	storeKey := key
	if bucket.VersioningEnabled {
		storeKey = versionedStoreKey(key, versionID)
	}

	dataStream := io.Reader(teeReader)

	// Encryption (Optional)
	if bucket.EncryptionEnabled && s.encryptSvc != nil {
		encryptedReader, err := s.encryptSvc.Encrypt(ctx, bucketName, teeReader)
		if err != nil {
			return nil, errors.Wrap(errors.Internal, "encryption failed", err)
		}
		dataStream = encryptedReader
	}

	size, err := s.store.Write(ctx, bucketName, storeKey, dataStream)
	if err != nil {
		// We leave the record as PENDING. The garbage collector will clean it up.
		return nil, errors.Wrap(errors.Internal, "failed to write to store", err)
	}

	// 6. Update metadata to AVAILABLE and set final size/checksum
	obj.SizeBytes = size
	obj.Checksum = hex.EncodeToString(hash.Sum(nil))
	obj.UploadStatus = domain.UploadStatusAvailable

	if err := s.repo.SaveMeta(ctx, obj); err != nil {
		return nil, err
	}

	_ = s.auditSvc.Log(ctx, obj.UserID, "storage.object_upload", "storage", obj.ARN, map[string]interface{}{
		"bucket": bucketName,
		"key":    key,
		"size":   size,
	})

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
	storeKey := key
	if obj.VersionID != "null" {
		storeKey = versionedStoreKey(key, obj.VersionID)
	}

	reader, err := s.store.Read(ctx, bucket, storeKey)
	if err != nil {
		return nil, nil, err
	}

	// Decryption
	// Check bucket status for efficiency (though key lookup is fast)
	b, err := s.repo.GetBucket(ctx, bucket)
	if err == nil && b.EncryptionEnabled && s.encryptSvc != nil {
		decryptedReader, err := s.encryptSvc.Decrypt(ctx, bucket, reader)
		if err != nil {
			if closeErr := reader.Close(); closeErr != nil {
				return nil, nil, fmt.Errorf("decrypt error: %w; close error: %w", err, closeErr)
			}
			return nil, nil, err
		}
		// Wrap with readCloserWrapper to keep the original closer
		reader = &readCloserWrapper{Reader: decryptedReader, Closer: reader}
	}

	platform.StorageOperations.WithLabelValues("download", bucket, "success").Inc()
	if obj != nil {
		platform.StorageBytesTransferred.WithLabelValues("download").Add(float64(obj.SizeBytes))
	}

	return reader, obj, nil
}

type readCloserWrapper struct {
	io.Reader
	io.Closer
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
		storeKey = versionedStoreKey(key, obj.VersionID)
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
		storeKey = versionedStoreKey(key, versionID)
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
	bucket := &domain.Bucket{
		ID:        uuid.New(),
		Name:      name,
		UserID:    appcontext.UserIDFromContext(ctx),
		IsPublic:  isPublic,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateBucket(ctx, bucket); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to create bucket", err)
	}

	_ = s.auditSvc.Log(ctx, bucket.UserID, "storage.bucket_create", "bucket", name, map[string]interface{}{
		"is_public": isPublic,
	})

	return bucket, nil
}

func (s *StorageService) GetBucket(ctx context.Context, name string) (*domain.Bucket, error) {
	return s.repo.GetBucket(ctx, name)
}

func (s *StorageService) DeleteBucket(ctx context.Context, name string, force bool) error {
	// 1. Check if empty
	objects, err := s.repo.List(ctx, name)
	if err == nil && len(objects) > 0 && !force {
		return errors.New(errors.Conflict, "bucket is not empty")
	}

	// 2. Delete all objects if force
	if force {
		for _, obj := range objects {
			_ = s.DeleteObject(ctx, name, obj.Key)
		}
	}

	// 3. Delete bucket record
	return s.repo.DeleteBucket(ctx, name)
}

func (s *StorageService) ListBuckets(ctx context.Context) ([]*domain.Bucket, error) {
	userID := appcontext.UserIDFromContext(ctx)
	return s.repo.ListBuckets(ctx, userID.String())
}

func (s *StorageService) SetBucketVersioning(ctx context.Context, name string, enabled bool) error {
	return s.repo.SetBucketVersioning(ctx, name, enabled)
}

func (s *StorageService) GetClusterStatus(ctx context.Context) (*domain.StorageCluster, error) {
	return s.store.GetClusterStatus(ctx)
}

// CleanupDeleted removes soft-deleted objects from the storage backend.
func (s *StorageService) CleanupDeleted(ctx context.Context, limit int) (int, error) {
	deleted, err := s.repo.ListDeleted(ctx, limit)
	if err != nil {
		return 0, errors.Wrap(errors.Internal, "failed to list deleted objects", err)
	}

	deletedCount := 0
	for _, obj := range deleted {
		storeKey := obj.Key
		if obj.VersionID != "null" {
			storeKey = versionedStoreKey(obj.Key, obj.VersionID)
		}

		// 1. Check if other versions still exist (if versioning was enabled)
		// 2. Delete the physical file
		// We ignore error from store.Delete if it's already missing
		_ = s.store.Delete(ctx, obj.Bucket, storeKey)

		// 3. Permanent delete from DB
		if err := s.repo.HardDelete(ctx, obj.Bucket, obj.Key, obj.VersionID); err != nil {
			return deletedCount, errors.Wrap(errors.Internal, "failed to hard delete object", err)
		}
		deletedCount++
	}

	return deletedCount, nil
}

// CleanupPendingUploads removes orphaned files from failed uploads.
func (s *StorageService) CleanupPendingUploads(ctx context.Context, olderThan time.Duration, limit int) (int, error) {
	threshold := time.Now().Add(-olderThan)
	pending, err := s.repo.ListPending(ctx, threshold, limit)
	if err != nil {
		return 0, errors.Wrap(errors.Internal, "failed to list pending uploads", err)
	}

	cleanedCount := 0
	for _, obj := range pending {
		storeKey := obj.Key
		if obj.VersionID != "null" {
			storeKey = versionedStoreKey(obj.Key, obj.VersionID)
		}

		// 1. Delete the physical file (if it exists)
		_ = s.store.Delete(ctx, obj.Bucket, storeKey)

		// 2. Delete the metadata record
		if err := s.repo.HardDelete(ctx, obj.Bucket, obj.Key, obj.VersionID); err != nil {
			return cleanedCount, errors.Wrap(errors.Internal, "failed to delete pending metadata", err)
		}
		cleanedCount++
	}

	return cleanedCount, nil
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
	// 1. Get upload
	upload, err := s.repo.GetMultipartUpload(ctx, uploadID)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, errMultipartNotFound, err)
	}

	// 2. Calculate checksum while streaming to store
	hash := sha256.New()
	teeReader := io.TeeReader(r, hash)

	// 3. Write to store (use temporary location)
	partKey := fmt.Sprintf(partPathFormat, upload.ID.String(), partNumber)
	size, err := s.store.Write(ctx, upload.Bucket, partKey, teeReader)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to write part", err)
	}

	// 4. Save part metadata
	part := &domain.Part{
		UploadID:   uploadID,
		PartNumber: partNumber,
		SizeBytes:  size,
		ETag:       hex.EncodeToString(hash.Sum(nil)),
	}

	if err := s.repo.SavePart(ctx, part); err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to save part metadata", err)
	}

	return part, nil
}

// CompleteMultipartUpload assembles all parts into a single object.
func (s *StorageService) CompleteMultipartUpload(ctx context.Context, uploadID uuid.UUID) (*domain.Object, error) {
	// 1. Get upload and parts
	upload, err := s.repo.GetMultipartUpload(ctx, uploadID)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, errMultipartNotFound, err)
	}

	parts, err := s.repo.ListParts(ctx, uploadID)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to list parts", err)
	}

	if len(parts) == 0 {
		return nil, errors.New(errors.InvalidInput, "no parts uploaded")
	}

	// 2. Assemble in store
	partKeys := make([]string, len(parts))
	for i, p := range parts {
		partKeys[i] = fmt.Sprintf(partPathFormat, upload.ID.String(), p.PartNumber)
	}

	bucket, _ := s.repo.GetBucket(ctx, upload.Bucket)
	versionID := "null"
	if bucket.VersioningEnabled {
		versionID = generateVersionID()
	}

	storeKey := upload.Key
	if bucket.VersioningEnabled {
		storeKey = versionedStoreKey(upload.Key, versionID)
	}

	size, err := s.store.Assemble(ctx, upload.Bucket, storeKey, partKeys)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to assemble parts", err)
	}

	// 3. Create final object metadata
	obj := &domain.Object{
		ID:           uuid.New(),
		UserID:       upload.UserID,
		Bucket:       upload.Bucket,
		Key:          upload.Key,
		VersionID:    versionID,
		IsLatest:     true,
		SizeBytes:    size,
		UploadStatus: domain.UploadStatusAvailable,
		CreatedAt:    time.Now(),
	}

	// Sniff content type and calculate checksum from the assembled file
	reader, err := s.store.Read(ctx, upload.Bucket, storeKey)
	if err == nil {
		defer reader.Close()
		hash := sha256.New()
		sniffBuf := make([]byte, 512)
		n, _ := io.ReadFull(reader, sniffBuf)
		obj.ContentType = http.DetectContentType(sniffBuf[:n])

		// Continue reading for checksum
		fullReader := io.MultiReader(bytes.NewReader(sniffBuf[:n]), reader)
		_, _ = io.Copy(hash, fullReader)
		obj.Checksum = hex.EncodeToString(hash.Sum(nil))
	} else {
		obj.ContentType = "application/octet-stream"
	}

	// Generate ARN
	obj.ARN = fmt.Sprintf("arn:thecloud:storage:local:default:object/%s/%s", upload.Bucket, upload.Key)
	if bucket.VersioningEnabled {
		obj.ARN += fmt.Sprintf("?versionId=%s", versionID)
	}

	// 4. Save metadata
	if err := s.repo.SaveMeta(ctx, obj); err != nil {
		return nil, err
	}

	// 5. Cleanup multipart records
	_ = s.repo.DeleteMultipartUpload(ctx, uploadID)

	_ = s.auditSvc.Log(ctx, obj.UserID, "storage.multipart_complete", "storage", obj.ARN, map[string]interface{}{
		"bucket": upload.Bucket,
		"key":    upload.Key,
		"size":   size,
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
		return errors.Wrap(errors.Internal, "failed to delete multipart upload metadata", err)
	}

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
	secret := s.cfg.SecretsEncryptionKey
	if secret == "" {
		return nil, errors.New(errors.Internal, "storage secret not configured")
	}

	// Dynamic base URL derived from config or default
	baseURL := "http://localhost:8080"
	if s.cfg != nil && s.cfg.Port != "" {
		baseURL = fmt.Sprintf("http://localhost:%s", s.cfg.Port)
	}

	signedURL, err := crypto.SignURL(secret, baseURL, method, bucket, key, expiresAt)
	if err != nil {
		return nil, errors.Wrap(errors.Internal, "failed to sign URL", err)
	}

	return &domain.PresignedURL{
		URL:       signedURL,
		Method:    method,
		ExpiresAt: expiresAt,
	}, nil
}
