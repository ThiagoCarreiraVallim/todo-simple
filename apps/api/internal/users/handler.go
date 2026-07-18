package users

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/thiago/todo-simple-api/internal/httpx"
)

// FarmEnsurer cria a fazenda do usuário no login, se ainda não existir. Fica
// como interface para o domínio users não depender do pacote farm.
type FarmEnsurer interface {
	EnsureFarm(ctx context.Context, userID string) error
}

type Handler struct {
	service *Service
	farms   FarmEnsurer
}

func NewHandler(service *Service, farms FarmEnsurer) *Handler {
	return &Handler{service: service, farms: farms}
}

type loginRequest struct {
	Username string `json:"username"`
}

// Login: get-or-create do usuário e garante a fazenda. Devolve {id, username}.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	user, err := h.service.Login(r.Context(), body.Username)
	if err != nil {
		if errors.Is(err, ErrInvalidUsername) {
			httpx.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Error("login", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if err := h.farms.EnsureFarm(r.Context(), user.ID); err != nil {
		slog.Error("ensure farm on login", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}
