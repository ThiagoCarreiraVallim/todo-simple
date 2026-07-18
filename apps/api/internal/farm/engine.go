package farm

import "time"

// Idle engine — todo o estado é derivado de timestamps, sem worker/cron.
// As funções são puras (recebem `now`) para serem testáveis.

const (
	decayPerHour     = 5.0  // conforto perdido por hora
	comfortThreshold = 60.0 // a partir daqui o animal fica "confortável" e bota ovos
	feedAmount       = 20.0 // conforto ganho por ração
	maxComfort       = 100.0
	eggIntervalHours = 4.0 // um ovo a cada 4h confortáveis
)

// currentComfort é o conforto atual, decaindo linearmente desde a última ração.
func currentComfort(comfortBase float64, lastFedAt, now time.Time) float64 {
	c := comfortBase - now.Sub(lastFedAt).Hours()*decayPerHour
	if c < 0 {
		return 0
	}
	if c > maxComfort {
		return maxComfort
	}
	return c
}

func comfortable(comfortBase float64, lastFedAt, now time.Time) bool {
	return currentComfort(comfortBase, lastFedAt, now) >= comfortThreshold
}

// comfortableUntil é o instante em que o conforto cruza o threshold para baixo.
// Antes dele o animal está confortável; depois, não. Se já nasce abaixo do
// threshold, retorna lastFedAt (janela vazia).
func comfortableUntil(comfortBase float64, lastFedAt time.Time) time.Time {
	if comfortBase < comfortThreshold {
		return lastFedAt
	}
	hours := (comfortBase - comfortThreshold) / decayPerHour
	return lastFedAt.Add(time.Duration(hours * float64(time.Hour)))
}

// feedComfort é o novo comfort_base após uma ração: parte do conforto atual e
// soma feedAmount (limitado a maxComfort).
func feedComfort(comfortBase float64, lastFedAt, now time.Time) float64 {
	c := currentComfort(comfortBase, lastFedAt, now) + feedAmount
	if c > maxComfort {
		return maxComfort
	}
	return c
}

// eggsSince conta quantos ovos foram botados entre lastEggAt e now, contando
// apenas o tempo em que o animal esteve confortável (janela
// [lastEggAt, min(now, comfortableUntil)]). Determinístico e independente de
// quando a fazenda é lida. Retorna o número de ovos e o novo lastEggAt
// (avançado por múltiplos exatos do intervalo, preservando o resto).
func eggsSince(comfortBase float64, lastFedAt, lastEggAt, now time.Time) (int, time.Time) {
	end := comfortableUntil(comfortBase, lastFedAt)
	if now.Before(end) {
		end = now
	}
	window := end.Sub(lastEggAt).Hours()
	if window < eggIntervalHours {
		return 0, lastEggAt
	}
	n := int(window / eggIntervalHours)
	newLastEggAt := lastEggAt.Add(time.Duration(float64(n) * eggIntervalHours * float64(time.Hour)))
	return n, newLastEggAt
}

// cropConfig descreve o crescimento de uma cultura.
type cropConfig struct {
	growHours float64 // tempo total até ficar pronta
	maxStage  int     // estágio final (pronta)
}

var crops = map[string]cropConfig{
	"corn":  {growHours: 24, maxStage: 3},
	"wheat": {growHours: 18, maxStage: 3},
	"apple": {growHours: 48, maxStage: 4},
}

// cropStage é o estágio atual (0..maxStage) de uma cultura, derivado do tempo
// desde o plantio. Culturas desconhecidas ficam no estágio 0.
func cropStage(cropType string, plantedAt, now time.Time) int {
	cfg, ok := crops[cropType]
	if !ok {
		return 0
	}
	elapsed := now.Sub(plantedAt).Hours()
	if elapsed <= 0 {
		return 0
	}
	stage := int(elapsed / (cfg.growHours / float64(cfg.maxStage)))
	if stage > cfg.maxStage {
		return cfg.maxStage
	}
	return stage
}

func cropReady(cropType string, plantedAt, now time.Time) bool {
	cfg, ok := crops[cropType]
	if !ok {
		return false
	}
	return cropStage(cropType, plantedAt, now) >= cfg.maxStage
}

// --- Economia -------------------------------------------------------------
//
// Catálogos estáticos e ordenados (fonte única de preços). O servidor é sempre
// a autoridade; o mesmo catálogo é enviado ao cliente para montar a loja.

// maxPlots é o número de canteiros de plantação da fazenda.
const maxPlots = 6

// animalCatalog são os animais compráveis na loja (a galinha inicial é grátis).
var animalCatalog = []PriceEntry{
	{Type: "chicken", Price: 25},
	{Type: "rooster", Price: 30},
	{Type: "duck", Price: 45},
	{Type: "rabbit", Price: 40},
	{Type: "pig", Price: 90},
	{Type: "cow", Price: 150},
}

// seedCatalog são as sementes plantáveis (preço por canteiro).
var seedCatalog = []PriceEntry{
	{Type: "corn", Price: 10},
	{Type: "wheat", Price: 8},
	{Type: "apple", Price: 20},
}

// sellCatalog são os itens de inventário que podem ser vendidos por moedas.
var sellCatalog = []PriceEntry{
	{Type: "egg", Price: 5},
	{Type: "milk", Price: 12},
	{Type: "corn", Price: 8},
	{Type: "wheat", Price: 6},
	{Type: "apple", Price: 15},
}

// animalProduct diz qual item o animal produz enquanto está confortável
// (mesma cadência de eggsSince). Tipos ausentes são decorativos (não produzem).
var animalProduct = map[string]string{
	"chicken": "egg",
	"duck":    "egg",
	"cow":     "milk",
}

// cropYield é quantos itens uma colheita madura rende.
var cropYield = map[string]int{
	"corn":  3,
	"wheat": 4,
	"apple": 2,
}

func lookupPrice(catalog []PriceEntry, typ string) (int, bool) {
	for _, e := range catalog {
		if e.Type == typ {
			return e.Price, true
		}
	}
	return 0, false
}

func animalPrice(typ string) (int, bool) { return lookupPrice(animalCatalog, typ) }
func seedPrice(typ string) (int, bool)   { return lookupPrice(seedCatalog, typ) }
func sellPrice(item string) (int, bool)  { return lookupPrice(sellCatalog, item) }

// shopCatalog é o catálogo estático completo enviado no estado da fazenda.
func shopCatalog() Shop {
	return Shop{
		Animals:  animalCatalog,
		Seeds:    seedCatalog,
		Sellable: sellCatalog,
		MaxPlots: maxPlots,
	}
}
