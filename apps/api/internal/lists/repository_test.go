package lists

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/thiago/todo-simple-api/internal/database"
	"github.com/thiago/todo-simple-api/internal/ids"
)

// Teste de integração: roda apenas com DATABASE_URL definido (CI ou dev local).
func newTestRepository(t *testing.T) *Repository {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping integration test")
	}
	ctx := context.Background()
	if err := database.Migrate(dsn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return NewRepository(pool)
}

func createTestList(t *testing.T, repo *Repository) List {
	t.Helper()
	slug, err := ids.NewSlug()
	if err != nil {
		t.Fatalf("slug: %v", err)
	}
	list, err := repo.CreateList(context.Background(), slug, "Lista de teste", DefaultColor, nil)
	if err != nil {
		t.Fatalf("create list: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.DeleteList(context.Background(), list.Slug)
	})
	return list
}

func TestReorderTasks(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()
	list := createTestList(t, repo)

	var taskIDs []string
	for _, title := range []string{"primeira", "segunda", "terceira"} {
		task, err := repo.AddTask(ctx, list.Slug, title)
		if err != nil {
			t.Fatalf("add task %q: %v", title, err)
		}
		taskIDs = append(taskIDs, task.ID)
	}

	// inverte a ordem
	reversed := []string{taskIDs[2], taskIDs[1], taskIDs[0]}
	if err := repo.ReorderTasks(ctx, list.Slug, reversed); err != nil {
		t.Fatalf("reorder: %v", err)
	}

	got, err := repo.GetList(ctx, list.Slug)
	if err != nil {
		t.Fatalf("get list: %v", err)
	}
	if len(got.Tasks) != 3 {
		t.Fatalf("got %d tasks, want 3", len(got.Tasks))
	}
	for i, task := range got.Tasks {
		if task.ID != reversed[i] {
			t.Errorf("position %d: got task %s, want %s", i, task.ID, reversed[i])
		}
		if task.Position != i+1 {
			t.Errorf("task %s: got position %d, want %d", task.ID, task.Position, i+1)
		}
	}
}

func TestReorderTasksStale(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()
	list := createTestList(t, repo)

	task, err := repo.AddTask(ctx, list.Slug, "única")
	if err != nil {
		t.Fatalf("add task: %v", err)
	}

	// ID que não pertence à lista
	foreign := "00000000-0000-4000-8000-000000000000"
	if err := repo.ReorderTasks(ctx, list.Slug, []string{foreign}); !errors.Is(err, ErrStaleOrder) {
		t.Errorf("foreign id: got %v, want ErrStaleOrder", err)
	}
	// quantidade errada
	if err := repo.ReorderTasks(ctx, list.Slug, []string{task.ID, task.ID}); !errors.Is(err, ErrStaleOrder) {
		t.Errorf("duplicated id: got %v, want ErrStaleOrder", err)
	}
}

func TestToggleTaskCompletedAt(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()
	list := createTestList(t, repo)

	task, err := repo.AddTask(ctx, list.Slug, "tarefa")
	if err != nil {
		t.Fatalf("add task: %v", err)
	}
	if task.CompletedAt != nil {
		t.Fatal("new task should not have completedAt")
	}

	done := true
	updated, err := repo.UpdateTask(ctx, list.Slug, task.ID, nil, &done)
	if err != nil {
		t.Fatalf("mark done: %v", err)
	}
	if !updated.Done || updated.CompletedAt == nil {
		t.Errorf("after done: done=%v completedAt=%v", updated.Done, updated.CompletedAt)
	}

	done = false
	updated, err = repo.UpdateTask(ctx, list.Slug, task.ID, nil, &done)
	if err != nil {
		t.Fatalf("mark undone: %v", err)
	}
	if updated.Done || updated.CompletedAt != nil {
		t.Errorf("after undone: done=%v completedAt=%v", updated.Done, updated.CompletedAt)
	}
}
