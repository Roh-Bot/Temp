package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/Roh-Bot/blog-api/internal/auth"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

const (
	requestIdKey  = "request_id"
	maxLogPreview = 4096
)

type rateLimiterState struct {
	globalLimiter *rate.Limiter
	ipLimiters    map[string]*rate.Limiter
	mu            sync.Mutex
}

func (s *Server) getIPLimiter(ip string, state *rateLimiterState) *rate.Limiter {
	state.mu.Lock()
	defer state.mu.Unlock()

	limiter, exists := state.ipLimiters[ip]
	if !exists {
		cfg := s.Config.Get().RateLimit
		limiter = rate.NewLimiter(rate.Limit(cfg.IPRate), cfg.IPBurst)
		state.ipLimiters[ip] = limiter
	}
	return limiter
}

func (s *Server) rateLimiter(next echo.HandlerFunc) echo.HandlerFunc {
	cfg := s.Config.Get().RateLimit
	state := &rateLimiterState{
		globalLimiter: rate.NewLimiter(rate.Limit(cfg.GlobalRate), cfg.GlobalBurst),
		ipLimiters:    make(map[string]*rate.Limiter),
	}

	return func(ctx echo.Context) error {
		if !state.globalLimiter.Allow() {
			return ctx.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "global rate limit exceeded",
			})
		}

		ip := ctx.RealIP()
		limiter := s.getIPLimiter(ip, state)
		if !limiter.Allow() {
			return ctx.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "rate limit exceeded for your IP",
			})
		}

		return next(ctx)
	}
}

func (s *Server) validateAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		authHeader := ctx.Request().Header.Get("Authorization")
		if authHeader == "" {
			return s.unauthorized(ctx, nil, "missing authorization header")
		}
		if len(authHeader) < len("Bearer ") || authHeader[:7] != "Bearer " {
			return s.unauthorized(ctx, nil, "invalid authorization format")
		}
		token := authHeader[7:]
		claims, err := s.App.Auth.ValidateToken(token)

		if err != nil {
			s.Logger.Error(ctx.Request().Context(), err.Error())
			if errors.Is(err, auth.ErrTokenExpired) {
				return s.unauthorized(ctx, err, auth.ErrTokenExpired.Error())
			}
			return s.unauthorized(ctx, err, "invalid token")
		}

		ctx.Set("user_id", claims["user_id"])
		ctx.Set("username", claims["username"])
		ctx.Set("is_admin", claims["role"] == "admin")

		return next(ctx)
	}
}

func (s *Server) httpLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		req := ctx.Request()
		res := ctx.Response()

		reqID := uuid.New().String()
		ctxUser := context.WithValue(req.Context(), requestIdKey, reqID)
		ctx.SetRequest(req.WithContext(ctxUser))

		start := time.Now()

		err := next(ctx)
		latency := time.Since(start)

		logEntry := map[string]any{
			"request_id": reqID,
			"time":       time.Now().Format(time.RFC3339),
			"method":     req.Method,
			"url":        req.URL.String(),
			"remote_ip":  ctx.RealIP(),
			"latency_ms": latency.Milliseconds(),
			"request": map[string]any{
				"headers": req.Header,
			},
			"response": map[string]any{
				"status": res.Status,
			},
		}

		if err != nil {
			logEntry["error"] = err.Error()
			s.Logger.Error(ctxUser, toJSON(logEntry))
		} else {
			s.Logger.Info(ctxUser, toJSON(logEntry))
		}

		return err
	}
}

func toJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
