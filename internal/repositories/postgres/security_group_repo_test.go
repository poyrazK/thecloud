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
	testSgName         = "test-sg"
	selectSg           = "SELECT id, user_id, vpc_id, name, description, arn, created_at FROM security_groups"
	selectRule         = "SELECT id, group_id, direction, protocol, port_min, port_max, cidr, priority, created_at FROM security_rules"
	selectInstanceUser = "SELECT user_id FROM instances WHERE id = \\$1"
	selectSgUser       = "SELECT user_id FROM security_groups WHERE id = \\$1"
)

func TestSecurityGroupRepositoryCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		sg := &domain.SecurityGroup{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			VPCID:       uuid.New(),
			Name:        testSgName,
			Description: "desc",
			ARN:         "arn",
			CreatedAt:   time.Now(),
		}

		mock.ExpectExec("INSERT INTO security_groups").
			WithArgs(sg.ID, sg.UserID, sg.VPCID, sg.Name, sg.Description, sg.ARN, sg.CreatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.Create(context.Background(), sg)
		assert.NoError(t, err)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		sg := &domain.SecurityGroup{
			ID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO security_groups").
			WillReturnError(errors.New(testDBError))

		err = repo.Create(context.Background(), sg)
		assert.Error(t, err)
	})
}

func TestSecurityGroupRepositoryGetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		mock.ExpectQuery(selectSg).
			WithArgs(id, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "vpc_id", "name", "description", "arn", "created_at"}).
				AddRow(id, userID, uuid.New(), testSgName, "desc", "arn", now))

		mock.ExpectQuery(selectRule).
			WithArgs(id).
			WillReturnRows(pgxmock.NewRows([]string{"id", "group_id", "direction", "protocol", "port_min", "port_max", "cidr", "priority", "created_at"}).
				AddRow(uuid.New(), id, string(domain.RuleIngress), "tcp", 80, 80, testutil.TestAnyCIDR, 100, now))

		sg, err := repo.GetByID(ctx, id)
		assert.NoError(t, err)
		if sg != nil {
			assert.Len(t, sg.Rules, 1)
		}
	})

	t.Run(testNotFound, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery(selectSg).
			WithArgs(id, userID).
			WillReturnError(pgx.ErrNoRows)

		sg, err := repo.GetByID(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, sg)
		theCloudErr, ok := err.(*theclouderrors.Error)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
	})
}

func TestSecurityGroupRepositoryGetByName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		vpcID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()
		name := testSgName

		mock.ExpectQuery(selectSg).
			WithArgs(vpcID, name, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "vpc_id", "name", "description", "arn", "created_at"}).
				AddRow(id, userID, vpcID, name, "desc", "arn", now))

		mock.ExpectQuery(selectRule).
			WithArgs(id).
			WillReturnRows(pgxmock.NewRows([]string{"id", "group_id", "direction", "protocol", "port_min", "port_max", "cidr", "priority", "created_at"}).
				AddRow(uuid.New(), id, string(domain.RuleIngress), "tcp", 80, 80, testutil.TestAnyCIDR, 100, now))

		sg, err := repo.GetByName(ctx, vpcID, name)
		assert.NoError(t, err)
		assert.NotNil(t, sg)
		assert.Equal(t, id, sg.ID)
	})

	t.Run(testNotFound, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		userID := uuid.New()
		vpcID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		name := testSgName

		mock.ExpectQuery(selectSg).
			WithArgs(vpcID, name, userID).
			WillReturnError(pgx.ErrNoRows)

		sg, err := repo.GetByName(ctx, vpcID, name)
		assert.Error(t, err)
		assert.Nil(t, sg)
		theCloudErr, ok := err.(*theclouderrors.Error)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
	})
}

