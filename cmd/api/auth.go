package api

import (
	"github.com/labstack/echo/v4"
)

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=user admin"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

// @Summary User login
// @Description Authenticate user and get JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} Response{data=AuthResponse}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Router /auth/login [post]
func (s *Server) login(ctx echo.Context) error {
	req := new(LoginRequest)
	if err := ctx.Bind(req); err != nil {
		return s.badRequest(ctx, err, err.Error())
	}
	if err := s.Validator.Struct(req); err != nil {
		return s.badRequest(ctx, err, validationToErrorMessage(err))
	}
	
	token, err := s.App.Auth.Login(ctx.Request().Context(), req.Username, req.Password)
	if err != nil {
		return s.unauthorized(ctx, err, err.Error())
	}
	
	return s.writeResponse(ctx, AuthResponse{Token: token})
}

// @Summary User registration
// @Description Register a new user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User details"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Router /auth/register [post]
func (s *Server) register(ctx echo.Context) error {
	req := new(RegisterRequest)
	if err := ctx.Bind(req); err != nil {
		return s.badRequest(ctx, err, err.Error())
	}
	if err := s.Validator.Struct(req); err != nil {
		return s.badRequest(ctx, err, validationToErrorMessage(err))
	}
	
	if err := s.App.Auth.Register(ctx.Request().Context(), req.Username, req.Email, req.Password, req.Role); err != nil {
		return s.internalServerError(ctx, err, err.Error())
	}
	
	return s.created(ctx, nil)
}
