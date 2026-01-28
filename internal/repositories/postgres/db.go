// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/poyrazk/thecloud/internal/platform"
)

// DB defines the interface for database operations, allowing for mocking in tests.
// It is compatible with both *pgxpool.Pool and pgxmock.
type DB interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
	Ping(ctx context.Context) error
}

// DualDB implements DB and routes reads to a replica if available.
type DualDB struct {
	primary        DB
	replica        DB
	replicaHealthy atomic.Bool
	cb             *platform.CircuitBreaker
}

// NewDualDB creates a DualDB that routes reads to the replica when provided.
func NewDualDB(primary, replica DB) *DualDB {
	d := &DualDB{
		primary: primary,
		replica: replica,
		cb:      platform.NewCircuitBreaker(3, 30*time.Second),
	}
	d.replicaHealthy.Store(replica != nil)
	if replica == nil {
		d.replica = primary
	}
	return d
}

// SetReplicaHealthy updates the health status of the replica.
func (d *DualDB) SetReplicaHealthy(healthy bool) {
	if d.replica == d.primary {
		return // No separate replica
	}
	d.replicaHealthy.Store(healthy)
	if healthy {
		d.cb.Reset()
	}
}

// GetReplica returns the replica DB instance.
func (d *DualDB) GetReplica() DB {
	return d.replica
}

// GetStatus returns the health status of database components.
func (d *DualDB) GetStatus(ctx context.Context) map[string]string {
	status := make(map[string]string)
	if d.replica == d.primary {
		status["database_replica"] = "NOT_CONFIGURED"
		return status
	}

	if d.replicaHealthy.Load() {
		status["database_replica"] = "CONNECTED"
	} else {
		status["database_replica"] = "UNHEALTHY"
	}
	return status
}

func (d *DualDB) getReadDB() DB {
	if d.replicaHealthy.Load() && d.cb.GetState() != platform.StateOpen {
		return d.replica
	}
	return d.primary
}

func (d *DualDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return d.primary.Exec(ctx, sql, args...)
}

func (d *DualDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	db := d.getReadDB()
	if db == d.primary {
		return db.Query(ctx, sql, args...)
	}

	var rows pgx.Rows
	err := d.cb.Execute(func() error {
		var qErr error
		rows, qErr = db.Query(ctx, sql, args...)
		return qErr
	})

	if err != nil {
		// If replica failed, retry once on primary
		return d.primary.Query(ctx, sql, args...)
	}
	return rows, nil
}

func (d *DualDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	db := d.getReadDB()
	if db == d.primary {
		return db.QueryRow(ctx, sql, args...)
	}

	// QueryRow is harder because errors happen during Scan
	// We'll just use the standard call, but if it was the replica,
	// we won't record failure here. The next Query will catch it.
	return db.QueryRow(ctx, sql, args...)
}

func (d *DualDB) Begin(ctx context.Context) (pgx.Tx, error) {
	return d.primary.Begin(ctx)
}

func (d *DualDB) Close() {
	d.primary.Close()
	if d.replica != d.primary {
		d.replica.Close()
	}
}

func (d *DualDB) Ping(ctx context.Context) error {
	return d.primary.Ping(ctx)
}
