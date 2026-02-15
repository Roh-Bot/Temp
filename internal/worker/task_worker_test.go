package worker

import (
	"context"
	"testing"
	"time"

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

func TestTaskWorker_ProcessTaskQueue(t *testing.T) {
	mockStore := new(MockTaskStore)
	worker := NewTaskWorker(MockStore{Tasks: mockStore}, &logger.MockLogger{}, 5)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	t.Run("success", func(t *testing.T) {
		mockStore.On("UpdateStatus", mock.Anything, "task1", "completed").Return(nil)

		go worker.processTaskQueue(ctx)
		worker.taskChan <- "task1"

		time.Sleep(50 * time.Millisecond)
		mockStore.AssertExpectations(t)
	})
}

func TestTaskWorker_ScanPendingTasks(t *testing.T) {
	mockStore := new(MockTaskStore)
	worker := &TaskWorker{
		store:           MockStore{Tasks: mockStore},
		logger:          &logger.MockLogger{},
		autoCompleteMin: 5,
		taskChan:        make(chan string, 100),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	t.Run("fetches and queues tasks", func(t *testing.T) {
		pendingTasks := []entity.Task{
			{ID: "task1", Status: "pending", CreatedAt: time.Now().Add(-10 * time.Minute)},
			{ID: "task2", Status: "in_progress", CreatedAt: time.Now().Add(-10 * time.Minute)},
		}
		mockStore.On("GetPendingTasks", mock.Anything, 5).Return(pendingTasks, nil)

		// Override ticker for testing
		worker.scanPendingTasks = func(ctx context.Context) {
			tasks, _ := worker.store.Tasks.GetPendingTasks(ctx, worker.autoCompleteMin)
			for _, task := range tasks {
				worker.taskChan <- task.ID
			}
		}

		worker.scanPendingTasks(ctx)

		assert.Equal(t, 2, len(worker.taskChan))
		mockStore.AssertExpectations(t)
	})

	t.Run("handles empty result", func(t *testing.T) {
		mockStore.On("GetPendingTasks", mock.Anything, 5).Return([]entity.Task{}, nil)

		worker.scanPendingTasks = func(ctx context.Context) {
			tasks, _ := worker.store.Tasks.GetPendingTasks(ctx, worker.autoCompleteMin)
			for _, task := range tasks {
				worker.taskChan <- task.ID
			}
		}

		worker.scanPendingTasks(ctx)

		assert.Equal(t, 0, len(worker.taskChan))
		mockStore.AssertExpectations(t)
	})
}

func TestTaskWorker_UpdateStatus(t *testing.T) {
	mockStore := new(MockTaskStore)
	worker := NewTaskWorker(MockStore{Tasks: mockStore}, &logger.MockLogger{}, 5)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockStore.On("UpdateStatus", ctx, "task1", "completed").Return(nil)

		err := worker.store.Tasks.UpdateStatus(ctx, "task1", "completed")

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockStore.On("UpdateStatus", ctx, "task2", "completed").Return(assert.AnError)

		err := worker.store.Tasks.UpdateStatus(ctx, "task2", "completed")

		assert.Error(t, err)
		mockStore.AssertExpectations(t)
	})
}
