package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/stretchr/testify/assert"
)

const (
	testEncKey    = "secret"
	testAlgorithm = "AES256"
)

func TestEncryptionRepository(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	bucketName := "encrypted-bucket"

	t.Run("SaveKey", func(t *testing.T) {
		t.Parallel()
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewEncryptionRepository(mock)
		key := ports.EncryptionKey{
			ID:           uuid.New().String(),
			BucketName:   bucketName,
			EncryptedKey: []byte(testEncKey),
			Algorithm:    testAlgorithm,
		}

		mock.ExpectExec("INSERT INTO encryption_keys").
			WithArgs(key.ID, key.BucketName, key.EncryptedKey, key.Algorithm).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := repo.SaveKey(ctx, key)
		assert.NoError(t, err)
	})

	t.Run("GetKey", func(t *testing.T) {
		t.Parallel()
		mock, _ := pgxmock.NewPool()
		defer mock.Close()
		repo := NewEncryptionRepository(mock)
		id := uuid.New().String()

		mock.ExpectQuery("SELECT id, bucket_name, encrypted_key, algorithm FROM encryption_keys").
			WithArgs(bucketName).
			WillReturnRows(pgxmock.NewRows([]string{"id", "bucket_name", "encrypted_key", "algorithm"}).
				AddRow(id, bucketName, []byte(testEncKey), testAlgorithm))

		key, err := repo.GetKey(ctx, bucketName)
		assert.NoError(t, err)
		assert.NotNil(t, key)
		assert.Equal(t, id, key.ID)
	})
}
