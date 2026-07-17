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
			tile, ok := g.drawInitialHandTile(&g.Players[playerIndex])
			if !ok {
				return
			}
			g.Players[playerIndex].Hand = append(g.Players[playerIndex].Hand, tile)
		}
	}

	dealerTile, ok := g.drawInitialHandTile(&g.Players[East])
	if !ok {
		return
	}
	g.Players[East].Hand = append(g.Players[East].Hand, dealerTile)
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
			player.Flowers = append(player.Flowers, tile)
		case tile.IsAnimal():
			player.Animals = append(player.Animals, tile)
		default:
			return tile, true
		}
	}
}
