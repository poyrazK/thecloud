package ports

import (
	"context"
	"io"

	"github.com/poyraz/cloud/internal/core/domain"
)

type StorageRepository interface {
	SaveMeta(ctx context.Context, obj *domain.Object) error
	GetMeta(ctx context.Context, bucket, key string) (*domain.Object, error)
	List(ctx context.Context, bucket string) ([]*domain.Object, error)
	SoftDelete(ctx context.Context, bucket, key string) error
}

type FileStore interface {
	Write(ctx context.Context, bucket, key string, r io.Reader) (int64, error)
	Read(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, key string) error
}

type StorageService interface {
	Upload(ctx context.Context, bucket, key string, r io.Reader) (*domain.Object, error)
	Download(ctx context.Context, bucket, key string) (io.ReadCloser, *domain.Object, error)
	ListObjects(ctx context.Context, bucket string) ([]*domain.Object, error)
	DeleteObject(ctx context.Context, bucket, key string) error
}
