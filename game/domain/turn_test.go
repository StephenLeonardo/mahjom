package domain

import (
	"errors"
	"testing"
)

func TestGameDiscardThenDraw(t *testing.T) {
	game := NewGame("turn-flow", GameConfig{}, 42)
	eastTile := game.State.Players[East].Hand[0]
	wallBeforeDiscard := game.State.Wall.Remaining()

	if err := game.Discard(eastTile.ID); err != nil {
		t.Fatalf("East discard: %v", err)
	}
	if len(game.State.Players[East].Hand) != 13 {
		t.Fatalf("East hand size = %d, want 13", len(game.State.Players[East].Hand))
	}
	if len(game.State.Players[East].Discards) != 1 || game.State.Players[East].Discards[0] != eastTile {
		t.Fatal("East discard was not recorded")
	}
	if game.State.Round.LastDiscard != eastTile || game.State.Round.LastDiscardBy != SeatIndex(East) {
		t.Fatal("round does not record East's discard")
	}
	if game.State.Round.NewlyDrawnTile != nil {
		t.Fatal("newly drawn tile remains after discard")
	}
	if game.State.Round.CurrentPlayer != SeatIndex(South) {
		t.Fatalf("current player = %d, want South", game.State.Round.CurrentPlayer)
	}
	if game.State.Wall.Remaining() != wallBeforeDiscard {
		t.Fatal("discard changed the wall")
	}

	if err := game.DrawForCurrentPlayer(); err != nil {
		t.Fatalf("South draw: %v", err)
	}
	south := &game.State.Players[South]
	if len(south.Hand) != 14 {
		t.Fatalf("South hand size = %d, want 14", len(south.Hand))
	}
	if game.State.Round.NewlyDrawnTile == nil {
		t.Fatal("draw did not set newly drawn tile")
	}
	if _, ok := south.FindInHand(game.State.Round.NewlyDrawnTile.ID); !ok {
		t.Fatal("newly drawn tile is not in South's hand")
	}
	if game.State.Wall.Remaining() >= wallBeforeDiscard {
		t.Fatal("draw did not reduce the wall")
	}
}

func TestGameTurnValidation(t *testing.T) {
	game := NewGame("turn-validation", GameConfig{}, 99)

	if err := game.DrawForCurrentPlayer(); !errors.Is(err, ErrMustDiscard) {
		t.Fatalf("East draw before discard = %v, want %v", err, ErrMustDiscard)
	}
	if err := game.Discard(TileID(255)); !errors.Is(err, ErrTileNotInHand) {
		t.Fatalf("discarding absent tile = %v, want %v", err, ErrTileNotInHand)
	}
}

func TestGameDrawReplacesBonusTiles(t *testing.T) {
	flower := &Tile{ID: 136, Suit: SuitFlower, Rank: 0}
	animal := &Tile{ID: 144, Suit: SuitAnimal, Rank: 0}
	normal := &Tile{ID: 0, Suit: SuitCharacter, Rank: 1}
	game := &Game{State: GameState{
		Round: Round{Phase: PhasePlay, CurrentPlayer: SeatIndex(East)},
		Wall:  Wall{DrawPile: []*Tile{flower, animal, normal}},
	}}

	if err := game.DrawForCurrentPlayer(); err != nil {
		t.Fatalf("draw with bonus replacements: %v", err)
	}
	player := &game.State.Players[East]
	if len(player.Hand) != 1 || player.Hand[0] != normal {
		t.Fatal("normal replacement tile was not added to hand")
	}
	if len(player.Flowers) != 1 || player.Flowers[0] != flower {
		t.Fatal("flower was not exposed")
	}
	if len(player.Animals) != 1 || player.Animals[0] != animal {
		t.Fatal("animal was not exposed")
	}
	if game.State.Round.NewlyDrawnTile != normal {
		t.Fatal("normal replacement tile was not recorded as newly drawn")
	}
	if game.State.Wall.Remaining() != 0 {
		t.Fatalf("wall remaining = %d, want 0", game.State.Wall.Remaining())
	}
}
