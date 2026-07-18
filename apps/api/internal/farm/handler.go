package farm

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
	r.Route("/{userID}", func(r chi.Router) {
		r.Get("/", h.getFarm)
		r.Patch("/", h.renameFarm)
		r.Post("/feed", h.feed)
		r.Post("/sell", h.sell)
		r.Post("/buy", h.buy)
		r.Post("/plant", h.plant)
		r.Post("/harvest", h.harvest)
	})
}

func writeError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		httpx.Error(w, http.StatusNotFound, "not found")
	case errors.Is(err, ErrInsufficientCoins), errors.Is(err, ErrInsufficientItems),
		errors.Is(err, ErrNoFreePlot), errors.Is(err, ErrCropNotReady), errors.Is(err, ErrNoCrop):
		httpx.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, ErrInvalidName), errors.Is(err, ErrEmptyName),
		errors.Is(err, ErrNotForSale), errors.Is(err, ErrUnknownAnimal),
		errors.Is(err, ErrUnknownSeed):
		httpx.Error(w, http.StatusBadRequest, err.Error())
	default:
		slog.Error(action, "error", err)
		httpx.Error(w, http.StatusInternalServerError, "internal error")
	}
}

// respondState devolve o estado atual (materializado) da fazenda — usado após
// mutações da economia para o cliente atualizar tudo (moedas incluídas) sem um
// segundo request.
func (h *Handler) respondState(w http.ResponseWriter, r *http.Request, userID, action string) {
	f, err := h.service.GetFarm(r.Context(), userID)
	if err != nil {
		writeError(w, err, action)
		return
	}
	httpx.JSON(w, http.StatusOK, f)
}

func (h *Handler) getFarm(w http.ResponseWriter, r *http.Request) {
	farm, err := h.service.GetFarm(r.Context(), chi.URLParam(r, "userID"))
	if err != nil {
		writeError(w, err, "get farm")
		return
	}
	httpx.JSON(w, http.StatusOK, farm)
}

type renameFarmRequest struct {
	Name string `json:"name"`
}

func (h *Handler) renameFarm(w http.ResponseWriter, r *http.Request) {
	var body renameFarmRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	userID := chi.URLParam(r, "userID")
	if err := h.service.Rename(r.Context(), userID, body.Name); err != nil {
		writeError(w, err, "rename farm")
		return
	}
	h.respondState(w, r, userID, "rename farm")
}

func (h *Handler) feed(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Feed(r.Context(), chi.URLParam(r, "userID")); err != nil {
		writeError(w, err, "feed farm")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type sellRequest struct {
	Item string `json:"item"`
	Qty  int    `json:"qty"`
}

func (h *Handler) sell(w http.ResponseWriter, r *http.Request) {
	var body sellRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	userID := chi.URLParam(r, "userID")
	if err := h.service.Sell(r.Context(), userID, body.Item, body.Qty); err != nil {
		writeError(w, err, "sell item")
		return
	}
	h.respondState(w, r, userID, "sell item")
}

type typeRequest struct {
	Type string `json:"type"`
}

func (h *Handler) buy(w http.ResponseWriter, r *http.Request) {
	var body typeRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	userID := chi.URLParam(r, "userID")
	if err := h.service.Buy(r.Context(), userID, body.Type); err != nil {
		writeError(w, err, "buy animal")
		return
	}
	h.respondState(w, r, userID, "buy animal")
}

func (h *Handler) plant(w http.ResponseWriter, r *http.Request) {
	var body typeRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	userID := chi.URLParam(r, "userID")
	if err := h.service.Plant(r.Context(), userID, body.Type); err != nil {
		writeError(w, err, "plant crop")
		return
	}
	h.respondState(w, r, userID, "plant crop")
}

type harvestRequest struct {
	PlotIndex int `json:"plotIndex"`
}

func (h *Handler) harvest(w http.ResponseWriter, r *http.Request) {
	var body harvestRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	userID := chi.URLParam(r, "userID")
	if err := h.service.Harvest(r.Context(), userID, body.PlotIndex); err != nil {
		writeError(w, err, "harvest crop")
		return
	}
	h.respondState(w, r, userID, "harvest crop")
}
