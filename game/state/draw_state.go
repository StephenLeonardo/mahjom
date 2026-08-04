package state

import "mahjom/game/domain"

type DrawState struct {
	defaultState
	game *GameHandler
}

func (s *DrawState) GetStateName() string {
	return "draw state"
}

func (s *DrawState) drawHandTile(player *domain.PlayerState) (*domain.Tile, bool) {
	for {
		tile, ok := s.game.State.Wall.Draw()
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

func (s *DrawState) Resolve() error {
	round := s.game.State.Round
	if s.game.currState.GetStateName() != s.GetStateName() {
		return domain.ErrRoundNotInPlay
	}

	tile, ok := s.drawHandTile(s.game.State.GetCurrentPlayerState())
	if !ok {
		return domain.ErrWallEmpty
	}
	round.NewlyDrawnTile = tile
	s.game.currState = s.game.checkWinState
	return nil
}
