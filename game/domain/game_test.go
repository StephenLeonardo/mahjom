package domain

import "testing"

func TestNewGameDealsIndependentWall(t *testing.T) {
	game := NewGame("test-game", &GameConfig{}, 42)

	if game.State.Round.Phase != PhasePlay {
		t.Fatalf("round phase = %v, want %v", game.State.Round.Phase, PhasePlay)
	}
	if game.State.Round.CurrentPlayer != SeatIndex(East) {
		t.Fatalf("current player = %v, want East", game.State.Round.CurrentPlayer)
	}
	if game.State.Round.LastDiscardBy != SeatNone {
		t.Fatalf("last discard by = %v, want no seat", game.State.Round.LastDiscardBy)
	}
	if game.State.Round.NewlyDrawnTile == nil {
		t.Fatal("East's newly drawn tile is nil")
	}

	for seat, player := range game.State.Players {
		wantHandSize := 13
		if seat == int(East) {
			wantHandSize = 14
		}
		if len(player.Hand) != wantHandSize {
			t.Fatalf("player %d hand size = %d, want %d", seat, len(player.Hand), wantHandSize)
		}
		for _, tile := range player.Hand {
			if tile.IsFlower() || tile.IsAnimal() {
				t.Fatalf("player %d has bonus tile %v in hand", seat, tile.ID)
			}
		}
	}

	accountedFor := len(game.State.Wall.DrawPile)
	for _, player := range game.State.Players {
		accountedFor += len(player.Hand) + len(player.Flowers) + len(player.Animals)
	}
	if accountedFor != TileSetSize {
		t.Fatalf("accounted tiles = %d, want %d", accountedFor, TileSetSize)
	}

	if len(game.State.Wall.DrawPile) > 0 && len(Catalog) > 0 && &game.State.Wall.DrawPile[0] == &Catalog[0] {
		t.Fatal("game wall reuses the canonical catalog slice")
	}
}
