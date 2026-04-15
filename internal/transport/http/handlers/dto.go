package handlers

import "time"

// taskDTO используется для отправки данных задачи клиенту (ответ API)
type taskDTO struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
	ScheduledAt *time.Time     `json:"scheduled_at"` // Указатель, так как дата может быть null
	Recurrence  *recurrenceDTO `json:"recurrence,omitempty"`
}

// taskMutationDTO используется для получения данных от клиента (при создании и обновлении)
type taskMutationDTO struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
	ScheduledAt *time.Time     `json:"scheduled_at"` // Указатель для обработки null
	Recurrence  *recurrenceDTO `json:"recurrence,omitempty"`
}

// recurrenceDTO описывает структуру настроек повторения в JSON
type recurrenceDTO struct {
	Type          string      `json:"type"`                     // "daily", "monthly", "specific_dates", "parity"
	EveryNDays    *int        `json:"every_n_days,omitempty"`   // Указатель, так как может отсутствовать
	DaysOfMonth   []int       `json:"days_of_month,omitempty"`  // Массив чисел (например, [1, 15])
	SpecificDates []time.Time `json:"specific_dates,omitempty"` // Массив конкретных дат
	Parity        *string     `json:"parity,omitempty"`         // "even" (четные) или "odd" (нечетные)
}

// newTaskDTO используется для возврата ID только что созданной задачи
type newTaskDTO struct {
	ID string `json:"id"`
}