func TestSecurityGroupRepositoryListByVPC(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		vpcID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		mock.ExpectQuery(selectSg).
			WithArgs(vpcID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "vpc_id", "name", "description", "arn", "created_at"}).
				AddRow(uuid.New(), userID, vpcID, testSgName, "desc", "arn", now))

		groups, err := repo.ListByVPC(ctx, vpcID)
		assert.NoError(t, err)
		assert.Len(t, groups, 1)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		vpcID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery(selectSg).
			WithArgs(vpcID, userID).
			WillReturnError(errors.New(testDBError))

		groups, err := repo.ListByVPC(ctx, vpcID)
		assert.Error(t, err)
		assert.Nil(t, groups)
	})
}

func TestSecurityGroupRepositoryAddRule(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		rule := &domain.SecurityRule{
			ID:        uuid.New(),
			GroupID:   uuid.New(),
			Direction: domain.RuleIngress,
			Protocol:  "tcp",
			PortMin:   80,
			PortMax:   80,
			CIDR:      testutil.TestAnyCIDR,
			Priority:  100,
			CreatedAt: time.Now(),
		}

		mock.ExpectExec("INSERT INTO security_rules").
			WithArgs(rule.ID, rule.GroupID, rule.Direction, rule.Protocol, rule.PortMin, rule.PortMax, rule.CIDR, rule.Priority, rule.CreatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.AddRule(context.Background(), rule)
		assert.NoError(t, err)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		rule := &domain.SecurityRule{
			ID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO security_rules").
			WillReturnError(errors.New(testDBError))

		err = repo.AddRule(context.Background(), rule)
		assert.Error(t, err)
	})
}

func TestSecurityGroupRepositoryDeleteRule(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		ruleID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec("DELETE FROM security_rules").
			WithArgs(ruleID, userID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.DeleteRule(ctx, ruleID)
		assert.NoError(t, err)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		ruleID := uuid.New()

		mock.ExpectExec("DELETE FROM security_rules").
			WithArgs(ruleID).
			WillReturnError(errors.New(testDBError))

		err = repo.DeleteRule(context.Background(), ruleID)
		assert.Error(t, err)
	})
}

func TestSecurityGroupRepositoryGetRuleByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		ruleID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)
		now := time.Now()

		rows := pgxmock.NewRows([]string{"id", "group_id", "direction", "protocol", "port_min", "port_max", "cidr", "priority", "created_at"}).
			AddRow(ruleID, uuid.New(), "ingress", "tcp", 80, 80, "0.0.0.0/0", 100, now)

		mock.ExpectQuery("SELECT .* FROM security_rules").
			WithArgs(ruleID, userID).
			WillReturnRows(rows)

		rule, err := repo.GetRuleByID(ctx, ruleID)
		assert.NoError(t, err)
		assert.NotNil(t, rule)
		assert.Equal(t, ruleID, rule.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		ruleID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectQuery("SELECT .* FROM security_rules").
			WithArgs(ruleID, userID).
			WillReturnError(pgx.ErrNoRows)

		rule, err := repo.GetRuleByID(ctx, ruleID)
		assert.Error(t, err)
		assert.Nil(t, rule)
		assert.IsType(t, theclouderrors.Error{}, err)
		assert.Equal(t, theclouderrors.NotFound, err.(theclouderrors.Error).Type)
	})
}

func TestSecurityGroupRepositoryDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec("DELETE FROM security_groups").
			WithArgs(id, userID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.Delete(ctx, id)
		assert.NoError(t, err)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		id := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectExec("DELETE FROM security_groups").
			WithArgs(id, userID).
			WillReturnError(errors.New(testDBError))

		err = repo.Delete(ctx, id)
		assert.Error(t, err)
	})
}

