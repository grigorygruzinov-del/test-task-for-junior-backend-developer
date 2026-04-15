package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	taskdomain "github.com/medods/test-task-for-junior-backend-developer/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const taskQuery = `
		INSERT INTO tasks (title, description, status, scheduled_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, description, status, scheduled_at, created_at, updated_at
	`

	row := tx.QueryRow(ctx, taskQuery, 
		task.Title, 
		task.Description, 
		task.Status, 
		task.ScheduledAt, 
		task.CreatedAt, 
		task.UpdatedAt,
	)
	
	created, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	if task.Recurrence != nil {
		if err := r.insertRecurrence(ctx, tx, created.ID, task.Recurrence); err != nil {
			return nil, err
		}
		created.Recurrence = task.Recurrence
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `SELECT id, title, description, status, scheduled_at, created_at, updated_at FROM tasks WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}
	found.Recurrence, _ = r.getRecurrence(ctx, found.ID)
	return found, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const query = `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, scheduled_at = $4, updated_at = $5
		WHERE id = $6
		RETURNING id, title, description, status, scheduled_at, created_at, updated_at
	`

	row := tx.QueryRow(ctx, query, task.Title, task.Description, task.Status, task.ScheduledAt, task.UpdatedAt, task.ID)
	updated, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	_, _ = tx.Exec(ctx, `DELETE FROM task_recurrence_rules WHERE task_id = $1`, task.ID)
	if task.Recurrence != nil {
		if err := r.insertRecurrence(ctx, tx, updated.ID, task.Recurrence); err != nil {
			return nil, err
		}
		updated.Recurrence = task.Recurrence
	}

	return updated, tx.Commit(ctx)
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `SELECT id, title, description, status, scheduled_at, created_at, updated_at FROM tasks ORDER BY id DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []taskdomain.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		t.Recurrence, _ = r.getRecurrence(ctx, t.ID)
		tasks = append(tasks, *t)
	}
	return tasks, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	return err
}

func (r *Repository) insertRecurrence(ctx context.Context, tx pgx.Tx, taskID int64, rule *taskdomain.RecurrenceRule) error {
	const query = `
		INSERT INTO task_recurrence_rules (task_id, type, every_n_days, days_of_month, specific_dates, parity)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := tx.Exec(ctx, query, taskID, rule.Type, rule.EveryNDays, rule.DaysOfMonth, rule.SpecificDates, rule.Parity)
	return err
}

func (r *Repository) getRecurrence(ctx context.Context, taskID int64) (*taskdomain.RecurrenceRule, error) {
	var rule taskdomain.RecurrenceRule
	err := r.pool.QueryRow(ctx, `SELECT type, every_n_days, days_of_month, specific_dates, parity FROM task_recurrence_rules WHERE task_id = $1`, taskID).
		Scan(&rule.Type, &rule.EveryNDays, &rule.DaysOfMonth, &rule.SpecificDates, &rule.Parity)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

type taskScanner interface { Scan(dest ...any) error }

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		t      taskdomain.Task
		status string
	)
	if err := scanner.Scan(&t.ID, &t.Title, &t.Description, &status, &t.ScheduledAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, err
	}
	t.Status = taskdomain.Status(status)
	return &t, nil
}