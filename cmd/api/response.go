package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	errBadRequest          = "Invalid request received"
	errInternalServerError = "Internal Server Error"
	errUnauthorized        = "Unauthorized"
	stringEmpty            = ""
)

// Response represents a standard response model
// swagger:model Response
type Response struct {
	// Status
	// example: 1 / -1
	Status int `json:"Status"`

	// Error message
	// example: Something went wrong
	Error string `json:"Error,omitempty"`

	// Data actual response data
	// example: Success
	Data any `json:"Data,omitempty"`
}

func (s *Server) writeResponse(ctx echo.Context, data any) error {
	return ctx.JSON(http.StatusOK, Response{
		Status: 1,
		Data:   data,
	})
}

func (s *Server) writeResponseWithStatusCode(ctx echo.Context, statusCode int, data any) error {
	return ctx.JSON(statusCode, Response{
		Status: 1,
		Data:   data,
	})
}

func (s *Server) writeErrorResponse(ctx echo.Context, statusCode int, error string) error {
	return ctx.JSON(statusCode, Response{
		Status: -1,
		Error:  error,
	})
}

func (s *Server) internalServerError(ctx echo.Context, err error) error {
	s.Logger.Error(ctx.Request().Context(), err.Error())

	return s.writeErrorResponse(ctx, http.StatusInternalServerError, errInternalServerError)
}

func (s *Server) badRequest(ctx echo.Context, err error, errorMessage string) error {
	if err != nil {
		errorMessage = err.Error()
	}
	if errorMessage == "" {
		errorMessage = errBadRequest
	}
	return s.writeErrorResponse(ctx, http.StatusBadRequest, errBadRequest)
}

func (s *Server) unauthorized(ctx echo.Context, errorMessage string) error {
	return s.writeErrorResponse(ctx, http.StatusUnauthorized, errorMessage)
}

func (s *Server) notFound(ctx echo.Context, errorMessage string) error {
	return s.writeErrorResponse(ctx, http.StatusNotFound, errorMessage)
}

func (s *Server) conflict(ctx echo.Context, errorMessage string) error {
	return s.writeErrorResponse(ctx, http.StatusConflict, errorMessage)
}

//func (s *Server) handleServiceError(c echo.Context, err error) error {
//	// Map service errors to appropriate HTTP status codes
//	switch err {
//	case application.ErrInvalidInput:
//		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
//	case application.ErrUnauthorized:
//		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
//	case application.ErrForbidden:
//		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
//	case application.ErrNotFound:
//		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
//	case application.ErrInternalServer:
//		fallthrough // Fallback for unknown errors
//	default:
//		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
//	}
//}

func (s *Server) created(ctx echo.Context, data any) error {
	return ctx.JSON(http.StatusCreated, Response{
		Status: 1,
		Data:   data,
	})
}
