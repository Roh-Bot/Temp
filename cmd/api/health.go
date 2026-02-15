package api

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func (s *Server) Health(ctx echo.Context) error {
	return ctx.NoContent(http.StatusOK)
}
