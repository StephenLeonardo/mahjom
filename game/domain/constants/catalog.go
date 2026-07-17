// Package constants contains the canonical, ordered Singapore Mahjong tile catalog.
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

const (
	firstWindID   = 108
	firstDragonID = 124
	firstFlowerID = 136
	firstAnimalID = 144
)

// Catalog maps every TileID to its physical tile. Within each repeated group,
// tiles are ordered by suit, rank, then copy. Flowers have two copies of each
// of the four kinds; animals have one copy each.
var Catalog = newCatalog()

func newCatalog() []*domain.Tile {
	catalog := make([]*domain.Tile, domain.TileSetSize)

	for suit := domain.SuitCharacter; suit <= domain.SuitDot; suit++ {
		for rank := uint8(1); rank <= 9; rank++ {
			for copy := 0; copy < 4; copy++ {
				id := int(suit)*36 + int(rank-1)*4 + copy
				catalog[id] = &domain.Tile{
					ID:   domain.TileID(id),
					Suit: suit,
					Rank: rank,
				}
			}
		}
	}

	for rank := uint8(0); rank < 4; rank++ {
		for copy := 0; copy < 4; copy++ {
			id := firstWindID + int(rank)*4 + copy
			catalog[id] = &domain.Tile{ID: domain.TileID(id), Suit: domain.SuitWind, Rank: rank}
		}
	}

	for rank := uint8(0); rank < 3; rank++ {
		for copy := 0; copy < 4; copy++ {
			id := firstDragonID + int(rank)*4 + copy
			catalog[id] = &domain.Tile{ID: domain.TileID(id), Suit: domain.SuitDragon, Rank: rank}
		}
	}

	for rank := uint8(0); rank < 4; rank++ {
		for copy := 0; copy < 2; copy++ {
			id := firstFlowerID + int(rank)*2 + copy
			catalog[id] = &domain.Tile{ID: domain.TileID(id), Suit: domain.SuitFlower, Rank: rank}
		}
	}

	for rank := uint8(0); rank < 4; rank++ {
		id := firstAnimalID + int(rank)
		catalog[id] = &domain.Tile{ID: domain.TileID(id), Suit: domain.SuitAnimal, Rank: rank}
	}

	return catalog
}
