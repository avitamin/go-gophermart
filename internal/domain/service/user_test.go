package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/avitamin/go-gophermart/internal/domain/repository"
	"github.com/avitamin/go-gophermart/internal/domain/service/mocks"
	"github.com/avitamin/go-gophermart/internal/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserService_Register(t *testing.T) {
	ctx := context.Background()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	t.Run("successful registration", func(t *testing.T) {
		userRepo := new(mocks.UserRepository)
		hasher := new(mocks.PasswordHasher)

		userRepo.On("Create", ctx, "testuser", "hashed_password").Return(int64(1), nil)
		hasher.On("Hash", "12345678901234567890").Return("hashed_password", nil)

		service := NewUserService(userRepo, jwtManager, hasher)

		token, err := service.Register(ctx, &entity.UserCredentials{
			Login:    "testuser",
			Password: "12345678901234567890",
		})

		require.NoError(t, err)
		assert.NotEmpty(t, token)
		userRepo.AssertExpectations(t)
		hasher.AssertExpectations(t)
	})

	t.Run("user already exists", func(t *testing.T) {
		userRepo := new(mocks.UserRepository)
		hasher := new(mocks.PasswordHasher)

		userRepo.On("Create", ctx, "testuser", "hashed_password").Return(int64(0), repository.ErrUserExists)
		hasher.On("Hash", "12345678901234567890").Return("hashed_password", nil)

		service := NewUserService(userRepo, jwtManager, hasher)

		_, err := service.Register(ctx, &entity.UserCredentials{
			Login:    "testuser",
			Password: "12345678901234567890",
		})

		assert.ErrorIs(t, err, repository.ErrUserExists)
	})

	t.Run("invalid credentials - empty login", func(t *testing.T) {
		userRepo := new(mocks.UserRepository)
		hasher := new(mocks.PasswordHasher)

		service := NewUserService(userRepo, jwtManager, hasher)

		_, err := service.Register(ctx, &entity.UserCredentials{
			Login:    "",
			Password: "12345678901234567890",
		})

		assert.Error(t, err)
	})

	t.Run("invalid credentials - password too short", func(t *testing.T) {
		userRepo := new(mocks.UserRepository)
		hasher := new(mocks.PasswordHasher)

		service := NewUserService(userRepo, jwtManager, hasher)

		_, err := service.Register(ctx, &entity.UserCredentials{
			Login:    "testuser",
			Password: "short",
		})

		assert.ErrorIs(t, err, entity.ErrPasswordTooShort)
	})
}

func TestUserService_Login(t *testing.T) {
	ctx := context.Background()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	t.Run("successful login", func(t *testing.T) {
		userRepo := new(mocks.UserRepository)
		hasher := new(mocks.PasswordHasher)

		user := &entity.User{
			ID:           1,
			Login:        "testuser",
			PasswordHash: "hashed_password",
		}

		userRepo.On("GetByLogin", ctx, "testuser").Return(user, nil)
		hasher.On("Compare", "password123", "hashed_password").Return(true)

		service := NewUserService(userRepo, jwtManager, hasher)

		token, err := service.Login(ctx, &entity.UserCredentials{
			Login:    "testuser",
			Password: "password123",
		})

		require.NoError(t, err)
		assert.NotEmpty(t, token)
		userRepo.AssertExpectations(t)
		hasher.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		userRepo := new(mocks.UserRepository)
		hasher := new(mocks.PasswordHasher)

		userRepo.On("GetByLogin", ctx, "unknown").Return(nil, repository.ErrUserNotFound)

		service := NewUserService(userRepo, jwtManager, hasher)

		_, err := service.Login(ctx, &entity.UserCredentials{
			Login:    "unknown",
			Password: "password123",
		})

		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("wrong password", func(t *testing.T) {
		userRepo := new(mocks.UserRepository)
		hasher := new(mocks.PasswordHasher)

		user := &entity.User{
			ID:           1,
			Login:        "testuser",
			PasswordHash: "hashed_password",
		}

		userRepo.On("GetByLogin", ctx, "testuser").Return(user, nil)
		hasher.On("Compare", "wrong_password", "hashed_password").Return(false)

		service := NewUserService(userRepo, jwtManager, hasher)

		_, err := service.Login(ctx, &entity.UserCredentials{
			Login:    "testuser",
			Password: "wrong_password",
		})

		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("empty credentials", func(t *testing.T) {
		userRepo := new(mocks.UserRepository)
		hasher := new(mocks.PasswordHasher)

		service := NewUserService(userRepo, jwtManager, hasher)

		_, err := service.Login(ctx, &entity.UserCredentials{
			Login:    "",
			Password: "",
		})

		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("database error", func(t *testing.T) {
		userRepo := new(mocks.UserRepository)
		hasher := new(mocks.PasswordHasher)

		dbErr := errors.New("database connection error")
		userRepo.On("GetByLogin", ctx, "testuser").Return(nil, dbErr)

		service := NewUserService(userRepo, jwtManager, hasher)

		_, err := service.Login(ctx, &entity.UserCredentials{
			Login:    "testuser",
			Password: "password123",
		})

		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrInvalidCredentials)
	})
}

func TestUserService_HashError(t *testing.T) {
	ctx := context.Background()
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	userRepo := new(mocks.UserRepository)
	hasher := new(mocks.PasswordHasher)

	hashErr := errors.New("hash error")
	hasher.On("Hash", mock.Anything).Return("", hashErr)

	service := NewUserService(userRepo, jwtManager, hasher)

	_, err := service.Register(ctx, &entity.UserCredentials{
		Login:    "testuser",
		Password: "12345678901234567890",
	})

	assert.Error(t, err)
}
