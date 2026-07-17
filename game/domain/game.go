package domain

import "errors"

var (
	ErrRoundNotInPlay = errors.New("round is not in play")
	ErrMustDraw       = errors.New("player must draw before discarding")
	ErrMustDiscard    = errors.New("player must discard before drawing")
	ErrTileNotInHand  = errors.New("tile is not in the player's hand")
	ErrWallEmpty      = errors.New("wall is empty")
)

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

// DrawForCurrentPlayer draws a tile for the current player. Flowers and
// animals are exposed and replaced until a normal hand tile is drawn.
func (g *Game) DrawForCurrentPlayer() error {
	round := &g.State.Round
	if round.Phase != PhasePlay {
		return ErrRoundNotInPlay
	}
	if round.NewlyDrawnTile != nil {
		return ErrMustDiscard
	}

	tile, ok := g.State.drawHandTile(&g.State.Players[round.CurrentPlayer])
	if !ok {
		return ErrWallEmpty
	}
	round.NewlyDrawnTile = tile
	return nil
}

// Discard removes a tile from the current player's hand and advances play to
// the next player. Claim handling is added in Phase 4.
func (g *Game) Discard(id TileID) error {
	round := &g.State.Round

	// Todo: change this once we implement claim
	if round.Phase != PhasePlay {
		return ErrRoundNotInPlay
	}
	if round.NewlyDrawnTile == nil {
		return ErrMustDraw
	}

	player := round.CurrentPlayer
	tile, ok := g.State.Players[player].RemoveFromHand(id)
	if !ok {
		return ErrTileNotInHand
	}
	g.State.Players[player].Discards = append(g.State.Players[player].Discards, tile)
	round.LastDiscard = tile
	round.LastDiscardBy = player
	round.NewlyDrawnTile = nil
	round.CurrentPlayer = nextSeat(player)
	return nil
}

func nextSeat(seat SeatIndex) SeatIndex {
	return SeatIndex((int(seat) + 1) % 4)
}
