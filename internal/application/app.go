package application

import (
	"context"

	"github.com/Roh-Bot/task-manager/internal/auth"
	"github.com/Roh-Bot/task-manager/internal/config"
	"github.com/Roh-Bot/task-manager/internal/entity"
	store2 "github.com/Roh-Bot/task-manager/internal/store"
	"github.com/Roh-Bot/task-manager/pkg/logger"
)

type App struct {
	Task ITaskUseCase
	Auth IAuthUseCase
}

type ITaskUseCase interface {
	Create(ctx context.Context, dto *CreateTaskDto) error
	GetByID(ctx context.Context, id, userID string, isAdmin bool) (*entity.Task, error)
	List(ctx context.Context, userID string, isAdmin bool, status string, limit int, scrollID *string) ([]entity.Task, string, error)
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
			repo:   store.Tasks,
		},
		Auth: &AuthUseCase{
			config: config,
			logger: logger,
			auth:   auth,
			repo:   store.Users,
		},
	}
}
