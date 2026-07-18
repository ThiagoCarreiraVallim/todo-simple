package farm

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// EnsureFarm cria a fazenda do usuário (com 1 galinha e 1 milho) se ainda não
// existir. Idempotente — chamada no login. A corrida de dois logins simultâneos
// é coberta pelo ON CONFLICT no user_id único.
func (r *Repository) EnsureFarm(ctx context.Context, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin ensure farm: %w", err)
	}
	defer tx.Rollback(ctx)

	var farmID string
	err = tx.QueryRow(ctx,
		`INSERT INTO farms (user_id) VALUES ($1)
		 ON CONFLICT (user_id) DO NOTHING
		 RETURNING id`,
		userID,
	).Scan(&farmID)
	if errors.Is(err, pgx.ErrNoRows) {
		// Já existia (ON CONFLICT DO NOTHING não retorna linha): nada a semear.
		return tx.Commit(ctx)
	}
	if err != nil {
		return fmt.Errorf("insert farm: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO animals (farm_id, type) VALUES ($1, 'chicken')`, farmID,
	); err != nil {
		return fmt.Errorf("insert chicken: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO crops (farm_id, plot_index, type) VALUES ($1, 0, 'corn')`, farmID,
	); err != nil {
		return fmt.Errorf("insert crop: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit ensure farm: %w", err)
	}
	return nil
}

func (r *Repository) Rename(ctx context.Context, userID, name string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE farms SET name = $2, updated_at = now() WHERE user_id = $1`, userID, name)
	if err != nil {
		return fmt.Errorf("rename farm: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// farmID busca o id interno da fazenda pelo user_id, bloqueando a linha
// (FOR UPDATE) para serializar operações concorrentes de economia.
func farmID(ctx context.Context, tx pgx.Tx, userID string) (string, error) {
	var id string
	if err := tx.QueryRow(ctx,
		`SELECT id FROM farms WHERE user_id = $1 FOR UPDATE`, userID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", pgx.ErrNoRows
		}
		return "", fmt.Errorf("query farm: %w", err)
	}
	return id, nil
}

// Feed dá ração a todos os animais da fazenda: sobe o conforto e reinicia o
// relógio de decaimento. A fórmula vive no engine (feedComfort).
func (r *Repository) Feed(ctx context.Context, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin feed: %w", err)
	}
	defer tx.Rollback(ctx)

	id, err := farmID(ctx, tx, userID)
	if err != nil {
		return err
	}

	type animalRow struct {
		id          string
		comfortBase float64
		lastFedAt   time.Time
	}
	rows, err := tx.Query(ctx,
		`SELECT id, comfort_base, last_fed_at FROM animals WHERE farm_id = $1 FOR UPDATE`, id)
	if err != nil {
		return fmt.Errorf("query animals for feed: %w", err)
	}
	var animals []animalRow
	for rows.Next() {
		var a animalRow
		if err := rows.Scan(&a.id, &a.comfortBase, &a.lastFedAt); err != nil {
			rows.Close()
			return fmt.Errorf("scan animal: %w", err)
		}
		animals = append(animals, a)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate animals: %w", err)
	}

	now := time.Now()
	for _, a := range animals {
		newBase := feedComfort(a.comfortBase, a.lastFedAt, now)
		if _, err := tx.Exec(ctx,
			`UPDATE animals SET comfort_base = $2, last_fed_at = $3 WHERE id = $1`,
			a.id, newBase, now,
		); err != nil {
			return fmt.Errorf("update animal comfort: %w", err)
		}
	}
	return tx.Commit(ctx)
}

// Sell vende `qty` unidades de `item` do inventário por moedas (preço do
// sellCatalog). Falha se o inventário não tiver o suficiente.
func (r *Repository) Sell(ctx context.Context, userID, item string, qty int) error {
	price, ok := sellPrice(item)
	if !ok {
		return ErrNotForSale
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin sell: %w", err)
	}
	defer tx.Rollback(ctx)

	id, err := farmID(ctx, tx, userID)
	if err != nil {
		return err
	}

	var have int
	err = tx.QueryRow(ctx,
		`SELECT qty FROM farm_items WHERE farm_id = $1 AND item = $2 FOR UPDATE`, id, item).Scan(&have)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("query inventory: %w", err)
	}
	if have < qty {
		return ErrInsufficientItems
	}
	if _, err := tx.Exec(ctx,
		`UPDATE farm_items SET qty = qty - $3 WHERE farm_id = $1 AND item = $2`, id, item, qty,
	); err != nil {
		return fmt.Errorf("deduct items: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`UPDATE farms SET coins = coins + $2, updated_at = now() WHERE id = $1`, id, price*qty,
	); err != nil {
		return fmt.Errorf("credit coins: %w", err)
	}
	return tx.Commit(ctx)
}

// Buy compra um animal do animalCatalog, debitando moedas.
func (r *Repository) Buy(ctx context.Context, userID, animalType string) error {
	price, ok := animalPrice(animalType)
	if !ok {
		return ErrUnknownAnimal
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin buy: %w", err)
	}
	defer tx.Rollback(ctx)

	var (
		id    string
		coins int
	)
	if err := tx.QueryRow(ctx,
		`SELECT id, coins FROM farms WHERE user_id = $1 FOR UPDATE`, userID).Scan(&id, &coins); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("query farm: %w", err)
	}
	if coins < price {
		return ErrInsufficientCoins
	}
	if _, err := tx.Exec(ctx,
		`UPDATE farms SET coins = coins - $2, updated_at = now() WHERE id = $1`, id, price,
	); err != nil {
		return fmt.Errorf("debit coins: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO animals (farm_id, type) VALUES ($1, $2)`, id, animalType,
	); err != nil {
		return fmt.Errorf("insert animal: %w", err)
	}
	return tx.Commit(ctx)
}

