// Package postgres provides PostgreSQL database adapter.
package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/avitamin/go-gophermart/internal/domain/entity"
	"github.com/avitamin/go-gophermart/internal/domain/repository"
)

// UserRepository implements repository.UserRepository for PostgreSQL.
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user and returns their ID.
func (r *UserRepository) Create(ctx context.Context, login, passwordHash string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id",
		login, passwordHash,
	).Scan(&id)

	if err != nil {
		if IsUniqueViolation(err) {
			return 0, repository.ErrUserExists
		}
		r.db.LogError("create user", err)
		return 0, err
	}
	return id, nil
}

// GetByLogin retrieves a user by their login.
func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*entity.User, error) {
	user := &entity.User{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, login, password_hash, created_at FROM users WHERE login = $1",
		login,
	).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		r.db.LogError("get user by login", err)
		return nil, err
	}
	return user, nil
}

// GetByID retrieves a user by their ID.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	user := &entity.User{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, login, password_hash, created_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		r.db.LogError("get user by id", err)
		return nil, err
	}
	return user, nil
}
