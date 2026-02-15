package application

import (
	"context"
	"errors"
	"time"

	"github.com/Roh-Bot/blog-api/internal/auth"
	"github.com/Roh-Bot/blog-api/internal/config"
	"github.com/Roh-Bot/blog-api/internal/entity"
	"github.com/Roh-Bot/blog-api/internal/store"
	"github.com/Roh-Bot/blog-api/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	config *config.AtomicConfig
	logger logger.Logger
	auth   auth.Authentication
	store  store.Store
}

func (u *AuthUseCase) Login(ctx context.Context, username, password string) (string, error) {
	user, err := u.store.Users.GetByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Minute * time.Duration(u.config.Get().Auth.TokenTTL)).Unix(),
		"iss":      u.config.Get().Auth.Issuer,
		"aud":      u.config.Get().Auth.Audience,
	}

	return u.auth.JWT.GenerateToken(claims)
}

func (u *AuthUseCase) Register(ctx context.Context, username, email, password, role string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &entity.User{
		ID:       uuid.New().String(),
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
	}

	return u.store.Users.Create(ctx, user)
}

func (u *AuthUseCase) ValidateToken(token string) (map[string]interface{}, error) {
	claims, err := u.auth.JWT.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
