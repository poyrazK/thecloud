package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	theclouderrors "github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestInstanceTypeRepositoryList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewInstanceTypeRepository(mock)

		mock.ExpectQuery(`SELECT id, name, vcpus, memory_mb, disk_gb, network_mbps, price_per_hour, category FROM instance_types`).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "vcpus", "memory_mb", "disk_gb", "network_mbps", "price_per_hour", "category"}).
				AddRow("basic-1", "Basic 1", 1, 1024, 10, 1000, 0.05, "general"))

		list, err := repo.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "basic-1", list[0].ID)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewInstanceTypeRepository(mock)

		mock.ExpectQuery(`SELECT id, name`).
			WillReturnError(errors.New("db error"))

		list, err := repo.List(context.Background())
		assert.Error(t, err)
		assert.Nil(t, list)
	})
}

func TestInstanceTypeRepositoryGetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewInstanceTypeRepository(mock)
		id := "basic-1"

		mock.ExpectQuery(`SELECT id, name`).
			WithArgs(id).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "vcpus", "memory_mb", "disk_gb", "network_mbps", "price_per_hour", "category"}).
				AddRow(id, "Basic 1", 1, 1024, 10, 1000, 0.05, "general"))

		it, err := repo.GetByID(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, id, it.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewInstanceTypeRepository(mock)
		id := "non-existent"

		mock.ExpectQuery(`SELECT id, name`).
			WithArgs(id).
			WillReturnError(pgx.ErrNoRows)

		it, err := repo.GetByID(context.Background(), id)
		assert.Error(t, err)
		assert.Nil(t, it)
		var target *theclouderrors.Error
		if errors.As(err, &target) {
			assert.Equal(t, theclouderrors.NotFound, target.Type)
		}
	})
}

func TestInstanceTypeRepositoryCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewInstanceTypeRepository(mock)
		it := &domain.InstanceType{
			ID:          "basic-new",
			Name:        "Basic New",
			VCPUs:       2,
			MemoryMB:    2048,
			DiskGB:      20,
			NetworkMbps: 2000,
			PricePerHr:  0.10,
			Category:    "compute",
		}

		mock.ExpectQuery(`INSERT INTO instance_types`).
			WithArgs(it.ID, it.Name, it.VCPUs, it.MemoryMB, it.DiskGB, it.NetworkMbps, it.PricePerHr, it.Category).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "vcpus", "memory_mb", "disk_gb", "network_mbps", "price_per_hour", "category"}).
				AddRow(it.ID, it.Name, it.VCPUs, it.MemoryMB, it.DiskGB, it.NetworkMbps, it.PricePerHr, it.Category))

		created, err := repo.Create(context.Background(), it)
		assert.NoError(t, err)
		assert.Equal(t, it.ID, created.ID)
	})
}

func TestInstanceTypeRepositoryUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewInstanceTypeRepository(mock)
		it := &domain.InstanceType{
			ID:          "basic-update",
			Name:        "Basic Update",
			VCPUs:       4,
			MemoryMB:    4096,
			DiskGB:      40,
			NetworkMbps: 4000,
			PricePerHr:  0.20,
			Category:    "memory",
		}

		mock.ExpectQuery(`UPDATE instance_types`).
			WithArgs(it.ID, it.Name, it.VCPUs, it.MemoryMB, it.DiskGB, it.NetworkMbps, it.PricePerHr, it.Category).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name", "vcpus", "memory_mb", "disk_gb", "network_mbps", "price_per_hour", "category"}).
				AddRow(it.ID, it.Name, it.VCPUs, it.MemoryMB, it.DiskGB, it.NetworkMbps, it.PricePerHr, it.Category))

		updated, err := repo.Update(context.Background(), it)
		assert.NoError(t, err)
		assert.Equal(t, it.VCPUs, updated.VCPUs)
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewInstanceTypeRepository(mock)
		it := &domain.InstanceType{ID: "non-existent"}

		mock.ExpectQuery(`UPDATE instance_types`).
			WithArgs(it.ID, it.Name, it.VCPUs, it.MemoryMB, it.DiskGB, it.NetworkMbps, it.PricePerHr, it.Category).
			WillReturnError(pgx.ErrNoRows)

		updated, err := repo.Update(context.Background(), it)
		assert.Error(t, err)
		assert.Nil(t, updated)
		var target *theclouderrors.Error
		if errors.As(err, &target) {
			assert.Equal(t, theclouderrors.NotFound, target.Type)
		}
	})
}

func TestInstanceTypeRepositoryDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewInstanceTypeRepository(mock)
		id := "basic-del"

		mock.ExpectExec(`DELETE FROM instance_types`).
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.Delete(context.Background(), id)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewInstanceTypeRepository(mock)
		id := "non-existent"

		mock.ExpectExec(`DELETE FROM instance_types`).
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err = repo.Delete(context.Background(), id)
		assert.Error(t, err)
		var target *theclouderrors.Error
		if errors.As(err, &target) {
			assert.Equal(t, theclouderrors.NotFound, target.Type)
		}
	})
}
