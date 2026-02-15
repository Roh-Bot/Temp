package application

import (
	"context"

	"github.com/Roh-Bot/blog-api/internal/auth"
	"github.com/Roh-Bot/blog-api/internal/config"
	"github.com/Roh-Bot/blog-api/internal/entity"
	store2 "github.com/Roh-Bot/blog-api/internal/store"
	"github.com/Roh-Bot/blog-api/pkg/logger"
)

type App struct {
	Task ITaskUseCase
	Auth IAuthUseCase
}

type ITaskUseCase interface {
	Create(ctx context.Context, dto *CreateTaskDto) error
	GetByID(ctx context.Context, id, userID string, isAdmin bool) (*entity.Task, error)
	List(ctx context.Context, userID string, isAdmin bool, status string, limit int, scrollID string) ([]entity.Task, int, string, error)
	Delete(ctx context.Context, id, userID string, isAdmin bool) error
}

type IAuthUseCase interface {
	Login(ctx context.Context, username, password string) (string, error)
	Register(ctx context.Context, username, email, password, role string) error
	ValidateToken(token string) (map[string]interface{}, error)
}

func NewService(config *config.AtomicConfig, auth auth.Authentication, store store2.Store, logger logger.Logger) App {
	return App{
		Task: &TaskUseCase{
			logger: logger,
			config: config,
			store:  store,
		},
		Auth: &AuthUseCase{
			config: config,
			logger: logger,
			auth:   auth,
			store:  store,
		},
	}
}
