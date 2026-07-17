package domain

type IGameState interface {
	DealInitialHands()
}

type GameState struct {
	Round   Round
	Wall    Wall
	Players [4]PlayerState
}

func (g *GameState) DealInitialHands() {
	for playerIndex := range g.Players {
		for range 13 {
			_, ok := g.drawInitialHandTile(&g.Players[playerIndex])
			if !ok {
				return
			}
		}
	}

	dealerTile, ok := g.drawInitialHandTile(&g.Players[East])
	if !ok {
		return
	}
	g.Round.CurrentPlayer = SeatIndex(East)
	g.Round.NewlyDrawnTile = dealerTile
	g.Round.Phase = PhasePlay
}

// drawInitialHandTile reveals flowers and animals and draws replacements until
// it finds a tile that belongs in the player's concealed hand.
func (g *GameState) drawInitialHandTile(player *PlayerState) (*Tile, bool) {
	for {
		tile, ok := g.Wall.Draw()
		if !ok {
			return nil, false
		}

		switch {
		case tile.IsFlower():
			player.AddTile(tile)
		case tile.IsAnimal():
			player.AddTile(tile)
		default:
			player.AddTile(tile)
			return tile, true
		}
	}
}