// Plant planta uma semente no próximo canteiro livre (0..maxPlots-1),
// debitando moedas. Falha se não houver canteiro livre ou moedas.
func (r *Repository) Plant(ctx context.Context, userID, cropType string) error {
	price, ok := seedPrice(cropType)
	if !ok {
		return ErrUnknownSeed
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin plant: %w", err)
	}
	defer tx.Rollback(ctx)

	var (
		id    string
		coins int
	)
	if err := tx.QueryRow(ctx,
		`SELECT id, coins FROM farms WHERE user_id = $1 FOR UPDATE`, userID).Scan(&id, &coins); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("query farm: %w", err)
	}

	used := map[int]bool{}
	prows, err := tx.Query(ctx, `SELECT plot_index FROM crops WHERE farm_id = $1`, id)
	if err != nil {
		return fmt.Errorf("query plots: %w", err)
	}
	for prows.Next() {
		var idx int
		if err := prows.Scan(&idx); err != nil {
			prows.Close()
			return fmt.Errorf("scan plot: %w", err)
		}
		used[idx] = true
	}
	prows.Close()
	if err := prows.Err(); err != nil {
		return fmt.Errorf("iterate plots: %w", err)
	}

	plot := -1
	for i := 0; i < maxPlots; i++ {
		if !used[i] {
			plot = i
			break
		}
	}
	if plot == -1 {
		return ErrNoFreePlot
	}
	if coins < price {
		return ErrInsufficientCoins
	}
	if _, err := tx.Exec(ctx,
		`UPDATE farms SET coins = coins - $2, updated_at = now() WHERE id = $1`, id, price,
	); err != nil {
		return fmt.Errorf("debit coins: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO crops (farm_id, plot_index, type) VALUES ($1, $2, $3)`, id, plot, cropType,
	); err != nil {
		return fmt.Errorf("insert crop: %w", err)
	}
	return tx.Commit(ctx)
}

// Harvest colhe um canteiro pronto: credita o rendimento no inventário e libera
// o canteiro. Falha se não houver cultura ali ou se ainda não estiver madura.
func (r *Repository) Harvest(ctx context.Context, userID string, plotIndex int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin harvest: %w", err)
	}
	defer tx.Rollback(ctx)

	id, err := farmID(ctx, tx, userID)
	if err != nil {
		return err
	}

	var (
		cropType  string
		plantedAt time.Time
	)
	err = tx.QueryRow(ctx,
		`SELECT type, planted_at FROM crops WHERE farm_id = $1 AND plot_index = $2 FOR UPDATE`,
		id, plotIndex).Scan(&cropType, &plantedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoCrop
		}
		return fmt.Errorf("query crop: %w", err)
	}
	if !cropReady(cropType, plantedAt, time.Now()) {
		return ErrCropNotReady
	}

	yield := cropYield[cropType]
	if yield > 0 {
		if _, err := tx.Exec(ctx,
			`INSERT INTO farm_items (farm_id, item, qty) VALUES ($1, $2, $3)
			 ON CONFLICT (farm_id, item) DO UPDATE SET qty = farm_items.qty + $3`,
			id, cropType, yield,
		); err != nil {
			return fmt.Errorf("credit harvest: %w", err)
		}
	}
	if _, err := tx.Exec(ctx,
		`DELETE FROM crops WHERE farm_id = $1 AND plot_index = $2`, id, plotIndex,
	); err != nil {
		return fmt.Errorf("clear plot: %w", err)
	}
	return tx.Commit(ctx)
}

