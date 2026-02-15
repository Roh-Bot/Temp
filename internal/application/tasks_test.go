package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Roh-Bot/blog-api/internal/config"
	"github.com/Roh-Bot/blog-api/internal/entity"
	"github.com/Roh-Bot/blog-api/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTaskStore struct {
	mock.Mock
}

func (m *MockTaskStore) Create(ctx context.Context, task *entity.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskStore) GetByID(ctx context.Context, id, userID string, isAdmin bool) (*entity.Task, error) {
	args := m.Called(ctx, id, userID, isAdmin)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Task), args.Error(1)
}

func (m *MockTaskStore) List(ctx context.Context, userID string, isAdmin bool, limit int, scrollID, status string) ([]entity.Task, int, string, error) {
	args := m.Called(ctx, userID, isAdmin, limit, scrollID, status)
	return args.Get(0).([]entity.Task), args.Int(1), args.String(2), args.Error(3)
}

func (m *MockTaskStore) Delete(ctx context.Context, id, userID string, isAdmin bool) error {
	args := m.Called(ctx, id, userID, isAdmin)
	return args.Error(0)
}

func (m *MockTaskStore) UpdateStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockTaskStore) GetPendingTasks(ctx context.Context, olderThan int) ([]entity.Task, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).([]entity.Task), args.Error(1)
}

type MockStore struct {
	Tasks *MockTaskStore
}

func setupTaskUseCase() (*TaskUseCase, *MockTaskStore) {
	mockStore := new(MockTaskStore)
	cfg := &config.AtomicConfig{}
	cfg.Set(&config.Config{})

	useCase := &TaskUseCase{
		logger: &logger.MockLogger{},
		config: cfg,
		store: MockStore{
			Tasks: mockStore,
		},
	}

	return useCase, mockStore
}

func TestTaskUseCase_Create(t *testing.T) {
	useCase, mockStore := setupTaskUseCase()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockStore.On("Create", ctx, mock.MatchedBy(func(task *entity.Task) bool {
			return task.Title == "Test Task" && task.Status == "pending"
		})).Return(nil)

		dto := &CreateTaskDto{
			Title:       "Test Task",
			Description: "Test Description",
			UserID:      "user123",
		}

		err := useCase.Create(ctx, dto)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("store error", func(t *testing.T) {
		mockStore.On("Create", ctx, mock.Anything).Return(errors.New("db error"))

		dto := &CreateTaskDto{
			Title:       "Test Task",
			Description: "Test Description",
			UserID:      "user123",
		}

		err := useCase.Create(ctx, dto)

		assert.Error(t, err)
		mockStore.AssertExpectations(t)
	})
}

func TestTaskUseCase_GetByID(t *testing.T) {
	useCase, mockStore := setupTaskUseCase()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expectedTask := &entity.Task{
			ID:          "1",
			Title:       "Task 1",
			Description: "Description",
			Status:      "pending",
			UserID:      "user123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		mockStore.On("GetByID", ctx, "1", "user123", false).Return(expectedTask, nil)

		task, err := useCase.GetByID(ctx, "1", "user123", false)

		assert.NoError(t, err)
		assert.Equal(t, expectedTask.ID, task.ID)
		assert.Equal(t, expectedTask.Title, task.Title)
		mockStore.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockStore.On("GetByID", ctx, "999", "user123", false).Return(nil, nil)

		task, err := useCase.GetByID(ctx, "999", "user123", false)

		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Equal(t, "task not found", err.Error())
		mockStore.AssertExpectations(t)
	})
}

func TestTaskUseCase_List(t *testing.T) {
	useCase, mockStore := setupTaskUseCase()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expectedTasks := []entity.Task{
			{ID: "1", Title: "Task 1", Status: "pending", UserID: "user123"},
			{ID: "2", Title: "Task 2", Status: "completed", UserID: "user123"},
		}
		mockStore.On("List", ctx, "user123", false, 10, "", "pending").Return(expectedTasks, 2, "next123", nil)

		tasks, total, nextScroll, err := useCase.List(ctx, "user123", false, 10, "", "pending")

		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		assert.Equal(t, 2, total)
		assert.Equal(t, "next123", nextScroll)
		mockStore.AssertExpectations(t)
	})

	t.Run("empty result", func(t *testing.T) {
		mockStore.On("List", ctx, "user123", false, 10, "", "").Return([]entity.Task{}, 0, "", nil)

		tasks, total, nextScroll, err := useCase.List(ctx, "user123", false, 10, "", "")

		assert.NoError(t, err)
		assert.Empty(t, tasks)
		assert.Equal(t, 0, total)
		assert.Empty(t, nextScroll)
		mockStore.AssertExpectations(t)
	})
}

func TestTaskUseCase_Delete(t *testing.T) {
	useCase, mockStore := setupTaskUseCase()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockStore.On("Delete", ctx, "1", "user123", false).Return(nil)

		err := useCase.Delete(ctx, "1", "user123", false)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("store error", func(t *testing.T) {
		mockStore.On("Delete", ctx, "999", "user123", false).Return(errors.New("not found"))

		err := useCase.Delete(ctx, "999", "user123", false)

		assert.Error(t, err)
		mockStore.AssertExpectations(t)
	})
}
