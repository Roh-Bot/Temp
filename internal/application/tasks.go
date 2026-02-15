package application

import (
	"context"
	"errors"
	"time"

	"github.com/Roh-Bot/blog-api/internal/config"
	"github.com/Roh-Bot/blog-api/internal/entity"
	"github.com/Roh-Bot/blog-api/internal/store"
	"github.com/Roh-Bot/blog-api/pkg/logger"
	"github.com/google/uuid"
)

type TaskUseCase struct {
	logger logger.Logger
	config *config.AtomicConfig
	repo   store.ITaskStore
}

type CreateTaskDto struct {
	Title       string
	Description string
	UserID      string
}

func (u *TaskUseCase) Create(ctx context.Context, dto *CreateTaskDto) error {
	task := &entity.Task{
		ID:          uuid.New().String(),
		Title:       dto.Title,
		Description: dto.Description,
		Status:      "pending",
		UserID:      dto.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return u.repo.Create(ctx, task)
}

func (u *TaskUseCase) GetByID(ctx context.Context, id, userID string, isAdmin bool) (*entity.Task, error) {
	task, err := u.repo.GetByID(ctx, id, userID, isAdmin)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	return task, nil
}

func (u *TaskUseCase) List(ctx context.Context, userID string, isAdmin bool, status string, limit int, scrollID string) ([]entity.Task, int, string, error) {
	return u.repo.List(ctx, userID, isAdmin, status, limit, scrollID)
}

func (u *TaskUseCase) Delete(ctx context.Context, id, userID string, isAdmin bool) error {
	return u.repo.Delete(ctx, id, userID, isAdmin)
}
