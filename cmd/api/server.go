package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Roh-Bot/blog-api/internal/application"
	"github.com/Roh-Bot/blog-api/internal/config"
	"github.com/Roh-Bot/blog-api/pkg/global"
	"github.com/Roh-Bot/blog-api/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/swaggo/echo-swagger"
)

type Server struct {
	Config    *config.AtomicConfig
	App       application.App
	Validator *validator.Validate
	Logger    logger.Logger
	AppCtx    *global.ApplicationContext
	Router    *echo.Echo
}

func NewServer(config *config.AtomicConfig, services application.App, validator *validator.Validate, logger logger.Logger, appCtx *global.ApplicationContext) *Server {
	return &Server{
		Config:    config,
		App:       services,
		Router:    echo.New(),
		Validator: validator,
		Logger:    logger,
		AppCtx:    appCtx,
	}
}

func (s *Server) Run() {
	defer s.AppCtx.Done()
	go func() {
		s.registerMiddlewares()
		s.registerSwagger()
		s.registerHandlers()
		if err := s.Router.Start(s.Config.Get().Application.Address); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	for range s.AppCtx.Context().Done() {
		if err := s.Shutdown(); err != nil {
			log.Println("An error occurred while shutting down the server")
		}
		break
	}
	log.Println("Server shutdown completed")
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return s.Router.Shutdown(ctx)
}

func (s *Server) registerSwagger() {
	s.Router.GET("/swagger/*", echoSwagger.WrapHandler)
}

func (s *Server) registerHandlers() {
	apiGroup := s.Router.Group("/api")
	apiGroup.Use(s.rateLimiter)
	apiGroup.Use(s.httpLogger)

	apiGroup.GET("/health", s.Health)

	authGroup := apiGroup.Group("/auth")
	authGroup.POST("/login", s.login)
	authGroup.POST("/register", s.register)

	protectedGroup := apiGroup.Group("")
	protectedGroup.Use(s.validateAuth)
	protectedGroup.POST("/tasks", s.createTask)
	protectedGroup.GET("/tasks", s.listTasks)
	protectedGroup.GET("/tasks/:id", s.getTask)
	protectedGroup.DELETE("/tasks/:id", s.deleteTask)
}

func (s *Server) registerMiddlewares() {
	s.Router.Use(middleware.Recover())
	s.Router.Use(middleware.CORS())
}
