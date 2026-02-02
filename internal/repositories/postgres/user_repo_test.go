package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/poyrazk/thecloud/internal/core/domain"
	theclouderrors "github.com/poyrazk/thecloud/internal/errors"
	"github.com/stretchr/testify/assert"
)

const (
	userRepoTestEmail        = "test@example.com"
	userRepoTestName         = "Test User"
	userRepoDBErrorMsg       = "db error"
	userRepoSelectByEmailSQL = "SELECT id, email, password_hash, name, role, default_tenant_id, created_at, updated_at FROM users WHERE email = \\$1"
	userRepoSelectByIDSQL    = "SELECT id, email, password_hash, name, role, default_tenant_id, created_at, updated_at FROM users WHERE id = \\$1"
	userRepoSelectAllSQL     = "SELECT id, email, password_hash, name, role, default_tenant_id, created_at, updated_at FROM users"
)

func TestUserRepoCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepo(mock)
		user := &domain.User{
			ID:           uuid.New(),
			Email:        userRepoTestEmail,
			PasswordHash: "hashed",
			Name:         userRepoTestName,
			Role:         "user",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mock.ExpectExec("INSERT INTO users").
			WithArgs(user.ID, user.Email, user.PasswordHash, user.Name, user.Role, user.DefaultTenantID, user.CreatedAt, user.UpdatedAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = repo.Create(context.Background(), user)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run(userRepoDBErrorMsg, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepo(mock)
		user := &domain.User{
			ID: uuid.New(),
		}

		mock.ExpectExec("INSERT INTO users").
			WillReturnError(errors.New(userRepoDBErrorMsg))

		err = repo.Create(context.Background(), user)
		assert.Error(t, err)
	})
}

func TestUserRepoGetByEmail(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepo(mock)
		email := userRepoTestEmail
		id := uuid.New()
		now := time.Now()

		mock.ExpectQuery(userRepoSelectByEmailSQL).
			WithArgs(email).
			WillReturnRows(pgxmock.NewRows([]string{"id", "email", "password_hash", "name", "role", "default_tenant_id", "created_at", "updated_at"}).
				AddRow(id, email, "hashed", userRepoTestName, "user", nil, now, now))

		user, err := repo.GetByEmail(context.Background(), email)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, id, user.ID)
		assert.Equal(t, email, user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepo(mock)
		email := userRepoTestEmail

		mock.ExpectQuery(userRepoSelectByEmailSQL).
			WithArgs(email).
			WillReturnError(pgx.ErrNoRows)

		user, err := repo.GetByEmail(context.Background(), email)
		assert.Error(t, err)
		assert.Nil(t, user)
		// Assuming repo returns custom error or native error
		// Checking if it matches domain logic. Usually repo wraps error.
	})

	t.Run(userRepoDBErrorMsg, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepo(mock)
		email := userRepoTestEmail

		mock.ExpectQuery(userRepoSelectByEmailSQL).
			WithArgs(email).
			WillReturnError(errors.New(userRepoDBErrorMsg))

		user, err := repo.GetByEmail(context.Background(), email)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserRepoGetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepo(mock)
		id := uuid.New()
		email := userRepoTestEmail
		now := time.Now()

		mock.ExpectQuery(userRepoSelectByIDSQL).
			WithArgs(id).
			WillReturnRows(pgxmock.NewRows([]string{"id", "email", "password_hash", "name", "role", "default_tenant_id", "created_at", "updated_at"}).
				AddRow(id, email, "hashed", userRepoTestName, "user", nil, now, now))

		user, err := repo.GetByID(context.Background(), id)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, id, user.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepo(mock)
		id := uuid.New()

		mock.ExpectQuery(userRepoSelectByIDSQL).
			WithArgs(id).
			WillReturnError(pgx.ErrNoRows)

		user, err := repo.GetByID(context.Background(), id)
		assert.Error(t, err)
		assert.Nil(t, user)
		theCloudErr, ok := err.(*theclouderrors.Error)
		if ok {
			assert.Equal(t, theclouderrors.NotFound, theCloudErr.Type)
		}
	})

	t.Run(userRepoDBErrorMsg, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepo(mock)
		id := uuid.New()

		mock.ExpectQuery(userRepoSelectByIDSQL).
			WithArgs(id).
			WillReturnError(errors.New(userRepoDBErrorMsg))

		user, err := repo.GetByID(context.Background(), id)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserRepoUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepo(mock)
		user := &domain.User{
			ID:           uuid.New(),
			Email:        "updated@example.com",
			PasswordHash: "newhash",
			Name:         "Updated User",
			Role:         "admin",
			UpdatedAt:    time.Now(),
		}

		mock.ExpectExec("UPDATE users").
			WithArgs(user.Email, user.PasswordHash, user.Name, user.Role, user.DefaultTenantID, user.UpdatedAt, user.ID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = repo.Update(context.Background(), user)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run(userRepoDBErrorMsg, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepo(mock)
		user := &domain.User{
			ID: uuid.New(),
		}

		mock.ExpectExec("UPDATE users").
			WillReturnError(errors.New(userRepoDBErrorMsg))

		err = repo.Update(context.Background(), user)
		assert.Error(t, err)
	})
}

func TestUserRepoList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepo(mock)
		id1 := uuid.New()
		id2 := uuid.New()
		now := time.Now()

		mock.ExpectQuery(userRepoSelectAllSQL).
			WillReturnRows(pgxmock.NewRows([]string{"id", "email", "password_hash", "name", "role", "default_tenant_id", "created_at", "updated_at"}).
				AddRow(id1, "u1@ex.com", "h1", "U1", "user", nil, now, now).
				AddRow(id2, "u2@ex.com", "h2", "U2", "admin", nil, now, now))

		users, err := repo.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, id1, users[0].ID)
		assert.Equal(t, id2, users[1].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run(userRepoDBErrorMsg, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepo(mock)

		mock.ExpectQuery(userRepoSelectAllSQL).
			WillReturnError(errors.New(userRepoDBErrorMsg))

		users, err := repo.List(context.Background())
		assert.Error(t, err)
		assert.Nil(t, users)
	})

	t.Run("scan error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepo(mock)
		now := time.Now()

		mock.ExpectQuery(userRepoSelectAllSQL).
			WillReturnRows(pgxmock.NewRows([]string{"id", "email", "password_hash", "name", "role", "created_at", "updated_at"}).
				AddRow("invalid-uuid", "u1@ex.com", "h1", "U1", "user", now, now))

		users, err := repo.List(context.Background())
		assert.Error(t, err)
		assert.Nil(t, users)
	})
}

func TestUserRepoDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepo(mock)
		id := uuid.New()

		mock.ExpectExec("DELETE FROM users").
			WithArgs(id).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = repo.Delete(context.Background(), id)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run(userRepoDBErrorMsg, func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		assert.NoError(t, err)
		defer mock.Close()

		repo := NewUserRepo(mock)
		id := uuid.New()

		mock.ExpectExec("DELETE FROM users").
			WithArgs(id).
			WillReturnError(errors.New(userRepoDBErrorMsg))

		err = repo.Delete(context.Background(), id)
		assert.Error(t, err)
	})
}
