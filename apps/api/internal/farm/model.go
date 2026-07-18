package farm

import "time"

// Farm é a forma pública da fazenda. As chaves internas (id, user_id) nunca são
// serializadas — a fazenda é acessada pelo userId do dono.
type Farm struct {
	Name      string       `json:"name"`
	Coins     int          `json:"coins"`
	Animals   []AnimalView `json:"animals"`
	Crops     []CropView   `json:"crops"`
	Items     []ItemView   `json:"items"`
	Shop      Shop         `json:"shop"`
	CreatedAt time.Time    `json:"createdAt"`
}

// Shop é o catálogo (estático) enviado junto do estado para o cliente montar a
// loja sem duplicar preços — o servidor continua sendo a autoridade.
type Shop struct {
	Animals  []PriceEntry `json:"animals"`  // animais compráveis
	Seeds    []PriceEntry `json:"seeds"`    // sementes plantáveis
	Sellable []PriceEntry `json:"sellable"` // itens do inventário vendáveis
	MaxPlots int          `json:"maxPlots"`
}

// PriceEntry é um item de catálogo: um tipo/item e seu preço em moedas.
type PriceEntry struct {
	Type  string `json:"type"`
	Price int    `json:"price"`
}

// AnimalView expõe o estado derivado (conforto calculado), não os timestamps crus.
type AnimalView struct {
	Type        string `json:"type"`
	Comfort     int    `json:"comfort"`     // 0..100
	Comfortable bool   `json:"comfortable"` // anda feliz / bota ovos
}

type CropView struct {
	Type      string `json:"type"`
	PlotIndex int    `json:"plotIndex"`
	Stage     int    `json:"stage"`
	Ready     bool   `json:"ready"`
}

type ItemView struct {
	Item string `json:"item"`
	Qty  int    `json:"qty"`
}
