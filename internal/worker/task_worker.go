package worker

import (
	"context"
	"github.com/Roh-Bot/blog-api/internal/store"
	"github.com/Roh-Bot/blog-api/pkg/logger"
	"time"
)

type TaskWorker struct {
	store           store.Store
	logger          logger.Logger
	autoCompleteMin int
	taskChan        chan string
}

func NewTaskWorker(store store.Store, logger logger.Logger, autoCompleteMin int) *TaskWorker {
	return &TaskWorker{
		store:           store,
		logger:          logger,
		autoCompleteMin: autoCompleteMin,
		taskChan:        make(chan string, 100),
	}
}

func (w *TaskWorker) Start(ctx context.Context) {
	go w.processTaskQueue(ctx)
	go w.scanPendingTasks(ctx)
}

func (w *TaskWorker) processTaskQueue(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case taskID := <-w.taskChan:
			if err := w.store.Tasks.UpdateStatus(ctx, taskID, "completed"); err != nil {
				w.logger.Error(ctx, "Failed to auto-complete task "+taskID+": "+err.Error())
			} else {
				w.logger.Info(ctx, "Auto-completed task: "+taskID)
			}
		}
	}
}

func (w *TaskWorker) scanPendingTasks(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tasks, err := w.store.Tasks.GetPendingTasks(ctx, w.autoCompleteMin)
			if err != nil {
				w.logger.Error(ctx, "Failed to fetch pending tasks: "+err.Error())
				continue
			}

			for _, task := range tasks {
				select {
				case w.taskChan <- task.ID:
				default:
					w.logger.Error(ctx, "Task queue full, skipping task: "+task.ID)
				}
			}
		}
	}
}
