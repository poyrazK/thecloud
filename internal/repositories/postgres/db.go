// Package postgres provides PostgreSQL-backed repository implementations.
package postgres

import (
	"context"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
}

// NewDualDB creates a DualDB that routes reads to the replica when provided.
func NewDualDB(primary, replica DB) *DualDB {
	d := &DualDB{primary: primary, replica: replica}
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
}

// GetReplica returns the replica DB instance.
func (d *DualDB) GetReplica() DB {
	return d.replica
}

func (d *DualDB) getReadDB() DB {
	if d.replicaHealthy.Load() {
		return d.replica
	}
	return d.primary
}

func (d *DualDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return d.primary.Exec(ctx, sql, args...)
}

func (d *DualDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return d.getReadDB().Query(ctx, sql, args...)
}

func (d *DualDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return d.getReadDB().QueryRow(ctx, sql, args...)
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
