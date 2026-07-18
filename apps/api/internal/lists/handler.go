package lists

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/thiago/todo-simple-api/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes(r chi.Router) {
	r.Post("/", h.createList)
	r.Post("/claim", h.claimLists)
	r.Get("/by-user/{userID}", h.listsByUser)
	r.Route("/{slug}", func(r chi.Router) {
		r.Get("/", h.getList)
		r.Patch("/", h.updateList)
		r.Delete("/", h.deleteList)
		r.Post("/tasks", h.addTask)
		r.Put("/tasks/order", h.reorderTasks)
		r.Patch("/tasks/{taskID}", h.updateTask)
		r.Delete("/tasks/{taskID}", h.deleteTask)
	})
}

// writeError maps domain errors to HTTP responses; action is used for logging
// unexpected failures.
func writeError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		httpx.Error(w, http.StatusNotFound, "not found")
	case errors.Is(err, ErrInvalidName), errors.Is(err, ErrInvalidTitle),
		errors.Is(err, ErrInvalidColor), errors.Is(err, ErrEmptyUpdate),
		errors.Is(err, ErrInvalidUser):
		httpx.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrStaleOrder):
		httpx.Error(w, http.StatusConflict, err.Error())
	default:
		slog.Error(action, "error", err)
		httpx.Error(w, http.StatusInternalServerError, "internal error")
	}
}

type createListRequest struct {
	Name   string  `json:"name"`
	Color  string  `json:"color"`
	UserID *string `json:"userId"`
}

func (h *Handler) createList(w http.ResponseWriter, r *http.Request) {
	var body createListRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	list, err := h.service.CreateList(r.Context(), body.Name, body.Color, body.UserID)
	if err != nil {
		writeError(w, err, "create list")
		return
	}
	httpx.JSON(w, http.StatusCreated, list)
}

func (h *Handler) listsByUser(w http.ResponseWriter, r *http.Request) {
	lists, err := h.service.ListsByUser(r.Context(), chi.URLParam(r, "userID"))
	if err != nil {
		writeError(w, err, "lists by user")
		return
	}
	httpx.JSON(w, http.StatusOK, lists)
}

type claimListsRequest struct {
	UserID string   `json:"userId"`
	Slugs  []string `json:"slugs"`
}

func (h *Handler) claimLists(w http.ResponseWriter, r *http.Request) {
	var body claimListsRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	claimed, err := h.service.ClaimLists(r.Context(), body.UserID, body.Slugs)
	if err != nil {
		writeError(w, err, "claim lists")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]int64{"claimed": claimed})
}

func (h *Handler) getList(w http.ResponseWriter, r *http.Request) {
	list, err := h.service.GetList(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		writeError(w, err, "get list")
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

type updateListRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

func (h *Handler) updateList(w http.ResponseWriter, r *http.Request) {
	var body updateListRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	list, err := h.service.UpdateList(r.Context(), chi.URLParam(r, "slug"), body.Name, body.Color)
	if err != nil {
		writeError(w, err, "update list")
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

func (h *Handler) deleteList(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteList(r.Context(), chi.URLParam(r, "slug")); err != nil {
		writeError(w, err, "delete list")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type addTaskRequest struct {
	Title string `json:"title"`
}

func (h *Handler) addTask(w http.ResponseWriter, r *http.Request) {
	var body addTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	task, err := h.service.AddTask(r.Context(), chi.URLParam(r, "slug"), body.Title)
	if err != nil {
		writeError(w, err, "add task")
		return
	}
	httpx.JSON(w, http.StatusCreated, task)
}

type updateTaskRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request) {
	var body updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	task, err := h.service.UpdateTask(
		r.Context(),
		chi.URLParam(r, "slug"),
		chi.URLParam(r, "taskID"),
		body.Title,
		body.Done,
	)
	if err != nil {
		writeError(w, err, "update task")
		return
	}
	httpx.JSON(w, http.StatusOK, task)
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request) {
	err := h.service.DeleteTask(r.Context(), chi.URLParam(r, "slug"), chi.URLParam(r, "taskID"))
	if err != nil {
		writeError(w, err, "delete task")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type reorderTasksRequest struct {
	TaskIDs []string `json:"taskIds"`
}

func (h *Handler) reorderTasks(w http.ResponseWriter, r *http.Request) {
	var body reorderTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := h.service.ReorderTasks(r.Context(), chi.URLParam(r, "slug"), body.TaskIDs); err != nil {
		writeError(w, err, "reorder tasks")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
