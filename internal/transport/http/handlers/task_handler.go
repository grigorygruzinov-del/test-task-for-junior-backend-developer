package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	taskdomain "github.com/medods/test-task-for-junior-backend-developer/internal/domain/task"
)

// Usecase описывает интерфейс взаимодействия с бизнес-логикой
type Usecase interface {
	CreateTask(ctx context.Context, t taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, t taskdomain.Task) (*taskdomain.Task, error)
	List(ctx context.Context) ([]taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
}

type TaskHandler struct {
	usecase Usecase
}

func NewTaskHandler(u Usecase) *TaskHandler {
	return &TaskHandler{usecase: u}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t := taskdomain.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      taskdomain.Status(req.Status),
		ScheduledAt: req.ScheduledAt,
	}

	if req.Recurrence != nil {
		t.Recurrence = &taskdomain.RecurrenceRule{
			Type:          req.Recurrence.Type,
			EveryNDays:    req.Recurrence.EveryNDays,
			DaysOfMonth:   req.Recurrence.DaysOfMonth,
			SpecificDates: req.Recurrence.SpecificDates,
			Parity:        req.Recurrence.Parity,
		}
	}

	created, err := h.usecase.CreateTask(r.Context(), t)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newTaskDTO{ID: fmt.Sprintf("%d", created.ID)})
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.usecase.List(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	if tasks == nil {
		tasks = []taskdomain.Task{}
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id format")
		return
	}

	t, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id format")
		return
	}

	var req taskMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t := taskdomain.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      taskdomain.Status(req.Status),
		ScheduledAt: req.ScheduledAt,
	}

	if req.Recurrence != nil {
		t.Recurrence = &taskdomain.RecurrenceRule{
			Type:          req.Recurrence.Type,
			EveryNDays:    req.Recurrence.EveryNDays,
			DaysOfMonth:   req.Recurrence.DaysOfMonth,
			SpecificDates: req.Recurrence.SpecificDates,
			Parity:        req.Recurrence.Parity,
		}
	}

	updated, err := h.usecase.Update(r.Context(), id, t)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id format")
		return
	}

	if err := h.usecase.Delete(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Хелперы ---

func decodeJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	if err.Error() == "task not found" {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}