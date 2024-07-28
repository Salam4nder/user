//go:build testdb
// +build testdb

package db

import (
	"strings"
	"testing"
	"time"

	"github.com/Salam4nder/user/pkg/password"
	"github.com/Salam4nder/user/pkg/random"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestCreateUser(t *testing.T) {
	db, cleanup := NewTestSQLConnPool("users")
	t.Cleanup(cleanup)

	randomParams := CreateUserParams{
		ID:        uuid.New(),
		FullName:  random.FullName(),
		Email:     random.Email(),
		Password:  password.SafeString(random.String(10)),
		CreatedAt: time.Now().UTC(),
	}

	t.Run("ok", func(t *testing.T) {
		t.Cleanup(cleanup)

		err := CreateUser(ctx, db, randomParams)
		require.NoError(t, err)

		got, err := ReadUser(ctx, db, randomParams.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, randomParams.ID, got.ID)
		require.Equal(t, randomParams.FullName, got.FullName)
		require.Equal(t, randomParams.Email, got.Email)
		require.NotEqual(t, randomParams.Password, got.PasswordHash)
		require.True(t, time.Now().After(got.CreatedAt))

		require.NoError(
			t,
			bcrypt.CompareHashAndPassword([]byte(got.PasswordHash), []byte(randomParams.Password)),
		)
	})

	t.Run("name exceeds 255 chars returns err", func(t *testing.T) {
		t.Cleanup(cleanup)

		randomParams.FullName = strings.Repeat("a", 256)
		err := CreateUser(ctx, db, randomParams)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrStringTooLong)

		randomParams.FullName = random.FullName()
	})

	t.Run("email exceeds 255 chars returns err", func(t *testing.T) {
		t.Cleanup(cleanup)

		randomParams.Email = strings.Repeat("a", 256)

		err := CreateUser(ctx, db, randomParams)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrStringTooLong)
		randomParams.Email = random.Email()
	})

	t.Run("duplicate email returns error", func(t *testing.T) {
		t.Cleanup(cleanup)

		err := CreateUser(ctx, db, CreateUserParams{
			ID:        uuid.New(),
			FullName:  "Same User",
			Email:     "email@email.com",
			Password:  "password",
			CreatedAt: time.Now().UTC(),
		})
		require.NoError(t, err)

		err = CreateUser(ctx, db, CreateUserParams{
			ID:        uuid.New(),
			FullName:  "Same Email User",
			Email:     "email@email.com",
			Password:  "password",
			CreatedAt: time.Now().UTC(),
		})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrDuplicateEmail)
	})
}

