package task

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	taskdomain "github.com/medods/test-task-for-junior-backend-developer/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) CreateTask(ctx context.Context, t taskdomain.Task) (*taskdomain.Task, error) {
	t.Title = strings.TrimSpace(t.Title)
	if t.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	if t.Status == "" {
		t.Status = taskdomain.StatusNew
	}

	if err := s.validateRecurrence(t.Recurrence); err != nil {
		return nil, err
	}

	if t.ScheduledAt == nil && t.Recurrence != nil {
		nextRun := s.calculateNextRun(t.Recurrence, s.now())
		t.ScheduledAt = &nextRun
	}

	now := s.now()
	t.CreatedAt = now
	t.UpdatedAt = now

	return s.repo.Create(ctx, &t)
}

func (s *Service) Update(ctx context.Context, id int64, t taskdomain.Task) (*taskdomain.Task, error) {
	oldTask, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	t.ID = id
	t.UpdatedAt = s.now()

	if t.Status == taskdomain.StatusDone && oldTask.Status != taskdomain.StatusDone {
		if oldTask.Recurrence != nil {
			go s.spawnNextOccurrence(context.Background(), *oldTask)
		}
	}

	return s.repo.Update(ctx, &t)
}

func (s *Service) spawnNextOccurrence(ctx context.Context, old taskdomain.Task) {
	baseTime := s.now()
	if old.ScheduledAt != nil {
		baseTime = *old.ScheduledAt
	}
	
	nextDate := s.calculateNextRun(old.Recurrence, baseTime)
	
	if old.ScheduledAt != nil && nextDate.Equal(*old.ScheduledAt) {
		return 
	}

	newTask := taskdomain.Task{
		Title:       old.Title,
		Description: old.Description,
		Status:      taskdomain.StatusNew,
		ScheduledAt: &nextDate,
		Recurrence:  old.Recurrence,
	}
	
	_, _ = s.CreateTask(ctx, newTask)
}

func (s *Service) calculateNextRun(rule *taskdomain.RecurrenceRule, from time.Time) time.Time {
	if rule == nil {
		return from.AddDate(0, 0, 1)
	}

	switch rule.Type {
	case "daily":
		days := 1
		if rule.EveryNDays != nil && *rule.EveryNDays > 0 {
			days = *rule.EveryNDays
		}
		return from.AddDate(0, 0, days)

	case "monthly":
		if len(rule.DaysOfMonth) == 0 {
			return from.AddDate(0, 1, 0)
		}
		sort.Ints(rule.DaysOfMonth)
		for _, d := range rule.DaysOfMonth {
			if d > from.Day() {
				trialDate := time.Date(from.Year(), from.Month(), d, from.Hour(), from.Minute(), 0, 0, from.Location())
				if trialDate.Month() == from.Month() {
					return trialDate
				}
			}
		}
		return time.Date(from.Year(), from.Month()+1, rule.DaysOfMonth[0], from.Hour(), from.Minute(), 0, 0, from.Location())

	case "parity":
		next := from.AddDate(0, 0, 1)
		for i := 0; i < 366; i++ {
			isEven := next.Day()%2 == 0
			if rule.Parity != nil {
				if (*rule.Parity == "even" && isEven) || (*rule.Parity == "odd" && !isEven) {
					return next
				}
			}
			next = next.AddDate(0, 0, 1)
		}
		return next

	case "specific_dates":
		if len(rule.SpecificDates) == 0 {
			return from.AddDate(0, 0, 1)
		}
		for _, d := range rule.SpecificDates {
			if d.After(from) {
				return d
			}
		}
		return from 

	default:
		return from.AddDate(0, 0, 1)
	}
}

func (s *Service) validateRecurrence(r *taskdomain.RecurrenceRule) error {
	if r == nil {
		return nil
	}
	switch r.Type {
	case "daily":
		if r.EveryNDays != nil && *r.EveryNDays <= 0 {
			return fmt.Errorf("every_n_days must be positive")
		}
	case "monthly":
		if len(r.DaysOfMonth) == 0 {
			return fmt.Errorf("days_of_month is required for monthly recurrence")
		}
		for _, day := range r.DaysOfMonth {
			if day < 1 || day > 31 {
				return fmt.Errorf("invalid day of month: %d", day)
			}
		}
	case "parity":
		if r.Parity == nil || (*r.Parity != "even" && *r.Parity != "odd") {
			return fmt.Errorf("parity must be 'even' or 'odd'")
		}
	case "specific_dates":
		if len(r.SpecificDates) == 0 {
			return fmt.Errorf("specific_dates list cannot be empty")
		}
	}
	return nil
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) { return s.repo.List(ctx) }
func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) { return s.repo.GetByID(ctx, id) }
func (s *Service) Delete(ctx context.Context, id int64) error { return s.repo.Delete(ctx, id) }