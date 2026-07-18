package users

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

var ErrInvalidUsername = errors.New("username must be 3-20 chars: letters, digits, _ or -")

// usernamePattern valida o login antes de tocar o DB. Guardamos sempre em
// minúsculo para a unicidade ser case-insensitive.
var usernamePattern = regexp.MustCompile(`^[a-z0-9_-]{3,20}$`)

func normalizeUsername(u string) (string, bool) {
	u = strings.ToLower(strings.TrimSpace(u))
	if !usernamePattern.MatchString(u) {
		return "", false
	}
	return u, true
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Login é um get-or-create por username (sem senha, por escolha de design).
func (s *Service) Login(ctx context.Context, username string) (User, error) {
	u, ok := normalizeUsername(username)
	if !ok {
		return User{}, ErrInvalidUsername
	}
	return s.repo.GetOrCreate(ctx, u)
}