func TestReadUser(t *testing.T) {
	db, cleanup := NewTestSQLConnPool("users")
	t.Cleanup(cleanup)

	randomParams := CreateUserParams{
		ID:        uuid.New(),
		FullName:  random.FullName(),
		Email:     random.Email(),
		Password:  password.SafeString(random.String(10)),
		CreatedAt: time.Now().UTC(),
	}

	err := CreateUser(ctx, db, randomParams)
	require.NoError(t, err)

	got, err := ReadUser(ctx, db, randomParams.ID)
	require.NoError(t, err)
	require.NotNil(t, got)

	t.Run("Not found", func(t *testing.T) {
		_, err := ReadUser(ctx, db, uuid.New())
		require.Error(t, err)
		require.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("InputError on nil UUID", func(t *testing.T) {
		_, err := ReadUser(ctx, db, uuid.Nil)
		require.Error(t, err)
		require.ErrorAs(t, err, &InputError{})
	})
}

func TestReadUserByEmail(t *testing.T) {
	db, cleanup := NewTestSQLConnPool("users")
	t.Cleanup(cleanup)

	randomParams := CreateUserParams{
		ID:        uuid.New(),
		FullName:  random.FullName(),
		Email:     random.Email(),
		Password:  password.SafeString(random.String(10)),
		CreatedAt: time.Now().UTC(),
	}

	err := CreateUser(ctx, db, randomParams)
	require.NoError(t, err)

	got, err := ReadUserByEmail(ctx, db, randomParams.Email)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, randomParams.ID, got.ID)
	require.Equal(t, randomParams.FullName, got.FullName)
	require.Equal(t, randomParams.Email, got.Email)
	require.True(t, time.Now().After(got.CreatedAt))

	t.Run("Not found", func(t *testing.T) {
		_, err := ReadUserByEmail(ctx, db, random.Email())
		require.Error(t, err)
		require.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("Email is empty", func(t *testing.T) {
		_, err := ReadUserByEmail(ctx, db, "")
		require.Error(t, err)
		require.ErrorAs(t, err, &InputError{})
	})

	t.Run("Email is not found", func(t *testing.T) {
		_, err := ReadUserByEmail(ctx, db, "ass@ass.com")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestSQL_UpdateUser(t *testing.T) {
	db, cleanup := NewTestSQLConnPool("users")
	t.Cleanup(cleanup)

	randomParams := CreateUserParams{
		ID:        uuid.New(),
		FullName:  random.FullName(),
		Email:     random.Email(),
		Password:  password.SafeString(random.String(10)),
		CreatedAt: time.Now().UTC(),
	}

	t.Run("ok", func(t *testing.T) {
		t.Cleanup(cleanup)

		newFullName := "New Name"
		newEmail := "new@email.com"

		err := CreateUser(ctx, db, randomParams)
		require.NoError(t, err)

		err = UpdateUser(ctx, db, UpdateUserParams{
			ID:       randomParams.ID,
			FullName: newFullName,
			Email:    newEmail,
		})
		require.NoError(t, err)

		got, err := ReadUser(ctx, db, randomParams.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, randomParams.ID, got.ID)
		require.Equal(t, newFullName, got.FullName)
		require.Equal(t, newEmail, got.Email)
		require.True(t, time.Now().After(got.CreatedAt))
	})

	t.Run("name exceeds 255 chars returns err", func(t *testing.T) {
		t.Cleanup(cleanup)

		ID := uuid.New()
		err := CreateUser(ctx, db, CreateUserParams{
			ID:        ID,
			FullName:  random.FullName(),
			Email:     random.Email(),
			Password:  password.SafeString(random.String(10)),
			CreatedAt: time.Now().UTC(),
		})
		require.NoError(t, err)

		err = UpdateUser(ctx, db, UpdateUserParams{
			ID:       ID,
			FullName: strings.Repeat("a", 256),
			Email:    random.Email(),
		})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrStringTooLong)
	})

	t.Run("email exceeds 255 chars returns err", func(t *testing.T) {
		t.Cleanup(cleanup)

		ID := uuid.New()

		err := CreateUser(ctx, db, CreateUserParams{
			ID:        ID,
			FullName:  random.FullName(),
			Email:     random.Email(),
			Password:  password.SafeString(random.String(10)),
			CreatedAt: time.Now().UTC(),
		})
		require.NoError(t, err)

		err = UpdateUser(ctx, db, UpdateUserParams{
			ID:       ID,
			FullName: random.FullName(),
			Email:    strings.Repeat("a", 256),
		})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrStringTooLong)
	})
}

func TestSQL_DeleteUser(t *testing.T) {
	db, cleanup := NewTestSQLConnPool("users")
	t.Cleanup(cleanup)

	ID := uuid.New()

	err := CreateUser(ctx, db, CreateUserParams{
		ID:        ID,
		FullName:  random.FullName(),
		Email:     random.Email(),
		Password:  password.SafeString(random.String(15)),
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	err = DeleteUser(ctx, db, ID)
	require.NoError(t, err)

	t.Run("Not found", func(t *testing.T) {
		err := DeleteUser(ctx, db, ID)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNoRowsAffected)
	})
}
