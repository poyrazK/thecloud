package workers

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDB struct {
	mock.Mock
}

func (m *mockDB) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	r0, _ := args.Get(0).(pgconn.CommandTag)
	return r0, args.Error(1)
}
func (m *mockDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	a := m.Called(ctx, sql, args)
	r0, _ := a.Get(0).(pgx.Rows)
	return r0, a.Error(1)
}
func (m *mockDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	r0, _ := m.Called(ctx, sql, args).Get(0).(pgx.Row)
	return r0
}
func (m *mockDB) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	r0, _ := args.Get(0).(pgx.Tx)
	return r0, args.Error(1)
}
func (m *mockDB) Close() {
	m.Called()
}
func (m *mockDB) Ping(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func TestReplicaMonitor(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("Healthy Replica", func(t *testing.T) {
		primary := new(mockDB)
		replica := new(mockDB)
		dual := postgres.NewDualDB(primary, replica)
		monitor := NewReplicaMonitor(dual, replica, logger)

		replica.On("Ping", mock.Anything).Return(nil).Once()

		monitor.check(ctx)

		replica.AssertExpectations(t)
		assert.True(t, monitor.IsHealthy())
	})

	t.Run("Unhealthy Replica", func(t *testing.T) {
		primary := new(mockDB)
		replica := new(mockDB)
		dual := postgres.NewDualDB(primary, replica)
		monitor := NewReplicaMonitor(dual, replica, logger)

		replica.On("Ping", mock.Anything).Return(errors.New("ping failed")).Once()

		monitor.check(ctx)

		replica.AssertExpectations(t)
		assert.False(t, monitor.IsHealthy())
	})

	t.Run("Recovery", func(t *testing.T) {
		primary := new(mockDB)
		replica := new(mockDB)
		dual := postgres.NewDualDB(primary, replica)
		monitor := NewReplicaMonitor(dual, replica, logger)

		// First fail
		replica.On("Ping", mock.Anything).Return(errors.New("ping failed")).Once()
		monitor.check(ctx)
		assert.False(t, monitor.IsHealthy())

		// Then recover
		replica.On("Ping", mock.Anything).Return(nil).Once()
		monitor.check(ctx)
		assert.True(t, monitor.IsHealthy())

		replica.AssertExpectations(t)
	})
}
