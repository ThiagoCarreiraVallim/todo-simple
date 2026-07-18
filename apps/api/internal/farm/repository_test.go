package farm

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/thiago/todo-simple-api/internal/database"
)

// Integração: roda apenas com DATABASE_URL definido (CI ou dev local).
func newTestRepository(t *testing.T) (*Repository, *pgxpool.Pool) {
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
	return NewRepository(pool), pool
}

// createTestFarm cria um usuário de teste + sua fazenda inicial (EnsureFarm) e
// devolve o userId. Limpa o usuário (cascade na fazenda) antes e depois.
func createTestFarm(t *testing.T, repo *Repository, username string) string {
	t.Helper()
	ctx := context.Background()
	_, _ = repo.pool.Exec(ctx, `DELETE FROM users WHERE username = $1`, username)
	var userID string
	if err := repo.pool.QueryRow(ctx,
		`INSERT INTO users (username) VALUES ($1) RETURNING id`, username).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := repo.EnsureFarm(ctx, userID); err != nil {
		t.Fatalf("ensure farm: %v", err)
	}
	t.Cleanup(func() {
		_, _ = repo.pool.Exec(ctx, `DELETE FROM users WHERE username = $1`, username)
	})
	return userID
}

func mustGet(t *testing.T, repo *Repository, userID string) Farm {
	t.Helper()
	f, err := repo.GetFarm(context.Background(), userID)
	if err != nil {
		t.Fatalf("get farm: %v", err)
	}
	return f
}

func itemQty(f Farm, item string) int {
	for _, it := range f.Items {
		if it.Item == item {
			return it.Qty
		}
	}
	return 0
}

func TestEnsureFarmSeedsChickenAndCrop(t *testing.T) {
	repo, _ := newTestRepository(t)
	userID := createTestFarm(t, repo, "test_seed")

	// EnsureFarm de novo é idempotente (não duplica galinha/milho).
	if err := repo.EnsureFarm(context.Background(), userID); err != nil {
		t.Fatalf("second ensure: %v", err)
	}

	f := mustGet(t, repo, userID)
	if f.Name != "" {
		t.Errorf("nome inicial deveria ser vazio, veio %q", f.Name)
	}
	if len(f.Animals) != 1 || f.Animals[0].Type != "chicken" {
		t.Errorf("esperava 1 galinha, veio %+v", f.Animals)
	}
	if len(f.Crops) != 1 || f.Crops[0].Type != "corn" {
		t.Errorf("esperava 1 milho, veio %+v", f.Crops)
	}
	if f.Animals[0].Comfortable {
		t.Errorf("galinha nova não deveria estar confortável (comfort=%d)", f.Animals[0].Comfort)
	}
	if f.Shop.MaxPlots != maxPlots || len(f.Shop.Animals) == 0 {
		t.Errorf("esperava catálogo da loja no estado, veio %+v", f.Shop)
	}
}

func TestFeedRaisesComfort(t *testing.T) {
	repo, _ := newTestRepository(t)
	userID := createTestFarm(t, repo, "test_feed")

	if err := repo.Feed(context.Background(), userID); err != nil {
		t.Fatalf("feed: %v", err)
	}
	f := mustGet(t, repo, userID)
	// 50 + 20 = 70 → confortável.
	if f.Animals[0].Comfort < 70 || !f.Animals[0].Comfortable {
		t.Errorf("após feed esperava conforto ~70 e confortável, veio comfort=%d comfortable=%v",
			f.Animals[0].Comfort, f.Animals[0].Comfortable)
	}
}

func TestGetFarmMaterializesEggs(t *testing.T) {
	repo, pool := newTestRepository(t)
	userID := createTestFarm(t, repo, "test_eggs")
	ctx := context.Background()

	// Galinha bem confortável e "voltando no tempo": alimentada há 1h, sem ovo há
	// 12h. Com base 100 ela fica confortável por 8h → ~2 ovos.
	if _, err := pool.Exec(ctx,
		`UPDATE animals SET comfort_base = 100,
		     last_fed_at = now() - interval '1 hour',
		     last_egg_at = now() - interval '12 hours'
		 WHERE farm_id = (SELECT id FROM farms WHERE user_id = $1)`, userID,
	); err != nil {
		t.Fatalf("time travel: %v", err)
	}

	eggs := itemQty(mustGet(t, repo, userID), "egg")
	if eggs < 1 {
		t.Fatalf("esperava ovos materializados, veio %d", eggs)
	}
	// Segunda leitura logo em seguida não deve creditar de novo (idempotente).
	if eggs2 := itemQty(mustGet(t, repo, userID), "egg"); eggs2 != eggs {
		t.Errorf("segunda leitura não deveria creditar mais ovos: %d → %d", eggs, eggs2)
	}
}

