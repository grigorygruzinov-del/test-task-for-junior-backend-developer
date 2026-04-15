package transporthttp

import (
	"net/http"

	"github.com/gorilla/mux"

	// Используем правильные псевдонимы для импортов
	swaggerdocs "github.com/medods/test-task-for-junior-backend-developer/internal/transport/http/docs"
	httphandlers "github.com/medods/test-task-for-junior-backend-developer/internal/transport/http/handlers"
)

func NewRouter(taskHandler *httphandlers.TaskHandler, docsHandler *swaggerdocs.Handler) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// 1. Swagger эндпоинты
	router.HandleFunc("/swagger/openapi.json", docsHandler.ServeSpec).Methods(http.MethodGet)
	router.HandleFunc("/swagger/", docsHandler.ServeUI).Methods(http.MethodGet)
	router.HandleFunc("/swagger", docsHandler.RedirectToUI).Methods(http.MethodGet)

	// 2. API эндпоинты под префиксом v1
	api := router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/tasks", taskHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/tasks", taskHandler.List).Methods(http.MethodGet)
	
	// Используем простой формат {id}, чтобы твой хендлер (strconv.ParseInt) его точно поймал
	api.HandleFunc("/tasks/{id}", taskHandler.GetByID).Methods(http.MethodGet)
	api.HandleFunc("/tasks/{id}", taskHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/tasks/{id}", taskHandler.Delete).Methods(http.MethodDelete)

	return router
}