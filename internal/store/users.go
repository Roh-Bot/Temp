package store

import (
	"context"
	"github.com/Roh-Bot/blog-api/internal/config"
	"github.com/Roh-Bot/blog-api/internal/database"
	"github.com/Roh-Bot/blog-api/internal/entity"
	"github.com/jackc/pgx/v5"
)

type UserStore struct {
	db     *database.Database
	config *config.AtomicConfig
}

func (s *UserStore) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `SELECT id, username, email, password, role FROM users WHERE username = $1`
	var user entity.User
	err := s.db.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (s *UserStore) Create(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (id, username, email, password, role) VALUES ($1, $2, $3, $4, $5)`
	_, err := s.db.Exec(ctx, query, user.ID, user.Username, user.Email, user.Password, user.Role)
	return err
}
