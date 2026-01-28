package postgres

import (
	"context"
	"testing"

	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
)

func TestNewDualDB(t *testing.T) {
	primary, _ := pgxmock.NewPool()
	replica, _ := pgxmock.NewPool()
	defer primary.Close()
	defer replica.Close()

	dual := NewDualDB(primary, replica)
	assert.NotNil(t, dual)
	assert.Equal(t, primary, dual.primary)
	assert.Equal(t, replica, dual.replica)

	// Test fallback
	dual2 := NewDualDB(primary, nil)
	assert.Equal(t, primary, dual2.primary)
	assert.Equal(t, primary, dual2.replica)
}

const testQuery = "SELECT id FROM test"

func TestDualDBOperations(t *testing.T) {
	primary, _ := pgxmock.NewPool()
	replica, _ := pgxmock.NewPool()
	defer primary.Close()
	defer replica.Close()

	dual := NewDualDB(primary, replica)
	ctx := context.Background()

	// Exec should go to primary
	primary.ExpectExec("INSERT").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	_, err := dual.Exec(ctx, "INSERT INTO test DEFAULT VALUES")
	assert.NoError(t, err)

	// Query should go to replica
	replica.ExpectQuery("SELECT").WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
	rows, err := dual.Query(ctx, testQuery)
	assert.NoError(t, err)
	rows.Close()

	// QueryRow should go to replica
	replica.ExpectQuery("SELECT").WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
	row := dual.QueryRow(ctx, "SELECT id FROM test WHERE id = 1")
	var id int
	err = row.Scan(&id)
	assert.NoError(t, err)

	// Begin should go to primary
	primary.ExpectBegin()
	_, err = dual.Begin(ctx)
	assert.NoError(t, err)

	// Ping should go to primary
	primary.ExpectPing()
	err = dual.Ping(ctx)
	assert.NoError(t, err)

	// Close should close both
	primary.ExpectClose()
	replica.ExpectClose()
	dual.Close()
}

func TestDualDBCloseSame(t *testing.T) {
	primary, _ := pgxmock.NewPool()
	defer primary.Close()

	dual := NewDualDB(primary, nil)
	primary.ExpectClose()
	dual.Close()
}

func TestDualDBStatus(t *testing.T) {
	primary, _ := pgxmock.NewPool()
	replica, _ := pgxmock.NewPool()
	defer primary.Close()
	defer replica.Close()

	t.Run("Not Configured", func(t *testing.T) {
		dual := NewDualDB(primary, nil)
		status := dual.GetStatus(context.Background())
		assert.Equal(t, "NOT_CONFIGURED", status["database_replica"])
	})

	t.Run("Healthy", func(t *testing.T) {
		dual := NewDualDB(primary, replica)
		dual.SetReplicaHealthy(true)
		status := dual.GetStatus(context.Background())
		assert.Equal(t, "CONNECTED", status["database_replica"])
	})

	t.Run("Unhealthy", func(t *testing.T) {
		dual := NewDualDB(primary, replica)
		dual.SetReplicaHealthy(false)
		status := dual.GetStatus(context.Background())
		assert.Equal(t, "UNHEALTHY", status["database_replica"])
	})
}

func TestDualDBFailover(t *testing.T) {
	primary, _ := pgxmock.NewPool()
	replica, _ := pgxmock.NewPool()
	defer primary.Close()
	defer replica.Close()

	dual := NewDualDB(primary, replica)
	ctx := context.Background()

	t.Run("Primary Fallback on Replica Error", func(t *testing.T) {
		dual.SetReplicaHealthy(true) // Start healthy

		// Replica returns error
		replica.ExpectQuery("SELECT").WillReturnError(assert.AnError)
		// Primary should be called as fallback
		primary.ExpectQuery("SELECT").WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))

		rows, err := dual.Query(ctx, testQuery)
		assert.NoError(t, err)
		rows.Close()

		assert.NoError(t, primary.ExpectationsWereMet())
		assert.NoError(t, replica.ExpectationsWereMet())
	})

	t.Run("Wait for Circuit Breaker to Trip", func(t *testing.T) {
		dual.SetReplicaHealthy(true)

		// Trip it (threshold is 3)
		for i := 0; i < 3; i++ {
			replica.ExpectQuery("SELECT").WillReturnError(assert.AnError)
			primary.ExpectQuery("SELECT").WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
			_, _ = dual.Query(ctx, testQuery)
		}

		// Now it should go straight to primary without calling replica
		primary.ExpectQuery("SELECT").WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
		_, _ = dual.Query(ctx, testQuery)

		assert.NoError(t, replica.ExpectationsWereMet()) // Replica should not have extra calls
		assert.NoError(t, primary.ExpectationsWereMet())
	})
}
