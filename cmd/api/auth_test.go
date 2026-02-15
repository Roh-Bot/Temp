package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Roh-Bot/blog-api/internal/application"
	"github.com/Roh-Bot/blog-api/internal/validator"
	"github.com/Roh-Bot/blog-api/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	authUrl = "/api/authentication/login"
)

type MockAuthService struct {
	ShouldValidate bool
	Token          string
	TokenErr       error
}

func (m *MockAuthService) IsValid(username string) bool {
	return m.ShouldValidate
}

func (m *MockAuthService) GenerateToken(username string) (string, error) {
	return m.Token, m.TokenErr
}

func (m *MockAuthService) ValidateToken(token string) (bool, error) {
	return true, nil
}

func setupAuthTestServer(auth application.IAuthUseCase) *echo.Echo {
	mockService := application.App{
		Auth: auth,
	}

	e := echo.New()
	server := &Server{
		App:       mockService,
		Validator: validator.NewValidator(),
		Logger:    &logger.MockLogger{},
		Router:    e,
	}

	e.POST(authUrl, server.authLoginUser)
	return e
}

func TestAuthLoginUser_Success(t *testing.T) {
	mockAuth := &MockAuthService{
		ShouldValidate: true,
		Token:          "mock-token",
	}

	app := setupAuthTestServer(mockAuth)

	body := AuthLoginUserRequest{Username: "devadiga.rohit"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, authUrl, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthLoginUser_InvalidUsername(t *testing.T) {
	mockAuth := &MockAuthService{
		ShouldValidate: false,
	}

	app := setupAuthTestServer(mockAuth)

	body := AuthLoginUserRequest{Username: "invalid"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, authUrl, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthLoginUser_GenerateTokenError(t *testing.T) {
	mockAuth := &MockAuthService{
		ShouldValidate: false,
		TokenErr:       errors.New("token generation failed"),
	}

	app := setupAuthTestServer(mockAuth)

	body := AuthLoginUserRequest{Username: "invalid"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, authUrl, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthLoginUser_InvalidPayload(t *testing.T) {
	mockAuth := &MockAuthService{}
	app := setupAuthTestServer(mockAuth)

	req := httptest.NewRequest(http.MethodPost, authUrl, bytes.NewReader([]byte(`invalid-json`)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthLoginUser_ValidationFailure(t *testing.T) {
	mockAuth := &MockAuthService{}
	app := setupAuthTestServer(mockAuth)

	body := map[string]string{
		"Username": "",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, authUrl, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
