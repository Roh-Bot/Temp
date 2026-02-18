package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
	return args.Get(0).(map[string]interface{}), args.Error(1)
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
