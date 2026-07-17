package lists

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/thiago/todo-simple-api/internal/ids"
)

var (
	ErrInvalidName  = errors.New("name must be 1-120 characters")
	ErrInvalidTitle = errors.New("title must be 1-500 characters")
	ErrInvalidColor = errors.New("color must be one of the palette tokens")
	ErrEmptyUpdate  = errors.New("nothing to update")
)

// Colors is the fixed palette; the web app maps each token to its styles.
var Colors = []string{"zinc", "red", "orange", "amber", "green", "teal", "blue", "violet", "pink"}

const DefaultColor = "zinc"

var (
	slugPattern = regexp.MustCompile(fmt.Sprintf(`^[A-Za-z0-9_-]{%d}$`, ids.SlugLength))
	uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func validColor(color string) bool {
	for _, c := range Colors {
		if c == color {
			return true
		}
	}
	return false
}

func validName(name string) bool {
	return name != "" && len(name) <= 120
}

func validTitle(title string) bool {
	return title != "" && len(title) <= 500
}

// validSlug pre-checks the capability slug shape so garbage never hits the DB.
func validSlug(slug string) bool {
	return slugPattern.MatchString(slug)
}

func (s *Service) CreateList(ctx context.Context, name, color string) (List, error) {
	name = strings.TrimSpace(name)
	if !validName(name) {
		return List{}, ErrInvalidName
	}
	if color == "" {
		color = DefaultColor
	}
	if !validColor(color) {
		return List{}, ErrInvalidColor
	}
	slug, err := ids.NewSlug()
	if err != nil {
		return List{}, fmt.Errorf("generate slug: %w", err)
	}
	return s.repo.CreateList(ctx, slug, name, color)
}

func (s *Service) GetList(ctx context.Context, slug string) (ListWithTasks, error) {
	if !validSlug(slug) {
		return ListWithTasks{}, pgx.ErrNoRows
	}
	return s.repo.GetList(ctx, slug)
}

func (s *Service) UpdateList(ctx context.Context, slug string, name, color *string) (List, error) {
	if !validSlug(slug) {
		return List{}, pgx.ErrNoRows
	}
	if name == nil && color == nil {
		return List{}, ErrEmptyUpdate
	}
	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if !validName(trimmed) {
			return List{}, ErrInvalidName
		}
		name = &trimmed
	}
	if color != nil && !validColor(*color) {
		return List{}, ErrInvalidColor
	}
	return s.repo.UpdateList(ctx, slug, name, color)
}

func (s *Service) DeleteList(ctx context.Context, slug string) error {
	if !validSlug(slug) {
		return pgx.ErrNoRows
	}
	return s.repo.DeleteList(ctx, slug)
}

func (s *Service) AddTask(ctx context.Context, slug, title string) (Task, error) {
	if !validSlug(slug) {
		return Task{}, pgx.ErrNoRows
	}
	title = strings.TrimSpace(title)
	if !validTitle(title) {
		return Task{}, ErrInvalidTitle
	}
	return s.repo.AddTask(ctx, slug, title)
}

func (s *Service) UpdateTask(ctx context.Context, slug, taskID string, title *string, done *bool) (Task, error) {
	if !validSlug(slug) || !uuidPattern.MatchString(taskID) {
		return Task{}, pgx.ErrNoRows
	}
	if title == nil && done == nil {
		return Task{}, ErrEmptyUpdate
	}
	if title != nil {
		trimmed := strings.TrimSpace(*title)
		if !validTitle(trimmed) {
			return Task{}, ErrInvalidTitle
		}
		title = &trimmed
	}
	return s.repo.UpdateTask(ctx, slug, taskID, title, done)
}

func (s *Service) DeleteTask(ctx context.Context, slug, taskID string) error {
	if !validSlug(slug) || !uuidPattern.MatchString(taskID) {
		return pgx.ErrNoRows
	}
	return s.repo.DeleteTask(ctx, slug, taskID)
}

func (s *Service) ReorderTasks(ctx context.Context, slug string, taskIDs []string) error {
	if !validSlug(slug) {
		return pgx.ErrNoRows
	}
	for _, id := range taskIDs {
		if !uuidPattern.MatchString(id) {
			return ErrStaleOrder
		}
	}
	return s.repo.ReorderTasks(ctx, slug, taskIDs)
}
