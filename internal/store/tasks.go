package store

import (
	"context"
	"errors"

	"github.com/Roh-Bot/blog-api/internal/config"
	"github.com/Roh-Bot/blog-api/internal/database"
	"github.com/Roh-Bot/blog-api/internal/entity"
	"github.com/jackc/pgx/v5"
)

type TaskStore struct {
	db     *database.Database
	config *config.AtomicConfig
}

func (s *TaskStore) Create(ctx context.Context, task *entity.Task) error {
	_, err := s.db.Exec(ctx,
		`SELECT * FROM task_create($1,$2,$3,$4,$5,$6,$7)`,
		task.ID,
		task.Title,
		task.Description,
		task.Status,
		task.UserID,
		task.CreatedAt,
		task.UpdatedAt,
	)
	return err
}

func (s *TaskStore) GetByID(ctx context.Context, id, userID string, isAdmin bool) (*entity.Task, error) {
	query := `
		SELECT id, title, description, status, user_id, created_at, updated_at
		FROM task_get_by_id($1, $2, $3)
	`

	var task entity.Task

	err := s.db.QueryRow(ctx, query, id, userID, isAdmin).
		Scan(&task.ID, &task.Title, &task.Description,
			&task.Status, &task.UserID,
			&task.CreatedAt, &task.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (s *TaskStore) List(
	ctx context.Context,
	userID string,
	isAdmin bool,
	status string,
	limit int,
	scrollId string,
) ([]entity.Task, int, string, error) {

	query := `
		SELECT id, title, description, status, user_id, created_at, updated_at, total_count, next_scroll_token
		FROM list_tasks_paginated($1, $2, $3, $4, $5)
	`

	rows, err := s.db.Query(ctx, query, userID, isAdmin, status, limit, scrollId)
	if err != nil {
		return nil, 0, "", err
	}
	defer rows.Close()

	var tasks []entity.Task
	var total int
	var nextScrollToken string

	for rows.Next() {
		var task entity.Task
		var totalCount int64
		var nextToken *string

		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.UserID,
			&task.CreatedAt,
			&task.UpdatedAt,
			&totalCount,
			&nextToken,
		); err != nil {
			return nil, 0, "", err
		}

		tasks = append(tasks, task)
		total = int(totalCount)

		if nextToken != nil {
			nextScrollToken = *nextToken
		}
	}

	if err := rows.Err(); err != nil {
		return nil, 0, "", err
	}

	return tasks, total, nextScrollToken, nil
}

func (s *TaskStore) Delete(ctx context.Context, id, userID string, isAdmin bool) error {
	_, err := s.db.Exec(ctx,
		`SELECT * FROM task_delete($1, $2, $3)`,
		id,
		userID,
		isAdmin,
	)

	return err
}

func (s *TaskStore) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := s.db.Exec(ctx,
		`SELECT * FROM task_update_status($1, $2)`,
		id,
		status,
	)
	return err
}

func (s *TaskStore) AutoCompleteIfPending(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx,
		`SELECT * FROM task_auto_complete_if_pending($1)`,
		id,
	)
	return err
}

func (s *TaskStore) GetPendingTasks(ctx context.Context, olderThan int) ([]entity.Task, error) {
	query := `
		SELECT id, title, description, status, user_id, created_at, updated_at
		FROM task_get_pending_older_than($1)
	`

	rows, err := s.db.Query(ctx, query, olderThan)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []entity.Task

	for rows.Next() {
		var task entity.Task
		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.UserID,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}
