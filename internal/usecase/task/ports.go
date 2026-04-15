package task

import (
	"context"
	"time"

	// Исправленный путь импорта, соответствующий твоему go.mod
	taskdomain "github.com/medods/test-task-for-junior-backend-developer/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

// CreateInput включает настройки периодичности для создания задачи
type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	ScheduledAt *time.Time
	Recurrence  *taskdomain.RecurrenceRule
}

// UpdateInput позволяет обновлять поля задачи и менять правила повторения
type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
	ScheduledAt *time.Time
	Recurrence  *taskdomain.RecurrenceRule
}