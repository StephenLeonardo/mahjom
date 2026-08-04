package state

import (
	"mahjom/game/domain"
	"testing"
)

func TestChooseDiscardTilePrefersIsolatedHonorOverUsefulTiles(t *testing.T) {
	player := &domain.PlayerState{
		Hand: []*domain.Tile{
			{ID: 1, Suit: domain.SuitBamboo, Rank: 2},
			{ID: 2, Suit: domain.SuitBamboo, Rank: 2},
			{ID: 3, Suit: domain.SuitBamboo, Rank: 4},
			{ID: 4, Suit: domain.SuitBamboo, Rank: 5},
			{ID: 5, Suit: domain.SuitWind, Rank: 0},
		},
	}

	tile := chooseDiscardTile(player)

	if tile == nil {
		t.Fatal("chooseDiscardTile() = nil, want an isolated tile")
	}
	if tile.ID != 5 {
		t.Fatalf("chooseDiscardTile() = %d, want 5", tile.ID)
	}
}

func TestChooseDiscardTileAvoidsDiscardingAFlatPair(t *testing.T) {
	player := &domain.PlayerState{
		Hand: []*domain.Tile{
			{ID: 10, Suit: domain.SuitCharacter, Rank: 7},
			{ID: 11, Suit: domain.SuitCharacter, Rank: 7},
			{ID: 12, Suit: domain.SuitDot, Rank: 1},
			{ID: 13, Suit: domain.SuitWind, Rank: 2},
			{ID: 14, Suit: domain.SuitWind, Rank: 3},
		},
	}

	tile := chooseDiscardTile(player)

	if tile == nil {
		t.Fatal("chooseDiscardTile() = nil, want a tile")
	}
	if tile.ID == 10 || tile.ID == 11 {
		t.Fatalf("chooseDiscardTile() = %d, want to preserve the pair", tile.ID)
	}
}

func TestChooseDiscardTileAvoidsDiscardingTwoInARow(t *testing.T) {
	player := &domain.PlayerState{
		Hand: []*domain.Tile{
			{ID: 20, Suit: domain.SuitBamboo, Rank: 4},
			{ID: 21, Suit: domain.SuitBamboo, Rank: 5},
			{ID: 22, Suit: domain.SuitDragon, Rank: domain.Fa},
			{ID: 23, Suit: domain.SuitDot, Rank: 9},
			{ID: 24, Suit: domain.SuitWind, Rank: 1},
		},
	}

	tile := chooseDiscardTile(player)

	if tile == nil {
		t.Fatal("chooseDiscardTile() = nil, want a tile")
	}
	if tile.ID == 20 || tile.ID == 21 {
		t.Fatalf("chooseDiscardTile() = %d, want to preserve the run potential", tile.ID)
	}
}

func TestChooseDiscardTileAvoidsDiscardingOneGapSequence(t *testing.T) {
	player := &domain.PlayerState{
		Hand: []*domain.Tile{
			{ID: 30, Suit: domain.SuitDot, Rank: 3},
			{ID: 31, Suit: domain.SuitDot, Rank: 5},
			{ID: 32, Suit: domain.SuitDragon, Rank: domain.Zhong},
			{ID: 33, Suit: domain.SuitWind, Rank: 2},
			{ID: 34, Suit: domain.SuitWind, Rank: 3},
		},
	}

	tile := chooseDiscardTile(player)

	if tile == nil {
		t.Fatal("chooseDiscardTile() = nil, want a tile")
	}
	if tile.ID == 30 || tile.ID == 31 {
		t.Fatalf("chooseDiscardTile() = %d, want to preserve the gap sequence potential", tile.ID)
	}
}
