package domain

type IGameState interface {
	DealInitialHands()
}

type GameState struct {
	Round   Round
	Wall    Wall
	Players [4]*PlayerState
	IsWin   bool
}

func (g *GameState) IsGameEnd() bool {
	return g.Wall.Remaining() == 0 || g.IsWin
}

func (g *GameState) GetCurrentPlayerState() *PlayerState {
	currSeat := g.Round.CurrentPlayer
	return g.Players[currSeat]
}

func (g *GameState) DealInitialHands() {
	for playerIndex := range g.Players {
		for range 13 {
			_, ok := g.drawHandTile(g.Players[playerIndex])
			if !ok {
				return
			}
		}
	}

	dealerTile, ok := g.drawHandTile(g.Players[East])
	if !ok {
		return
	}
	g.Round.CurrentPlayer = SeatIndex(East)
	g.Round.NewlyDrawnTile = dealerTile
	g.Round.Phase = PhasePlay
}

// drawHandTile reveals flowers and animals and draws replacements until it
// finds a tile that belongs in the player's concealed hand.
func (g *GameState) drawHandTile(player *PlayerState) (*Tile, bool) {
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

func (g *GameState) PrintPlayersHand() {
	fmt.Println()
	fmt.Println("---------------------------------------------")
	for _, player := g.Players {
		playerJson, _ := json.MarshalIndent(player, "", "  ")
		fmt.Println(string(playerJson))
	}
	fmt.Println("---------------------------------------------")
	fmt.Println()
}