// GetFarm lê o estado materializando o tempo idle: credita os produtos (ovos,
// leite) gerados desde a última leitura e calcula conforto/estágios atuais.
func (r *Repository) GetFarm(ctx context.Context, userID string) (Farm, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Farm{}, fmt.Errorf("begin get farm: %w", err)
	}
	defer tx.Rollback(ctx)

	var (
		id string
		f  Farm
	)
	if err := tx.QueryRow(ctx,
		`SELECT id, name, coins, created_at FROM farms WHERE user_id = $1 FOR UPDATE`, userID,
	).Scan(&id, &f.Name, &f.Coins, &f.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Farm{}, pgx.ErrNoRows
		}
		return Farm{}, fmt.Errorf("query farm: %w", err)
	}

	now := time.Now()

	// Animais + materialização de produtos (ovos/leite).
	type animalRow struct {
		id          string
		typ         string
		comfortBase float64
		lastFedAt   time.Time
		lastEggAt   time.Time
	}
	rows, err := tx.Query(ctx,
		`SELECT id, type, comfort_base, last_fed_at, last_egg_at
		 FROM animals WHERE farm_id = $1 ORDER BY created_at`, id)
	if err != nil {
		return Farm{}, fmt.Errorf("query animals: %w", err)
	}
	var animals []animalRow
	for rows.Next() {
		var a animalRow
		if err := rows.Scan(&a.id, &a.typ, &a.comfortBase, &a.lastFedAt, &a.lastEggAt); err != nil {
			rows.Close()
			return Farm{}, fmt.Errorf("scan animal: %w", err)
		}
		animals = append(animals, a)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return Farm{}, fmt.Errorf("iterate animals: %w", err)
	}

	f.Animals = make([]AnimalView, 0, len(animals))
	credits := map[string]int{}
	for _, a := range animals {
		if product := animalProduct[a.typ]; product != "" {
			n, newLastEgg := eggsSince(a.comfortBase, a.lastFedAt, a.lastEggAt, now)
			if n > 0 {
				if _, err := tx.Exec(ctx,
					`UPDATE animals SET last_egg_at = $2 WHERE id = $1`, a.id, newLastEgg,
				); err != nil {
					return Farm{}, fmt.Errorf("advance last_egg_at: %w", err)
				}
				credits[product] += n
			}
		}
		f.Animals = append(f.Animals, AnimalView{
			Type:        a.typ,
			Comfort:     int(math.Round(currentComfort(a.comfortBase, a.lastFedAt, now))),
			Comfortable: comfortable(a.comfortBase, a.lastFedAt, now),
		})
	}
	for item, qty := range credits {
		if _, err := tx.Exec(ctx,
			`INSERT INTO farm_items (farm_id, item, qty) VALUES ($1, $2, $3)
			 ON CONFLICT (farm_id, item) DO UPDATE SET qty = farm_items.qty + $3`,
			id, item, qty,
		); err != nil {
			return Farm{}, fmt.Errorf("credit product: %w", err)
		}
	}

	// Plantações.
	type cropRow struct {
		plotIndex int
		typ       string
		plantedAt time.Time
	}
	crows, err := tx.Query(ctx,
		`SELECT plot_index, type, planted_at FROM crops WHERE farm_id = $1 ORDER BY plot_index`, id)
	if err != nil {
		return Farm{}, fmt.Errorf("query crops: %w", err)
	}
	var cropRows []cropRow
	for crows.Next() {
		var c cropRow
		if err := crows.Scan(&c.plotIndex, &c.typ, &c.plantedAt); err != nil {
			crows.Close()
			return Farm{}, fmt.Errorf("scan crop: %w", err)
		}
		cropRows = append(cropRows, c)
	}
	crows.Close()
	if err := crows.Err(); err != nil {
		return Farm{}, fmt.Errorf("iterate crops: %w", err)
	}
	f.Crops = make([]CropView, 0, len(cropRows))
	for _, c := range cropRows {
		f.Crops = append(f.Crops, CropView{
			Type:      c.typ,
			PlotIndex: c.plotIndex,
			Stage:     cropStage(c.typ, c.plantedAt, now),
			Ready:     cropReady(c.typ, c.plantedAt, now),
		})
	}

	// Inventário (após creditar produtos).
	irows, err := tx.Query(ctx,
		`SELECT item, qty FROM farm_items WHERE farm_id = $1 AND qty > 0 ORDER BY item`, id)
	if err != nil {
		return Farm{}, fmt.Errorf("query items: %w", err)
	}
	f.Items = make([]ItemView, 0)
	for irows.Next() {
		var it ItemView
		if err := irows.Scan(&it.Item, &it.Qty); err != nil {
			irows.Close()
			return Farm{}, fmt.Errorf("scan item: %w", err)
		}
		f.Items = append(f.Items, it)
	}
	irows.Close()
	if err := irows.Err(); err != nil {
		return Farm{}, fmt.Errorf("iterate items: %w", err)
	}

	f.Shop = shopCatalog()

	if err := tx.Commit(ctx); err != nil {
		return Farm{}, fmt.Errorf("commit get farm: %w", err)
	}
	return f, nil
}
