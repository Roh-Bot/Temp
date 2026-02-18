package store

import (
	"context"

	"github.com/Roh-Bot/task-manager/internal/config"
	"github.com/Roh-Bot/task-manager/internal/database"
	"github.com/Roh-Bot/task-manager/internal/entity"
)

var (
	stateUniqueViolation = "23505"
	stateNoDataFound     = "P0002"
)

type Store struct {
	Tasks ITaskStore
	Users IUserStore
}

type ITaskStore interface {
	Create(ctx context.Context, task *entity.Task) error
	GetByID(ctx context.Context, id, userID string, isAdmin bool) (*entity.Task, error)
	List(ctx context.Context, userID string, isAdmin bool, status string, limit int, scrollID *string) ([]entity.Task, string, error)
	Delete(ctx context.Context, id, userID string, isAdmin bool) error
	UpdateStatus(ctx context.Context, id, status string) error
	AutoCompleteIfPending(ctx context.Context, id string) error
	GetPendingTasks(ctx context.Context, olderThan int) ([]entity.Task, error)
}

type IUserStore interface {
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
}

func NewStorage(db *database.Database, config *config.AtomicConfig) Store {
	return Store{
		Tasks: &TaskStore{db: db, config: config},
		Users: &UserStore{db: db, config: config},
	}
}
