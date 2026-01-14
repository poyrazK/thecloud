package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	theclouderrors "github.com/poyrazk/thecloud/internal/errors"
	"github.com/poyrazk/thecloud/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

const (
	testVpcName = "test-vpc"
	selectVpc   = "SELECT id, user_id, name, COALESCE\\(cidr_block::text, ''\\), network_id, vxlan_id, status, arn, created_at FROM vpcs"
)

func TestVpcRepositoryCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVpcRepository(mock)
		vpc := &domain.VPC{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Name:      testVpcName,
			CIDRBlock: testutil.TestCIDR,
			NetworkID: "net-1",
			VXLANID:   100,
			Status:    "available",
			ARN:       "arn",
			CreatedAt: time.Now(),
		}

		mock.ExpectExec("INSERT INTO vpcs").
			WithArgs(vpc.ID, vpc.UserID, vpc.Name, vpc.CIDRBlock, vpc.NetworkID, vpc.VXLANID, vpc.Status, vpc.ARN, vpc.CreatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.Create(context.Background(), vpc)
		assert.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVpcRepository(mock)
		vpc := &domain.VPC{
			ID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO vpcs").
			WillReturnError(errors.New(testDBError))

		err = repo.Create(context.Background(), vpc)
		assert.Error(t, err)
	})
}

func TestVpcRepositoryGetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVpcRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		mock.ExpectQuery(selectVpc).
			WithArgs(id, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "cidr_block", "network_id", "vxlan_id", "status", "arn", "created_at"}).
				AddRow(id, userID, testVpcName, testutil.TestCIDR, "net-1", 100, "available", "arn", now))

		vpc, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, vpc)
		assert.Equal(t, id, vpc.ID)
	})

	t.Run(testNotFound, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVpcRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery("SELECT id, user_id, name, COALESCE\\(cidr_block::text, ''\\), network_id, vxlan_id, status, arn, created_at FROM vpcs").
			WithArgs(id, userID).
			WillReturnError(pgx.ErrNoRows)

		vpc, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, vpc)
		// Check if the error returned is of type errors.NotFound
		theCloudErr, ok := err.(*theclouderrors.Error)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
		// If it's wrapped or directly the error we expect (but it seems the repo wraps it)
		// The implementation: return nil, errors.New(errors.NotFound, fmt.Sprintf("vpc %s not found", id))
		// So it should be a custom error.
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVpcRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery(selectVpc).
			WithArgs(id, userID).
			WillReturnError(errors.New(testDBError))

		vpc, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, vpc)
	})
}

func TestVpcRepositoryGetByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVpcRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()
		name := testVpcName

		mock.ExpectQuery(selectVpc).
			WithArgs(name, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "cidr_block", "network_id", "vxlan_id", "status", "arn", "created_at"}).
				AddRow(id, userID, name, testutil.TestCIDR, "net-1", 100, "available", "arn", now))

		vpc, err := repo.GetByName(ctx, name)
		assert.NoError(t, err)
		assert.NotNil(t, vpc)
		assert.Equal(t, id, vpc.ID)
	})

	t.Run(testNotFound, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVpcRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		name := testVpcName

		mock.ExpectQuery(selectVpc).
			WithArgs(name, userID).
			WillReturnError(pgx.ErrNoRows)

		vpc, err := repo.GetByName(ctx, name)
		assert.Error(t, err)
		assert.Nil(t, vpc)
	})
}

func TestVpcRepositoryList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVpcRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		mock.ExpectQuery(selectVpc).
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "cidr_block", "network_id", "vxlan_id", "status", "arn", "created_at"}).
				AddRow(uuid.New(), userID, testVpcName, testutil.TestCIDR, "net-1", 100, "available", "arn", now))

		vpcs, err := repo.List(ctx)
		assert.NoError(t, err)
		assert.Len(t, vpcs, 1)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVpcRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery(selectVpc).
			WithArgs(userID).
			WillReturnError(errors.New(testDBError))

		vpcs, err := repo.List(ctx)
		assert.Error(t, err)
		assert.Nil(t, vpcs)
	})

	t.Run("scan error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVpcRepository(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		// Return a row with incompatible types to force scan error
		mock.ExpectQuery(selectVpc).
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "name", "cidr_block", "network_id", "vxlan_id", "status", "arn", "created_at"}).
				AddRow("invalid-uuid", userID, testVpcName, testutil.TestCIDR, "net-1", 100, "available", "arn", now))

		vpcs, err := repo.List(ctx)
		assert.Error(t, err)
		assert.Nil(t, vpcs)
	})
}

func TestVpcRepositoryDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVpcRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec("DELETE FROM vpcs").
			WithArgs(id, userID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.Delete(ctx, id)
		assert.NoError(t, err)
	})

	t.Run(testNotFound, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewVpcRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec("DELETE FROM vpcs").
			WithArgs(id, userID).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err = repo.Delete(ctx, id)
		assert.Error(t, err)
	})
}
