package lists

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrStaleOrder means a reorder payload is not an exact permutation of the
// list's current task IDs — the client must refetch and retry.
var ErrStaleOrder = errors.New("task order is stale")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateList(ctx context.Context, slug, name, color string) (List, error) {
	var l List
	row := r.pool.QueryRow(
		ctx,
		`INSERT INTO lists (slug, name, color) VALUES ($1, $2, $3) RETURNING slug, name, color, created_at`,
		slug, name, color,
	)
	if err := row.Scan(&l.Slug, &l.Name, &l.Color, &l.CreatedAt); err != nil {
		return List{}, fmt.Errorf("insert list: %w", err)
	}
	return l, nil
}

func (r *Repository) GetList(ctx context.Context, slug string) (ListWithTasks, error) {
	var (
		lt     ListWithTasks
		listID string
	)
	row := r.pool.QueryRow(
		ctx,
		`SELECT id, slug, name, color, created_at FROM lists WHERE slug = $1`,
		slug,
	)
	if err := row.Scan(&listID, &lt.Slug, &lt.Name, &lt.Color, &lt.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ListWithTasks{}, pgx.ErrNoRows
		}
		return ListWithTasks{}, fmt.Errorf("query list: %w", err)
	}

	rows, err := r.pool.Query(
		ctx,
		`SELECT id, title, done, position, created_at, completed_at
		 FROM tasks WHERE list_id = $1 ORDER BY position ASC, created_at ASC`,
		listID,
	)
	if err != nil {
		return ListWithTasks{}, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	lt.Tasks = make([]Task, 0)
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.Position, &t.CreatedAt, &t.CompletedAt); err != nil {
			return ListWithTasks{}, fmt.Errorf("scan task: %w", err)
		}
		lt.Tasks = append(lt.Tasks, t)
	}
	if err := rows.Err(); err != nil {
		return ListWithTasks{}, fmt.Errorf("iterate tasks: %w", err)
	}
	return lt, nil
}

func (r *Repository) UpdateList(ctx context.Context, slug string, name, color *string) (List, error) {
	var l List
	row := r.pool.QueryRow(
		ctx,
		`UPDATE lists
		 SET name = COALESCE($2, name), color = COALESCE($3, color), updated_at = now()
		 WHERE slug = $1
		 RETURNING slug, name, color, created_at`,
		slug, name, color,
	)
	if err := row.Scan(&l.Slug, &l.Name, &l.Color, &l.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return List{}, pgx.ErrNoRows
		}
		return List{}, fmt.Errorf("update list: %w", err)
	}
	return l, nil
}

func (r *Repository) DeleteList(ctx context.Context, slug string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM lists WHERE slug = $1`, slug)
	if err != nil {
		return fmt.Errorf("delete list: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) AddTask(ctx context.Context, slug, title string) (Task, error) {
	var t Task
	row := r.pool.QueryRow(
		ctx,
		`INSERT INTO tasks (list_id, title, position)
		 SELECT l.id, $2, COALESCE((SELECT MAX(position) FROM tasks WHERE list_id = l.id), 0) + 1
		 FROM lists l WHERE l.slug = $1
		 RETURNING id, title, done, position, created_at, completed_at`,
		slug, title,
	)
	if err := row.Scan(&t.ID, &t.Title, &t.Done, &t.Position, &t.CreatedAt, &t.CompletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Task{}, pgx.ErrNoRows
		}
		return Task{}, fmt.Errorf("insert task: %w", err)
	}
	return t, nil
}

func (r *Repository) UpdateTask(ctx context.Context, slug, taskID string, title *string, done *bool) (Task, error) {
	var t Task
	row := r.pool.QueryRow(
		ctx,
		`UPDATE tasks
		 SET title = COALESCE($3, title),
		     done = COALESCE($4, done),
		     completed_at = CASE
		         WHEN $4::boolean IS NULL THEN completed_at
		         WHEN $4 AND done THEN completed_at
		         WHEN $4 THEN now()
		         ELSE NULL
		     END,
		     updated_at = now()
		 WHERE id = $2 AND list_id = (SELECT id FROM lists WHERE slug = $1)
		 RETURNING id, title, done, position, created_at, completed_at`,
		slug, taskID, title, done,
	)
	if err := row.Scan(&t.ID, &t.Title, &t.Done, &t.Position, &t.CreatedAt, &t.CompletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Task{}, pgx.ErrNoRows
		}
		return Task{}, fmt.Errorf("update task: %w", err)
	}
	return t, nil
}

func (r *Repository) DeleteTask(ctx context.Context, slug, taskID string) error {
	tag, err := r.pool.Exec(
		ctx,
		`DELETE FROM tasks WHERE id = $2 AND list_id = (SELECT id FROM lists WHERE slug = $1)`,
		slug, taskID,
	)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ReorderTasks renumbers the list's tasks 1..n following taskIDs, which must
// be an exact permutation of the list's current task IDs.
func (r *Repository) ReorderTasks(ctx context.Context, slug string, taskIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin reorder: %w", err)
	}
	defer tx.Rollback(ctx)

	var listID string
	if err := tx.QueryRow(ctx, `SELECT id FROM lists WHERE slug = $1`, slug).Scan(&listID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("query list for reorder: %w", err)
	}

	rows, err := tx.Query(ctx, `SELECT id FROM tasks WHERE list_id = $1 FOR UPDATE`, listID)
	if err != nil {
		return fmt.Errorf("lock tasks: %w", err)
	}
	current := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return fmt.Errorf("scan task id: %w", err)
		}
		current[id] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate task ids: %w", err)
	}

	if len(taskIDs) != len(current) {
		return ErrStaleOrder
	}
	seen := make(map[string]bool, len(taskIDs))
	for _, id := range taskIDs {
		if !current[id] || seen[id] {
			return ErrStaleOrder
		}
		seen[id] = true
	}

	if _, err := tx.Exec(
		ctx,
		`UPDATE tasks SET position = u.ord, updated_at = now()
		 FROM unnest($2::uuid[]) WITH ORDINALITY AS u(id, ord)
		 WHERE tasks.id = u.id AND tasks.list_id = $1`,
		listID, taskIDs,
	); err != nil {
		return fmt.Errorf("renumber tasks: %w", err)
	}

	return tx.Commit(ctx)
}
