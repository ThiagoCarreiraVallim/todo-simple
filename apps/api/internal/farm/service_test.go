package farm

import "testing"

func TestCatalogLookups(t *testing.T) {
	if p, ok := animalPrice("cow"); !ok || p != 150 {
		t.Errorf("animalPrice(cow) = (%d, %v), want (150, true)", p, ok)
	}
	if _, ok := animalPrice("dragon"); ok {
		t.Error("animalPrice(dragon) deveria ser desconhecido")
	}
	if p, ok := seedPrice("corn"); !ok || p != 10 {
		t.Errorf("seedPrice(corn) = (%d, %v), want (10, true)", p, ok)
	}
	if p, ok := sellPrice("milk"); !ok || p != 12 {
		t.Errorf("sellPrice(milk) = (%d, %v), want (12, true)", p, ok)
	}
	if _, ok := sellPrice("chicken"); ok {
		t.Error("sellPrice(chicken) não deveria ser vendável")
	}
	if animalProduct["chicken"] != "egg" || animalProduct["cow"] != "milk" {
		t.Error("produtos dos animais inesperados")
	}
	if animalProduct["pig"] != "" {
		t.Error("porco não deveria produzir item")
	}
}
