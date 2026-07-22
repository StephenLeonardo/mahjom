package state

import (
	"fmt"
	"mahjom/game/domain"
	"time"
)

type DiscardState struct {
	defaultState
	game *GameHandler
}

func (s *DiscardState) GetStateName() string {
	return "discard state"
}

func (s *DiscardState) Resolve() error {
	if s.game.currState.GetStateName() != s.GetStateName() {
		return domain.ErrRoundNotInDiscard
	}

	round := s.game.State.Round

	// TODO: prompt user which tile to discard
	player := s.game.State.GetCurrentPlayerState()
	seed := uint64(time.Now().Unix())
	tile := player.DiscardRandomTileFromHand(seed)

	if tile.IsBonus() {
		// Defensive: a bonus tile should never be in the hand, but if a
		// caller manages to construct one, refuse to discard it.
		player.Hand = append(player.Hand, tile)
		fmt.Printf("Warn Log: player is discarding a bonus tile: %s\n", tile.Describe())
		return domain.ErrTileNotInHand
	}

	player.AddToDiscard(tile)

	// fmt.Printf("Player %s Discarded %s\n", player.ID, tile.Describe())

	round.LastDiscard = tile
	round.LastDiscardBy = s.game.State.Round.CurrentPlayer
	round.NewlyDrawnTile = nil

	s.game.currState = s.game.claimState
	return nil
}
