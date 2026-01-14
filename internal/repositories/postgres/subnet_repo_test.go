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
	testSubnetName    = "subnet-1"
	testAZ            = "us-east-1a"
	selectSubnet      = "SELECT id, user_id, vpc_id, name, cidr_block::text, availability_zone, COALESCE\\(gateway_ip::text, ''\\), arn, status, created_at FROM subnets"
	deleteSubnetQuery = "DELETE FROM subnets"
)

func TestSubnetRepositoryCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSubnetRepository(mock)
		s := &domain.Subnet{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			VPCID:            uuid.New(),
			Name:             testSubnetName,
			CIDRBlock:        testutil.TestSubnetCIDR,
			AvailabilityZone: testAZ,
			GatewayIP:        testutil.TestGatewayIP,
			ARN:              "arn",
			Status:           "available",
			CreatedAt:        time.Now(),
		}

		mock.ExpectExec("INSERT INTO subnets").
			WithArgs(s.ID, s.UserID, s.VPCID, s.Name, s.CIDRBlock, s.AvailabilityZone, s.GatewayIP, s.ARN, s.Status, s.CreatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.Create(context.Background(), s)
		assert.NoError(t, err)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSubnetRepository(mock)
		s := &domain.Subnet{
			ID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO subnets").
			WillReturnError(errors.New(testDBError))

		err = repo.Create(context.Background(), s)
		assert.Error(t, err)
	})
}

func TestSubnetRepositoryGetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSubnetRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		mock.ExpectQuery(selectSubnet).
			WithArgs(id, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "vpc_id", "name", "cidr_block", "availability_zone", "gateway_ip", "arn", "status", "created_at"}).
				AddRow(id, userID, uuid.New(), testSubnetName, testutil.TestSubnetCIDR, testAZ, testutil.TestGatewayIP, "arn", "available", now))

		s, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, s)
		assert.Equal(t, id, s.ID)

	})

	t.Run(testNotFound, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSubnetRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery("SELECT id, user_id, vpc_id, name, cidr_block::text, availability_zone, COALESCE\\(gateway_ip::text, ''\\), arn, status, created_at FROM subnets").
			WithArgs(id, userID).
			WillReturnError(pgx.ErrNoRows)

		s, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, s)
		theCloudErr, ok := err.(*theclouderrors.Error)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSubnetRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery(selectSubnet).
			WithArgs(id, userID).
			WillReturnError(errors.New(testDBError))

		s, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, s)
	})
}

func TestSubnetRepositoryGetByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSubnetRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		vpcID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()
		name := testSubnetName

		mock.ExpectQuery(selectSubnet).
			WithArgs(vpcID, name, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "vpc_id", "name", "cidr_block", "availability_zone", "gateway_ip", "arn", "status", "created_at"}).
				AddRow(id, userID, vpcID, name, testutil.TestSubnetCIDR, testAZ, testutil.TestGatewayIP, "arn", "available", now))

		s, err := repo.GetByName(ctx, vpcID, name)
		assert.NoError(t, err)
		assert.NotNil(t, s)
		assert.Equal(t, id, s.ID)

	})

	t.Run(testNotFound, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSubnetRepository(mock)
		userID := uuid.New()
		vpcID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		name := testSubnetName

		mock.ExpectQuery(selectSubnet).
			WithArgs(vpcID, name, userID).
			WillReturnError(pgx.ErrNoRows)

		s, err := repo.GetByName(ctx, vpcID, name)
		assert.Error(t, err)
		assert.Nil(t, s)
		theCloudErr, ok := err.(*theclouderrors.Error)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
	})
}

func TestSubnetRepositoryListByVPC(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSubnetRepository(mock)
		vpcID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		mock.ExpectQuery(selectSubnet).
			WithArgs(vpcID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "vpc_id", "name", "cidr_block", "availability_zone", "gateway_ip", "arn", "status", "created_at"}).
				AddRow(uuid.New(), userID, vpcID, testSubnetName, testutil.TestSubnetCIDR, testAZ, testutil.TestGatewayIP, "arn", "available", now))

		subnets, err := repo.ListByVPC(ctx, vpcID)
		assert.NoError(t, err)
		assert.Len(t, subnets, 1)

	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSubnetRepository(mock)
		vpcID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery(selectSubnet).
			WithArgs(vpcID, userID).
			WillReturnError(errors.New(testDBError))

		subnets, err := repo.ListByVPC(ctx, vpcID)
		assert.Error(t, err)
		assert.Nil(t, subnets)
	})

	t.Run("scan error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSubnetRepository(mock)
		vpcID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		mock.ExpectQuery(selectSubnet).
			WithArgs(vpcID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "vpc_id", "name", "cidr_block", "availability_zone", "gateway_ip", "arn", "status", "created_at"}).
				AddRow("invalid-uuid", userID, vpcID, testSubnetName, testutil.TestSubnetCIDR, testAZ, testutil.TestGatewayIP, "arn", "available", now))

		subnets, err := repo.ListByVPC(ctx, vpcID)
		assert.Error(t, err)
		assert.Nil(t, subnets)
	})
}

func TestSubnetRepositoryDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSubnetRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec(deleteSubnetQuery).
			WithArgs(id, userID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.Delete(ctx, id)
		assert.NoError(t, err)
	})

	t.Run(testNotFound, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSubnetRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec(deleteSubnetQuery).
			WithArgs(id, userID).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err = repo.Delete(ctx, id)
		assert.Error(t, err)
		theCloudErr, ok := err.(*theclouderrors.Error)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSubnetRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec(deleteSubnetQuery).
			WithArgs(id, userID).
			WillReturnError(errors.New(testDBError))

		err = repo.Delete(ctx, id)
		assert.Error(t, err)
	})
}
