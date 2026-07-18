package farm

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/thiago/todo-simple-api/internal/ids"
)

var (
	ErrInvalidName = errors.New("name must be at most 40 characters")
	ErrEmptyName   = errors.New("name must not be empty")

	// Conflitos de estado da economia (mapeados para 409).
	ErrInsufficientCoins = errors.New("not enough coins")
	ErrInsufficientItems = errors.New("not enough items")
	ErrNoFreePlot        = errors.New("no free plot available")
	ErrCropNotReady      = errors.New("crop is not ready to harvest")
	ErrNoCrop            = errors.New("no crop at that plot")

	// Entradas inválidas da economia (mapeadas para 400).
	ErrNotForSale    = errors.New("item cannot be sold")
	ErrUnknownAnimal = errors.New("unknown animal")
	ErrUnknownSeed   = errors.New("unknown seed")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// EnsureFarm cria a fazenda inicial do usuário se ela ainda não existir. É
// idempotente e chamada no login (users.FarmEnsurer).
func (s *Service) EnsureFarm(ctx context.Context, userID string) error {
	if !ids.ValidUUID(userID) {
		return pgx.ErrNoRows
	}
	return s.repo.EnsureFarm(ctx, userID)
}

func (s *Service) GetFarm(ctx context.Context, userID string) (Farm, error) {
	if !ids.ValidUUID(userID) {
		return Farm{}, pgx.ErrNoRows
	}
	return s.repo.GetFarm(ctx, userID)
}

func (s *Service) Feed(ctx context.Context, userID string) error {
	if !ids.ValidUUID(userID) {
		return pgx.ErrNoRows
	}
	return s.repo.Feed(ctx, userID)
}

func (s *Service) Rename(ctx context.Context, userID, name string) error {
	if !ids.ValidUUID(userID) {
		return pgx.ErrNoRows
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrEmptyName
	}
	if len(name) > 40 {
		return ErrInvalidName
	}
	return s.repo.Rename(ctx, userID, name)
}

func (s *Service) Sell(ctx context.Context, userID, item string, qty int) error {
	if !ids.ValidUUID(userID) {
		return pgx.ErrNoRows
	}
	if _, sellable := sellPrice(item); !sellable {
		return ErrNotForSale
	}
	if qty <= 0 {
		return ErrInsufficientItems
	}
	return s.repo.Sell(ctx, userID, item, qty)
}

func (s *Service) Buy(ctx context.Context, userID, animalType string) error {
	if !ids.ValidUUID(userID) {
		return pgx.ErrNoRows
	}
	if _, known := animalPrice(animalType); !known {
		return ErrUnknownAnimal
	}
	return s.repo.Buy(ctx, userID, animalType)
}

func (s *Service) Plant(ctx context.Context, userID, cropType string) error {
	if !ids.ValidUUID(userID) {
		return pgx.ErrNoRows
	}
	if _, known := seedPrice(cropType); !known {
		return ErrUnknownSeed
	}
	return s.repo.Plant(ctx, userID, cropType)
}

func (s *Service) Harvest(ctx context.Context, userID string, plotIndex int) error {
	if !ids.ValidUUID(userID) {
		return pgx.ErrNoRows
	}
	return s.repo.Harvest(ctx, userID, plotIndex)
}
