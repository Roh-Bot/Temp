package application

import (
	"context"
	"errors"
	"github.com/Roh-Bot/blog-api/internal/auth"
	config2 "github.com/Roh-Bot/blog-api/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockEncryption struct {
	EncryptedValue string
	Err            error
}

func (m *MockEncryption) Encrypt(data string) ([]byte, error) {
	return []byte(data), m.Err
}

type MockJWT struct {
	Token       string
	TokenErr    error
	Valid       bool
	ValidateErr error
}

func (m *MockJWT) GenerateToken(claims jwt.MapClaims) (string, error) {
	return m.Token, m.TokenErr
}

func (m *MockJWT) ValidateToken(token string) (bool, error) {
	return m.Valid, m.ValidateErr
}

func TestAuthService_IsValid(t *testing.T) {
	config, _ := config2.LoadConfiguration(context.Background())
	service := AuthUseCase{config: config}

	assert.True(t, service.IsValid("devadiga.rohit"))
	assert.False(t, service.IsValid("invalid"))
}

func TestAuthService_GenerateToken_Success(t *testing.T) {
	config, _ := config2.LoadConfiguration(context.Background())

	mockEncrytion := &MockEncryption{
		EncryptedValue: "devadiga.rohit",
		Err:            nil,
	}
	mockJwt := &MockJWT{
		Token: "mockToken",
	}
	mockAuth := auth.Authentication{
		JWT:        mockJwt,
		Encryption: mockEncrytion,
	}

	authService := &AuthUseCase{
		auth:   mockAuth,
		config: config,
	}

	token, err := authService.GenerateToken("devadiga.rohit")
	assert.NoError(t, err)
	assert.Equal(t, "mockToken", token)
}

func TestAuthService_GenerateToken_EncryptionError(t *testing.T) {
	config, _ := config2.LoadConfiguration(context.Background())

	mockEncrytion := &MockEncryption{
		Err: errors.New("encryption failed"),
	}
	mockJwt := &MockJWT{
		Token: "mockToken",
	}
	mockAuth := auth.Authentication{
		JWT:        mockJwt,
		Encryption: mockEncrytion,
	}

	authService := &AuthUseCase{
		auth:   mockAuth,
		config: config,
	}

	encryptedValue, err := authService.GenerateToken("user")
	assert.Error(t, err)
	assert.Empty(t, encryptedValue)
}

func TestAuthService_GenerateToken_JWTError(t *testing.T) {
	config, _ := config2.LoadConfiguration(context.Background())

	mockEncrytion := &MockEncryption{
		EncryptedValue: "gaming",
	}
	mockJwt := &MockJWT{
		TokenErr: errors.New("failed to generate token"),
	}
	mockAuth := auth.Authentication{
		JWT:        mockJwt,
		Encryption: mockEncrytion,
	}

	authService := &AuthUseCase{
		auth:   mockAuth,
		config: config,
	}

	token, err := authService.GenerateToken("user123")
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestAuthService_ValidateToken(t *testing.T) {
	config, _ := config2.LoadConfiguration(context.Background())

	mockEncrytion := &MockEncryption{
		EncryptedValue: "gaming",
	}
	mockJwt := &MockJWT{
		Valid: true,
	}
	mockAuth := auth.Authentication{
		JWT:        mockJwt,
		Encryption: mockEncrytion,
	}

	authService := &AuthUseCase{
		auth:   mockAuth,
		config: config,
	}

	valid, err := authService.ValidateToken("sometoken")
	assert.True(t, valid)
	assert.NoError(t, err)
}

func TestAuthService_ValidateToken_Error(t *testing.T) {
	config, _ := config2.LoadConfiguration(context.Background())

	mockEncrytion := &MockEncryption{
		EncryptedValue: "gaming",
	}
	mockJwt := &MockJWT{
		ValidateErr: errors.New("failed to generate token"),
	}
	mockAuth := auth.Authentication{
		JWT:        mockJwt,
		Encryption: mockEncrytion,
	}

	authService := &AuthUseCase{
		auth:   mockAuth,
		config: config,
	}

	valid, err := authService.ValidateToken("sometoken")
	assert.False(t, valid)
	assert.Error(t, err)
}
