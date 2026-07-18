package farm

import (
	"testing"
	"time"
)

var base = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func TestCurrentComfortDecays(t *testing.T) {
	// 80 de base, 4h depois a 5/h → 80 - 20 = 60.
	if got := currentComfort(80, base, base.Add(4*time.Hour)); got != 60 {
		t.Errorf("comfort após 4h: got %v, want 60", got)
	}
	// nunca abaixo de 0.
	if got := currentComfort(80, base, base.Add(100*time.Hour)); got != 0 {
		t.Errorf("comfort clamp inferior: got %v, want 0", got)
	}
	// nunca acima de 100.
	if got := currentComfort(200, base, base); got != 100 {
		t.Errorf("comfort clamp superior: got %v, want 100", got)
	}
}

func TestFeedComfort(t *testing.T) {
	// conforto atual 60 + 20 = 80.
	if got := feedComfort(80, base, base.Add(4*time.Hour)); got != 80 {
		t.Errorf("feed: got %v, want 80", got)
	}
	// limitado a 100.
	if got := feedComfort(95, base, base); got != 100 {
		t.Errorf("feed clamp: got %v, want 100", got)
	}
}

func TestEggsSinceCountsOnlyComfortableWindow(t *testing.T) {
	// base 100: confortável por (100-60)/5 = 8h. Em 8h → 2 ovos (a cada 4h).
	n, newLastEgg := eggsSince(100, base, base, base.Add(24*time.Hour))
	if n != 2 {
		t.Fatalf("ovos: got %d, want 2 (só a janela confortável de 8h conta)", n)
	}
	// lastEggAt avança 2*4h, preservando o resto.
	if !newLastEgg.Equal(base.Add(8 * time.Hour)) {
		t.Errorf("novo lastEggAt: got %v, want base+8h", newLastEgg)
	}
}

func TestEggsSinceNoneWhenUncomfortable(t *testing.T) {
	// base 50 já está abaixo do threshold: nenhuma janela confortável.
	n, newLastEgg := eggsSince(50, base, base, base.Add(48*time.Hour))
	if n != 0 || !newLastEgg.Equal(base) {
		t.Errorf("uncomfortable: got %d ovos / %v, want 0 / base", n, newLastEgg)
	}
}

func TestEggsSinceNoneBeforeInterval(t *testing.T) {
	// confortável, mas só 3h passadas → 0 ovos.
	n, _ := eggsSince(100, base, base, base.Add(3*time.Hour))
	if n != 0 {
		t.Errorf("antes do intervalo: got %d, want 0", n)
	}
}

func TestCropStage(t *testing.T) {
	// corn: 24h, maxStage 3 → cada estágio 8h.
	cases := []struct {
		h    int
		want int
	}{{0, 0}, {7, 0}, {8, 1}, {16, 2}, {24, 3}, {100, 3}}
	for _, c := range cases {
		if got := cropStage("corn", base, base.Add(time.Duration(c.h)*time.Hour)); got != c.want {
			t.Errorf("cropStage(corn, +%dh): got %d, want %d", c.h, got, c.want)
		}
	}
	// cultura desconhecida fica em 0.
	if got := cropStage("dragonfruit", base, base.Add(999*time.Hour)); got != 0 {
		t.Errorf("cultura desconhecida: got %d, want 0", got)
	}
	if !cropReady("corn", base, base.Add(24*time.Hour)) {
		t.Error("corn deveria estar pronta em 24h")
	}
}
