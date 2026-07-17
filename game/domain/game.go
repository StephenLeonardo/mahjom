package domain

type Game struct {
	ID     string
	Config GameConfig
	State  GameState
}

// NewGame creates and deals a new round using a deterministic shuffle seed.
func NewGame(id string, config GameConfig, seed uint64) *Game {
	wallTiles := append([]*Tile(nil), Catalog...)
	game := &Game{
		ID:     id,
		Config: config,
		State: GameState{
			Round: Round{
				Phase:         PhaseDealing,
				CurrentPlayer: SeatIndex(East),
				LastDiscardBy: SeatNone,
			},
			Wall: Wall{DrawPile: wallTiles},
		},
	}

	game.State.Wall.Shuffle(seed)
	game.State.DealInitialHands()
	return game
}
