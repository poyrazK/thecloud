package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	appcontext "github.com/poyrazk/thecloud/internal/core/context"
	"github.com/poyrazk/thecloud/internal/core/domain"
	theclouderrors "github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestAutoScalingRepo_CreateGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewAutoScalingRepo(mock)
		group := &domain.ScalingGroup{
			ID:             uuid.New(),
			UserID:         uuid.New(),
			IdempotencyKey: "key-1",
			Name:           "asg-1",
			VpcID:          uuid.New(),
			Image:          "ubuntu",
			Ports:          "80:80",
			MinInstances:   1,
			MaxInstances:   5,
			DesiredCount:   2,
			CurrentCount:   0,
			Status:         domain.ScalingGroupStatusActive,
			Version:        1,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		mock.ExpectExec("INSERT INTO scaling_groups").
			WithArgs(group.ID, group.UserID, group.IdempotencyKey, group.Name, group.VpcID, group.LoadBalancerID,
				group.Image, group.Ports, group.MinInstances, group.MaxInstances,
				group.DesiredCount, group.CurrentCount, group.Status, group.Version,
				group.CreatedAt, group.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.CreateGroup(context.Background(), group)
		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewAutoScalingRepo(mock)
		group := &domain.ScalingGroup{ID: uuid.New()}

		mock.ExpectExec("INSERT INTO scaling_groups").
			WillReturnError(errors.New("db error"))

		err = repo.CreateGroup(context.Background(), group)
		assert.Error(t, err)
	})
}

func TestAutoScalingRepo_GetGroupByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewAutoScalingRepo(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		idk := sql.NullString{String: "key-1", Valid: true}
		ports := sql.NullString{String: "80:80", Valid: true}
		var lbID *uuid.UUID = nil

		mock.ExpectQuery("SELECT id, user_id, idempotency_key, name, vpc_id, load_balancer_id, image, ports").
			WithArgs(id, userID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "user_id", "idempotency_key", "name", "vpc_id", "load_balancer_id", "image", "ports",
				"min_instances", "max_instances", "desired_count", "current_count", "status", "version", "created_at", "updated_at",
			}).
				AddRow(id, userID, idk, "asg-1", uuid.New(), lbID, "ubuntu", ports,
					1, 5, 2, 0, string(domain.ScalingGroupStatusActive), 1, now, now))

		g, err := repo.GetGroupByID(ctx, id)
		assert.NoError(t, err)
		assert.NotNil(t, g)
		assert.Equal(t, id, g.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewAutoScalingRepo(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery("SELECT id, user_id, idempotency_key, name, vpc_id, load_balancer_id, image, ports").
			WithArgs(id, userID).
			WillReturnError(pgx.ErrNoRows)

		g, err := repo.GetGroupByID(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, g)
		var target *theclouderrors.Error
		if errors.As(err, &target) {
			assert.Equal(t, theclouderrors.NotFound, target.Type)
		}
	})
}

func TestAutoScalingRepo_ListGroups(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewAutoScalingRepo(mock)
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		idk := sql.NullString{String: "key-1", Valid: true}
		ports := sql.NullString{String: "80:80", Valid: true}
		var lbID *uuid.UUID = nil

		mock.ExpectQuery("SELECT id, user_id, idempotency_key, name, vpc_id, load_balancer_id, image, ports").
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "user_id", "idempotency_key", "name", "vpc_id", "load_balancer_id", "image", "ports",
				"min_instances", "max_instances", "desired_count", "current_count", "status", "version", "created_at", "updated_at",
			}).
				AddRow(uuid.New(), userID, idk, "asg-1", uuid.New(), lbID, "ubuntu", ports,
					1, 5, 2, 0, string(domain.ScalingGroupStatusActive), 1, now, now))

		groups, err := repo.ListGroups(ctx)
		assert.NoError(t, err)
		assert.Len(t, groups, 1)
	})
}

func TestAutoScalingRepo_UpdateGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewAutoScalingRepo(mock)
		group := &domain.ScalingGroup{
			ID:           uuid.New(),
			UserID:       uuid.New(),
			Name:         "asg-updated",
			MinInstances: 2,
			MaxInstances: 10,
			DesiredCount: 5,
			CurrentCount: 0,
			Status:       domain.ScalingGroupStatusActive,
			Version:      1,
			UpdatedAt:    time.Now(),
		}

		mock.ExpectExec("UPDATE scaling_groups").
			WithArgs(group.Name, group.MinInstances, group.MaxInstances, group.DesiredCount, group.Status, group.UpdatedAt, group.ID, group.Version, group.UserID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = repo.UpdateGroup(context.Background(), group)
		assert.NoError(t, err)
		assert.Equal(t, 2, group.Version)
	})

	t.Run("conflict", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewAutoScalingRepo(mock)
		group := &domain.ScalingGroup{
			ID:      uuid.New(),
			UserID:  uuid.New(),
			Version: 1,
		}

		mock.ExpectExec("UPDATE scaling_groups").
			WithArgs(group.Name, group.MinInstances, group.MaxInstances, group.DesiredCount, group.Status, pgxmock.AnyArg(), group.ID, group.Version, group.UserID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

		err = repo.UpdateGroup(context.Background(), group)
		assert.Error(t, err)
		var target *theclouderrors.Error
		if errors.As(err, &target) {
			assert.Equal(t, theclouderrors.Conflict, target.Type)
		}
	})
}

func TestAutoScalingRepo_DeleteGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewAutoScalingRepo(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec("DELETE FROM scaling_groups").
			WithArgs(id, userID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.DeleteGroup(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewAutoScalingRepo(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec("DELETE FROM scaling_groups").
			WithArgs(id, userID).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err = repo.DeleteGroup(ctx, id)
		assert.Error(t, err)
		var target *theclouderrors.Error
		ok := errors.As(err, &target)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, target.Type)
		}
	})
}

func TestAutoScalingRepo_CreatePolicy(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewAutoScalingRepo(mock)
		now := time.Now()
		policy := &domain.ScalingPolicy{
			ID:             uuid.New(),
			ScalingGroupID: uuid.New(),
			Name:           "policy-1",
			MetricType:     "cpu",
			TargetValue:    50,
			ScaleOutStep:   1,
			ScaleInStep:    1,
			CooldownSec:    300,
			LastScaledAt:   &now,
		}

		mock.ExpectExec("INSERT INTO scaling_policies").
			WithArgs(policy.ID, policy.ScalingGroupID, policy.Name, policy.MetricType, policy.TargetValue, policy.ScaleOutStep, policy.ScaleInStep, policy.CooldownSec, policy.LastScaledAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.CreatePolicy(context.Background(), policy)
		assert.NoError(t, err)
	})
}

func TestAutoScalingRepo_GetPoliciesForGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewAutoScalingRepo(mock)
		groupID := uuid.New()
		now := time.Now()

		mock.ExpectQuery("SELECT id, scaling_group_id, name, metric_type, target_value, scale_out_step, scale_in_step, cooldown_sec, last_scaled_at FROM scaling_policies").
			WithArgs(groupID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "scaling_group_id", "name", "metric_type", "target_value", "scale_out_step", "scale_in_step", "cooldown_sec", "last_scaled_at"}).
				AddRow(uuid.New(), groupID, "policy-1", "cpu", 50.0, 1, 1, 300, now))

		policies, err := repo.GetPoliciesForGroup(context.Background(), groupID)
		assert.NoError(t, err)
		assert.Len(t, policies, 1)
	})
}
