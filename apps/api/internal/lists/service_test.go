package lists

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// As validações falham antes de tocar o repositório, então um repo nil basta.
func newTestService() *Service {
	return NewService(nil)
}

const testSlug = "abcdefghijklmnopqrstu"

func TestCreateListValidation(t *testing.T) {
	s := newTestService()
	ctx := context.Background()

	if _, err := s.CreateList(ctx, "   ", ""); !errors.Is(err, ErrInvalidName) {
		t.Errorf("blank name: got %v, want ErrInvalidName", err)
	}
	if _, err := s.CreateList(ctx, strings.Repeat("a", 121), ""); !errors.Is(err, ErrInvalidName) {
		t.Errorf("oversized name: got %v, want ErrInvalidName", err)
	}
	if _, err := s.CreateList(ctx, "Compras", "magenta"); !errors.Is(err, ErrInvalidColor) {
		t.Errorf("unknown color: got %v, want ErrInvalidColor", err)
	}
}

func TestSlugValidation(t *testing.T) {
	s := newTestService()
	ctx := context.Background()

	for _, slug := range []string{"", "short", strings.Repeat("a", 22), "invalid slug with spa"} {
		if _, err := s.GetList(ctx, slug); !errors.Is(err, pgx.ErrNoRows) {
			t.Errorf("GetList(%q): got %v, want pgx.ErrNoRows", slug, err)
		}
	}
}

func TestAddTaskValidation(t *testing.T) {
	s := newTestService()
	ctx := context.Background()

	if _, err := s.AddTask(ctx, testSlug, "  "); !errors.Is(err, ErrInvalidTitle) {
		t.Errorf("blank title: got %v, want ErrInvalidTitle", err)
	}
	if _, err := s.AddTask(ctx, testSlug, strings.Repeat("a", 501)); !errors.Is(err, ErrInvalidTitle) {
		t.Errorf("oversized title: got %v, want ErrInvalidTitle", err)
	}
	if _, err := s.AddTask(ctx, "bad-slug", "Tarefa"); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("bad slug: got %v, want pgx.ErrNoRows", err)
	}
}

func TestUpdateTaskValidation(t *testing.T) {
	s := newTestService()
	ctx := context.Background()
	title := "Tarefa"
	uuid := "5f7c9c1e-0a4b-4e6a-9d3f-2b8c1a7e5d90"

	if _, err := s.UpdateTask(ctx, testSlug, "not-a-uuid", &title, nil); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("bad task id: got %v, want pgx.ErrNoRows", err)
	}
	if _, err := s.UpdateTask(ctx, testSlug, uuid, nil, nil); !errors.Is(err, ErrEmptyUpdate) {
		t.Errorf("empty update: got %v, want ErrEmptyUpdate", err)
	}
}

func TestReorderTasksValidation(t *testing.T) {
	s := newTestService()
	ctx := context.Background()

	if err := s.ReorderTasks(ctx, testSlug, []string{"not-a-uuid"}); !errors.Is(err, ErrStaleOrder) {
		t.Errorf("bad task id: got %v, want ErrStaleOrder", err)
	}
}