func TestSellItems(t *testing.T) {
	repo, pool := newTestRepository(t)
	userID := createTestFarm(t, repo, "test_sell")
	ctx := context.Background()

	if _, err := pool.Exec(ctx,
		`INSERT INTO farm_items (farm_id, item, qty)
		 VALUES ((SELECT id FROM farms WHERE user_id = $1), 'egg', 5)`, userID,
	); err != nil {
		t.Fatalf("seed inventory: %v", err)
	}

	if err := repo.Sell(ctx, userID, "egg", 3); err != nil {
		t.Fatalf("sell: %v", err)
	}
	f := mustGet(t, repo, userID)
	if f.Coins != 15 { // 3 ovos * 5
		t.Errorf("esperava 15 moedas, veio %d", f.Coins)
	}
	if q := itemQty(f, "egg"); q != 2 {
		t.Errorf("esperava 2 ovos restantes, veio %d", q)
	}
	if err := repo.Sell(ctx, userID, "egg", 5); !errors.Is(err, ErrInsufficientItems) {
		t.Errorf("esperava ErrInsufficientItems vendendo mais do que tem, veio %v", err)
	}
}

func TestBuyAnimal(t *testing.T) {
	repo, pool := newTestRepository(t)
	userID := createTestFarm(t, repo, "test_buy")
	ctx := context.Background()

	if err := repo.Buy(ctx, userID, "cow"); !errors.Is(err, ErrInsufficientCoins) {
		t.Errorf("sem moedas, esperava ErrInsufficientCoins, veio %v", err)
	}
	if _, err := pool.Exec(ctx,
		`UPDATE farms SET coins = 200 WHERE user_id = $1`, userID); err != nil {
		t.Fatalf("give coins: %v", err)
	}
	if err := repo.Buy(ctx, userID, "cow"); err != nil {
		t.Fatalf("buy cow: %v", err)
	}
	f := mustGet(t, repo, userID)
	if f.Coins != 50 { // 200 - 150
		t.Errorf("esperava 50 moedas após comprar vaca, veio %d", f.Coins)
	}
	if len(f.Animals) != 2 {
		t.Errorf("esperava 2 animais, veio %d", len(f.Animals))
	}
}

func TestPlantAndHarvest(t *testing.T) {
	repo, pool := newTestRepository(t)
	userID := createTestFarm(t, repo, "test_farm")
	ctx := context.Background()

	if _, err := pool.Exec(ctx,
		`UPDATE farms SET coins = 100 WHERE user_id = $1`, userID); err != nil {
		t.Fatalf("give coins: %v", err)
	}
	if err := repo.Plant(ctx, userID, "wheat"); err != nil {
		t.Fatalf("plant: %v", err)
	}
	f := mustGet(t, repo, userID)
	if f.Coins != 92 { // 100 - 8
		t.Errorf("esperava 92 moedas após plantar trigo, veio %d", f.Coins)
	}
	if len(f.Crops) != 2 {
		t.Errorf("esperava 2 canteiros (milho inicial + trigo), veio %d", len(f.Crops))
	}

	// Colher milho ainda verde (plot 0) falha.
	if err := repo.Harvest(ctx, userID, 0); !errors.Is(err, ErrCropNotReady) {
		t.Errorf("milho verde deveria dar ErrCropNotReady, veio %v", err)
	}

	// Amadurece o milho do plot 0 (24h de crescimento) e colhe.
	if _, err := pool.Exec(ctx,
		`UPDATE crops SET planted_at = now() - interval '48 hours'
		 WHERE plot_index = 0 AND farm_id = (SELECT id FROM farms WHERE user_id = $1)`, userID,
	); err != nil {
		t.Fatalf("ripen: %v", err)
	}
	if err := repo.Harvest(ctx, userID, 0); err != nil {
		t.Fatalf("harvest: %v", err)
	}
	f = mustGet(t, repo, userID)
	if q := itemQty(f, "corn"); q != cropYield["corn"] {
		t.Errorf("esperava %d milhos colhidos, veio %d", cropYield["corn"], q)
	}
	// Plot 0 foi liberado; sobra só o trigo.
	for _, c := range f.Crops {
		if c.PlotIndex == 0 {
			t.Errorf("plot 0 deveria ter sido liberado após colher, crops=%+v", f.Crops)
		}
	}
}
