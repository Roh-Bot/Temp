package worker

import (
	"log"
	"time"

	"github.com/Roh-Bot/blog-api/internal/store"
	"github.com/Roh-Bot/blog-api/pkg/global"
	"github.com/Roh-Bot/blog-api/pkg/logger"
)

type TaskWorker struct {
	store           store.ITaskStore
	logger          logger.Logger
	autoCompleteMin int
	taskChan        chan string
}

func NewTaskWorker(store store.ITaskStore, logger logger.Logger, autoCompleteMin int) *TaskWorker {
	return &TaskWorker{
		store:           store,
		logger:          logger,
		autoCompleteMin: autoCompleteMin,
		taskChan:        make(chan string, 100),
	}
}

func (w *TaskWorker) Start(ctx *global.ApplicationContext) {
	go w.processTaskQueue(ctx)
	go w.scanPendingTasks(ctx)
}

func (w *TaskWorker) processTaskQueue(ctx *global.ApplicationContext) {
	defer ctx.Done()
	for {
		select {
		case <-ctx.Context().Done():
			log.Println("task processor scanner shutting down")
			return
		case taskID := <-w.taskChan:
			if err := w.store.AutoCompleteIfPending(ctx.Context(), taskID); err != nil {
				w.logger.Error(ctx.Context(), "Failed to auto-complete task "+taskID+": "+err.Error())
			} else {
				w.logger.Info(ctx.Context(), "Auto-completed task: "+taskID)
			}
		}
	}
}

func (w *TaskWorker) scanPendingTasks(ctx *global.ApplicationContext) {
	defer ctx.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Context().Done():
			log.Println("pending task scanner shutting down")
			return
		case <-ticker.C:
			tasks, err := w.store.GetPendingTasks(ctx.Context(), w.autoCompleteMin)
			if err != nil {
				w.logger.Error(ctx.Context(), "Failed to fetch pending tasks: "+err.Error())
				continue
			}

			for _, task := range tasks {
				select {
				case w.taskChan <- task.ID:
				default:
					w.logger.Error(ctx.Context(), "Task queue full, skipping task: "+task.ID)
				}
			}
		}
	}
}