func TestSecurityGroupRepositoryAddInstanceToGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		instanceID := uuid.New()
		groupID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectBegin()
		mock.ExpectQuery(selectInstanceUser).
			WithArgs(instanceID).
			WillReturnRows(pgxmock.NewRows([]string{"user_id"}).AddRow(userID))
		mock.ExpectQuery(selectSgUser).
			WithArgs(groupID).
			WillReturnRows(pgxmock.NewRows([]string{"user_id"}).AddRow(userID))
		mock.ExpectExec("INSERT INTO instance_security_groups").
			WithArgs(instanceID, groupID).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
		mock.ExpectCommit()

		err = repo.AddInstanceToGroup(ctx, instanceID, groupID)
		assert.NoError(t, err)
	})

	t.Run("instance not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		instanceID := uuid.New()
		groupID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectBegin()
		mock.ExpectQuery(selectInstanceUser).
			WithArgs(instanceID).
			WillReturnError(pgx.ErrNoRows)
		mock.ExpectRollback()

		err = repo.AddInstanceToGroup(ctx, instanceID, groupID)
		assert.Error(t, err)
		theCloudErr, ok := err.(*theclouderrors.Error)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
	})

	t.Run("security group not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		instanceID := uuid.New()
		groupID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectBegin()
		mock.ExpectQuery(selectInstanceUser).
			WithArgs(instanceID).
			WillReturnRows(pgxmock.NewRows([]string{"user_id"}).AddRow(userID))
		mock.ExpectQuery(selectSgUser).
			WithArgs(groupID).
			WillReturnError(pgx.ErrNoRows)
		mock.ExpectRollback()

		err = repo.AddInstanceToGroup(ctx, instanceID, groupID)
		assert.Error(t, err)
		theCloudErr, ok := err.(*theclouderrors.Error)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
	})
}

func TestSecurityGroupRepositoryRemoveInstanceFromGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		instanceID := uuid.New()
		groupID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectBegin()
		mock.ExpectQuery(selectInstanceUser).
			WithArgs(instanceID).
			WillReturnRows(pgxmock.NewRows([]string{"user_id"}).AddRow(userID))
		mock.ExpectQuery(selectSgUser).
			WithArgs(groupID).
			WillReturnRows(pgxmock.NewRows([]string{"user_id"}).AddRow(userID))
		mock.ExpectExec("DELETE FROM instance_security_groups").
			WithArgs(instanceID, groupID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		mock.ExpectCommit()

		err = repo.RemoveInstanceFromGroup(ctx, instanceID, groupID)
		assert.NoError(t, err)
	})

	t.Run("instance not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		instanceID := uuid.New()
		groupID := uuid.New()
		userID := uuid.New()
		ctx := appcontext.WithUserID(context.Background(), userID)

		mock.ExpectBegin()
		mock.ExpectQuery(selectInstanceUser).
			WithArgs(instanceID).
			WillReturnError(pgx.ErrNoRows)
		mock.ExpectRollback()

		err = repo.RemoveInstanceFromGroup(ctx, instanceID, groupID)
		assert.Error(t, err)
		theCloudErr, ok := err.(*theclouderrors.Error)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
	})
}

func TestSecurityGroupRepositoryListInstanceGroups(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		instanceID := uuid.New()
		now := time.Now()

		mock.ExpectQuery("SELECT sg.id, sg.user_id, sg.vpc_id, sg.name, sg.description, sg.arn, sg.created_at").
			WithArgs(instanceID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "vpc_id", "name", "description", "arn", "created_at"}).
				AddRow(uuid.New(), uuid.New(), uuid.New(), testSgName, "desc", "arn", now))

		groups, err := repo.ListInstanceGroups(context.Background(), instanceID)
		assert.NoError(t, err)
		assert.Len(t, groups, 1)
	})

	t.Run(testDBError, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewSecurityGroupRepository(mock)
		instanceID := uuid.New()

		mock.ExpectQuery("SELECT sg.id, sg.user_id, sg.vpc_id, sg.name, sg.description, sg.arn, sg.created_at").
			WithArgs(instanceID).
			WillReturnError(errors.New(testDBError))

		groups, err := repo.ListInstanceGroups(context.Background(), instanceID)
		assert.Error(t, err)
		assert.Nil(t, groups)
	})
}
