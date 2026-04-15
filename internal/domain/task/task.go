package task

import (
	"time"
)

// Status определяет возможные состояния задачи
type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusCancelled  Status = "cancelled"
)

// Task представляет собой основную сущность задачи в системе
type Task struct {
	ID          int64           `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Status      Status          `json:"status"`
	ScheduledAt *time.Time      `json:"scheduled_at"` // Указатель, так как дата может быть не задана (NULL)
	Recurrence  *RecurrenceRule `json:"recurrence"`   // Указатель, так как правил может не быть
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// RecurrenceRule описывает настройки периодичности задачи
type RecurrenceRule struct {
	// Тип повторения: "daily", "monthly", "specific_dates", "parity"
	Type string `json:"type"`

	// Для типа "daily": каждый n-ый день
	EveryNDays *int `json:"every_n_days,omitempty"`

	// Для типа "monthly": массив чисел месяца (от 1 до 31)
	DaysOfMonth []int `json:"days_of_month,omitempty"`

	// Для типа "specific_dates": массив конкретных дат
	SpecificDates []time.Time `json:"specific_dates,omitempty"`

	// Для типа "parity": "even" (четные) или "odd" (нечетные)
	Parity *string `json:"parity,omitempty"`
}