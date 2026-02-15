package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"

	"github.com/Roh-Bot/blog-api/internal/application"
	"github.com/Roh-Bot/blog-api/internal/config"
	"github.com/Roh-Bot/blog-api/internal/entity"
	"github.com/Roh-Bot/blog-api/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTaskUseCase struct {
	mock.Mock
}

func (m *MockTaskUseCase) Create(ctx context.Context, dto *application.CreateTaskDto) error {
	args := m.Called(ctx, dto)
	return args.Error(0)
}

func (m *MockTaskUseCase) GetByID(ctx context.Context, id, userID string, isAdmin bool) (*entity.Task, error) {
	args := m.Called(ctx, id, userID, isAdmin)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Task), args.Error(1)
}

func (m *MockTaskUseCase) List(ctx context.Context, userID string, isAdmin bool, status string, limit int, scrollID string) ([]entity.Task, int, string, error) {
	args := m.Called(ctx, userID, isAdmin, limit, scrollID, status)
	return args.Get(0).([]entity.Task), args.Int(1), args.String(2), args.Error(3)
}

func (m *MockTaskUseCase) Delete(ctx context.Context, id, userID string, isAdmin bool) error {
	args := m.Called(ctx, id, userID, isAdmin)
	return args.Error(0)
}

type MockAuthUseCase struct {
	mock.Mock
}

func (m *MockAuthUseCase) Login(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthUseCase) Register(ctx context.Context, username, email, password, role string) error {
	args := m.Called(ctx, username, email, password, role)
	return args.Error(0)
}

func (m *MockAuthUseCase) ValidateToken(token string) (map[string]interface{}, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func setupTestServer() (*Server, *MockTaskUseCase, *MockAuthUseCase) {
	mockTask := new(MockTaskUseCase)
	mockAuth := new(MockAuthUseCase)

	cfg := &config.AtomicConfig{}
	cfg.Set(&config.Config{})

	server := &Server{
		Config: cfg,
		App: application.App{
			Task: mockTask,
			Auth: mockAuth,
		},
		Validator: validator.New(),
		Logger:    &logger.MockLogger{},
		Router:    echo.New(),
	}

	return server, mockTask, mockAuth
}

func TestCreateTask(t *testing.T) {
	server, mockTask, _ := setupTestServer()

	t.Run("success", func(t *testing.T) {
		mockTask.On("Create", mock.Anything, mock.MatchedBy(func(dto *application.CreateTaskDto) bool {
			return dto.Title == "Test Task" && dto.UserID == "user123"
		})).Return(nil)

		body := `{"title":"Test Task","description":"Test Description"}`
		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := server.Router.NewContext(req, rec)
		c.Set("user_id", "user123")

		err := server.createTask(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
		mockTask.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		body := `{"title":""}`
		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := server.Router.NewContext(req, rec)
		c.Set("user_id", "user123")

		server.createTask(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestListTasks(t *testing.T) {
	server, mockTask, _ := setupTestServer()

	t.Run("success", func(t *testing.T) {
		tasks := []entity.Task{
			{ID: "1", Title: "Task 1", Status: "pending", UserID: "user123", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		mockTask.On("List", mock.Anything, "user123", false, 10, "", "").Return(tasks, 1, "next123", nil)

		req := httptest.NewRequest(http.MethodGet, "/tasks?limit=10", nil)
		rec := httptest.NewRecorder()
		c := server.Router.NewContext(req, rec)
		c.Set("user_id", "user123")
		c.Set("is_admin", false)

		err := server.listTasks(c)

		assert.NoError(t, err)
		mockTask.AssertExpectations(t)
	})
}

func TestGetTask(t *testing.T) {
	server, mockTask, _ := setupTestServer()

	t.Run("success", func(t *testing.T) {
		task := &entity.Task{ID: "1", Title: "Task 1", Status: "pending", UserID: "user123", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		mockTask.On("GetByID", mock.Anything, "1", "user123", false).Return(task, nil)

		req := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
		rec := httptest.NewRecorder()
		c := server.Router.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")
		c.Set("user_id", "user123")
		c.Set("is_admin", false)

		err := server.getTask(c)

		assert.NoError(t, err)
		mockTask.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockTask.On("GetByID", mock.Anything, "999", "user123", false).Return(nil, errors.New("not found"))

		req := httptest.NewRequest(http.MethodGet, "/tasks/999", nil)
		rec := httptest.NewRecorder()
		c := server.Router.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("999")
		c.Set("user_id", "user123")
		c.Set("is_admin", false)

		server.getTask(c)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		mockTask.AssertExpectations(t)
	})
}

func TestDeleteTask(t *testing.T) {
	server, mockTask, _ := setupTestServer()

	t.Run("success", func(t *testing.T) {
		mockTask.On("Delete", mock.Anything, "1", "user123", false).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
		rec := httptest.NewRecorder()
		c := server.Router.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")
		c.Set("user_id", "user123")
		c.Set("is_admin", false)

		err := server.deleteTask(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)
		mockTask.AssertExpectations(t)
	})
}

func TestLogin(t *testing.T) {
	server, _, mockAuth := setupTestServer()

	t.Run("success", func(t *testing.T) {
		mockAuth.On("Login", mock.Anything, "user1", "password123").Return("token123", nil)

		body := `{"username":"user1","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := server.Router.NewContext(req, rec)

		err := server.login(c)

		assert.NoError(t, err)

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		assert.Equal(t, "token123", response["token"])
		mockAuth.AssertExpectations(t)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		mockAuth.On("Login", mock.Anything, "user1", "wrong").Return("", errors.New("invalid credentials"))

		body := `{"username":"user1","password":"wrong"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := server.Router.NewContext(req, rec)

		server.login(c)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		mockAuth.AssertExpectations(t)
	})
}

func TestRegister(t *testing.T) {
	server, _, mockAuth := setupTestServer()

	t.Run("success", func(t *testing.T) {
		mockAuth.On("Register", mock.Anything, "newuser", "new@example.com", "password123", "user").Return(nil)

		body := `{"username":"newuser","email":"new@example.com","password":"password123"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := server.Router.NewContext(req, rec)

		err := server.register(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
		mockAuth.AssertExpectations(t)
	})
}
