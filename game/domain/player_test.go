package domain

import "testing"

func TestPlayerStateAddTileRoutesTiles(t *testing.T) {
	player := PlayerState{}
	normal := &Tile{ID: 1, Suit: SuitCharacter, Rank: 1}
	flower := &Tile{ID: 136, Suit: SuitFlower, Rank: 0}
	animal := &Tile{ID: 144, Suit: SuitAnimal, Rank: 0}

	player.AddTile(normal)
	player.AddTile(flower)
	player.AddTile(animal)
	player.AddTile(nil)

	if len(player.Hand) != 1 || player.Hand[0] != normal {
		t.Fatalf("hand = %#v, want normal tile", player.Hand)
	}
	if len(player.Flowers) != 1 || player.Flowers[0] != flower {
		t.Fatalf("flowers = %#v, want flower tile", player.Flowers)
	}
	if len(player.Animals) != 1 || player.Animals[0] != animal {
		t.Fatalf("animals = %#v, want animal tile", player.Animals)
	}
}

func TestPlayerStateSortHand(t *testing.T) {
	player := PlayerState{Hand: []*Tile{
		{ID: 9, Suit: SuitWind, Rank: 0},
		{ID: 7, Suit: SuitCharacter, Rank: 2},
		{ID: 6, Suit: SuitCharacter, Rank: 1},
		{ID: 8, Suit: SuitBamboo, Rank: 1},
		{ID: 5, Suit: SuitCharacter, Rank: 1},
	}}

	player.SortHand()
	want := []TileID{5, 6, 7, 8, 9}
	for i, id := range want {
		if player.Hand[i].ID != id {
			t.Fatalf("hand[%d] = %d, want %d", i, player.Hand[i].ID, id)
		}
	}
}

func TestPlayerStateFindAndRemoveFromHand(t *testing.T) {
	tileOne := &Tile{ID: 1}
	tileTwo := &Tile{ID: 2}
	player := PlayerState{Hand: []*Tile{tileOne, tileTwo}}

	found, ok := player.FindInHand(tileTwo.ID)
	if !ok || found != tileTwo {
		t.Fatalf("FindInHand() = %v, %t; want tile 2, true", found, ok)
	}

	removed, ok := player.RemoveFromHand(tileOne.ID)
	if !ok || removed != tileOne {
		t.Fatalf("RemoveFromHand() = %v, %t; want tile 1, true", removed, ok)
	}
	if len(player.Hand) != 1 || player.Hand[0] != tileTwo {
		t.Fatalf("hand after removal = %#v, want only tile 2", player.Hand)
	}

	missing, ok := player.RemoveFromHand(99)
	if ok || missing != nil {
		t.Fatalf("RemoveFromHand() = %v, %t; want nil, false", missing, ok)
	}
}
