package store

import (
	"context"
	"errors"

	"github.com/Roh-Bot/blog-api/internal/config"
	"github.com/Roh-Bot/blog-api/internal/database"
	"github.com/Roh-Bot/blog-api/internal/entity"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUserNotFound          = errors.New("user not found")
)

type UserStore struct {
	db     *database.Database
	config *config.AtomicConfig
}

func (s *UserStore) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `SELECT * FROM user_get_by_username($1)`
	var user entity.User
	err := s.db.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == stateNoDataFound {
			return nil, ErrUserNotFound
		}
	}
	return &user, err
}

func (s *UserStore) Create(ctx context.Context, user *entity.User) error {
	query := `SELECT * FROM user_create($1, $2, $3, $4, $5)`
	_, err := s.db.Exec(ctx, query, user.ID, user.Username, user.Email, user.Password, user.Role)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == stateUniqueViolation {
			switch pgErr.ConstraintName {
			case "users_username_unique":
				return ErrUsernameAlreadyExists
			case "users_email_unique":
				return ErrEmailAlreadyExists
			}
		}
	}
	return err
}
