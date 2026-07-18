package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetOrCreate resolve o usuário pelo username, criando-o se ainda não existir.
// É a base do login sem senha.
func (r *Repository) GetOrCreate(ctx context.Context, username string) (User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`SELECT id, username FROM users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return User{}, fmt.Errorf("select user: %w", err)
	}

	// INSERT ... ON CONFLICT cobre a corrida de dois logins simultâneos com o
	// mesmo username: o segundo cai no DO UPDATE e ainda retorna a linha.
	if err := r.pool.QueryRow(ctx,
		`INSERT INTO users (username) VALUES ($1)
		 ON CONFLICT (username) DO UPDATE SET username = EXCLUDED.username
		 RETURNING id, username`,
		username,
	).Scan(&u.ID, &u.Username); err != nil {
		return User{}, fmt.Errorf("insert user: %w", err)
	}
	return u, nil
}
