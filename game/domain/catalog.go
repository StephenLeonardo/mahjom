package domain

const (
	firstWindID   = 108
	firstDragonID = 124
	firstFlowerID = 136
	firstAnimalID = 144
)

// Catalog is the canonical, ordered Singapore Mahjong tile catalog. Within
// each repeated group, tiles are ordered by suit, rank, then copy.
//
// Do not shuffle this slice directly. Each game copies it into its own wall.
var Catalog = newCatalog()

func newCatalog() []*Tile {
	catalog := make([]*Tile, TileSetSize)

	for suit := SuitCharacter; suit <= SuitDot; suit++ {
		for rank := uint8(1); rank <= 9; rank++ {
			for copy := 0; copy < 4; copy++ {
				id := int(suit)*36 + int(rank-1)*4 + copy
				catalog[id] = &Tile{ID: TileID(id), Suit: suit, Rank: rank}
			}
		}
	}

	for rank := uint8(0); rank < 4; rank++ {
		for copy := 0; copy < 4; copy++ {
			id := firstWindID + int(rank)*4 + copy
			catalog[id] = &Tile{ID: TileID(id), Suit: SuitWind, Rank: rank}
		}
	}

	for rank := uint8(0); rank < 3; rank++ {
		for copy := 0; copy < 4; copy++ {
			id := firstDragonID + int(rank)*4 + copy
			catalog[id] = &Tile{ID: TileID(id), Suit: SuitDragon, Rank: rank}
		}
	}

	for rank := uint8(0); rank < 4; rank++ {
		for copy := 0; copy < 2; copy++ {
			id := firstFlowerID + int(rank)*2 + copy
			catalog[id] = &Tile{ID: TileID(id), Suit: SuitFlower, Rank: rank}
		}
	}

	for rank := uint8(0); rank < 4; rank++ {
		id := firstAnimalID + int(rank)
		catalog[id] = &Tile{ID: TileID(id), Suit: SuitAnimal, Rank: rank}
	}

	return catalog
}
