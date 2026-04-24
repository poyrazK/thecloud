//go:build integration
// +build integration

package services_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poyrazk/thecloud/internal/repositories/postgres"
)

func setupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	// Use helper from postgres package
	db, _ := postgres.SetupDB(t)
	return db
}

func setupTestUser(t *testing.T, db *pgxpool.Pool) context.Context {
	t.Helper()
	return postgres.SetupTestUser(t, db)
}


func cleanDB(t *testing.T, db *pgxpool.Pool) { 
	t.Helper() 
	postgres.CleanDB(t, db) 
} 

