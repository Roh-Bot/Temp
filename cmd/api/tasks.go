package api

import (
	"strconv"

	"github.com/Roh-Bot/blog-api/internal/application"
	"github.com/labstack/echo/v4"
)

type CreateTaskRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
}

type TaskResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	UserID      string `json:"user_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type TaskListResponse struct {
	Tasks      []TaskResponse `json:"tasks"`
	Total      int            `json:"total"`
	Limit      int            `json:"limit"`
	NextScroll string         `json:"next_scroll,omitempty"`
}

// @Summary Create a task
// @Description Create a new task
// @Tags Tasks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param task body CreateTaskRequest true "Task details"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Router /tasks [post]
func (s *Server) createTask(ctx echo.Context) error {
	userID := ctx.Get("user_id").(string)

	req := new(CreateTaskRequest)
	if err := ctx.Bind(req); err != nil {
		return s.badRequest(ctx, err, err.Error())
	}
	if err := s.Validator.Struct(req); err != nil {
		return s.badRequest(ctx, err, validationToErrorMessage(err))
	}

	dto := &application.CreateTaskDto{
		Title:       req.Title,
		Description: req.Description,
		UserID:      userID,
	}

	if err := s.App.Task.Create(ctx.Request().Context(), dto); err != nil {
		return s.internalServerError(ctx, err, err.Error())
	}

	return s.created(ctx, nil)
}

// @Summary List tasks
// @Description Get list of tasks with pagination and filtering
// @Tags Tasks
// @Produce json
// @Security ApiKeyAuth
// @Param scroll_id query string false "Scroll ID for pagination"
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status (pending, in_progress, completed)"
// @Success 200 {object} Response{data=TaskListResponse}
// @Failure 401 {object} Response
// @Router /tasks [get]
func (s *Server) listTasks(ctx echo.Context) error {
	userID := ctx.Get("user_id").(string)
	isAdmin := ctx.Get("is_admin").(bool)

	limit, _ := strconv.Atoi(ctx.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}
	scrollID := ctx.QueryParam("scroll_id")
	status := ctx.QueryParam("status")

	tasks, total, nextScroll, err := s.App.Task.List(ctx.Request().Context(), userID, isAdmin, status, limit, scrollID)
	if err != nil {
		return s.internalServerError(ctx, err, err.Error())
	}

	var taskResponses []TaskResponse
	for _, task := range tasks {
		taskResponses = append(taskResponses, TaskResponse{
			ID:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			Status:      task.Status,
			UserID:      task.UserID,
			CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	response := TaskListResponse{
		Tasks:      taskResponses,
		Total:      total,
		Limit:      limit,
		NextScroll: nextScroll,
	}

	return s.writeResponse(ctx, response)
}

// @Summary Get task by ID
// @Description Get a specific task by ID
// @Tags Tasks
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Task ID"
// @Success 200 {object} Response{data=TaskResponse}
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Router /tasks/{id} [get]
func (s *Server) getTask(ctx echo.Context) error {
	userID := ctx.Get("user_id").(string)
	isAdmin := ctx.Get("is_admin").(bool)
	id := ctx.Param("id")

	task, err := s.App.Task.GetByID(ctx.Request().Context(), id, userID, isAdmin)
	if err != nil {
		return s.notFound(ctx, err.Error())
	}

	response := TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		UserID:      task.UserID,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return s.writeResponse(ctx, response)
}

// @Summary Delete task
// @Description Delete a task by ID
// @Tags Tasks
// @Security ApiKeyAuth
// @Param id path string true "Task ID"
// @Success 204
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Router /tasks/{id} [delete]
func (s *Server) deleteTask(ctx echo.Context) error {
	userID := ctx.Get("user_id").(string)
	isAdmin := ctx.Get("is_admin").(bool)
	id := ctx.Param("id")

	if err := s.App.Task.Delete(ctx.Request().Context(), id, userID, isAdmin); err != nil {
		return s.internalServerError(ctx, err, err.Error())
	}

	return ctx.NoContent(204)
}
