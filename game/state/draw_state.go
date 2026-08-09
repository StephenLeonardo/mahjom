package state

import "mahjom/game/domain"

const (
	DRAW_STATE = "DRAW"
)

type DrawState struct {
	defaultState
	game *GameHandler
}

func (s *DrawState) GetStateName() string {
	return DRAW_STATE
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

	if s.checkAndResolveKongFromHand(tile) || s.checkAndResolveKongFromMelds(tile) {
		s.game.currState = s.game.drawState
		return nil
	}

	s.game.currState = s.game.checkWinState
	return nil
}

func (s *DrawState) checkAndResolveKongFromHand(tile *domain.Tile) (hasKong bool) {
	player := s.game.State.GetCurrentPlayerState()
	tiles := player.FindMatchingTilesFromHand(tile)
	if len(tiles) != 4 {
		return
	}

	hasKong = true

	meld := &domain.Meld{
		Type:         domain.MeldKong,
		Kong:         domain.KongExposed,
		Tiles:        tiles,
		FromSeat:     domain.SeatNone,
		FromPosition: "",
	}
	player.Melds = append(player.Melds, meld)

	for _, t := range tiles {
		player.RemoveFromHand(t.ID)
	}

	return
}

func (s *DrawState) checkAndResolveKongFromMelds(tile *domain.Tile) (hasKong bool) {
	player := s.game.State.GetCurrentPlayerState()

	for _, meld := range player.Melds {
		if meld.Type != domain.MeldPong || len(meld.Tiles) == 0 {
			continue
		}

		if meld.Tiles[0].GetRankSuit() != tile.GetRankSuit() {
			continue
		}

		hasKong = true
		meld.Tiles = append(meld.Tiles, tile)
		meld.Type = domain.MeldKong
		meld.Kong = domain.KongAdded
	}

	return
}
