// Package constants contains compatibility aliases for the Singapore Mahjong catalog.
package constants

import "mahjom/game/domain"

const (
	FlowerPlum uint8 = iota
	FlowerOrchid
	FlowerChrysanthemum
	FlowerBamboo
)

const (
	AnimalCat uint8 = iota
	AnimalMouse
	AnimalRooster
	AnimalCentipede
)

// Catalog is retained for callers that import the constants package. New
// domain code should use domain.Catalog directly.
var Catalog = domain.Catalog
