// Package service contains business logic for the loyalty system.
package service

import (
	"context"
	"errors"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/avitamin/go-gophermart/internal/domain/repository"
	"github.com/avitamin/go-gophermart/internal/pkg/auth"
)

// UserService errors.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserService handles user-related business logic.
type UserService struct {
	repo       repository.UserRepository
	jwtManager *auth.JWTManager
	hasher     auth.PasswordHasher
}

// NewUserService creates a new UserService.
func NewUserService(repo repository.UserRepository, jwtManager *auth.JWTManager, hasher auth.PasswordHasher) *UserService {
	return &UserService{
		repo:       repo,
		jwtManager: jwtManager,
		hasher:     hasher,
	}
}

// Register creates a new user and returns an authentication token.
func (s *UserService) Register(ctx context.Context, creds *entity.UserCredentials) (string, error) {
	if err := creds.Validate(); err != nil {
		return "", err
	}

	hash, err := s.hasher.Hash(creds.Password)
	if err != nil {
		return "", err
	}

	userID, err := s.repo.Create(ctx, creds.Login, hash)
	if err != nil {
		return "", err
	}

	return s.jwtManager.Generate(userID)
}

// Login authenticates a user and returns a token.
func (s *UserService) Login(ctx context.Context, creds *entity.UserCredentials) (string, error) {
	if creds.Login == "" || creds.Password == "" {
		return "", ErrInvalidCredentials
	}

	user, err := s.repo.GetByLogin(ctx, creds.Login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if !s.hasher.Compare(creds.Password, user.PasswordHash) {
		return "", ErrInvalidCredentials
	}

	return s.jwtManager.Generate(user.ID)
}
