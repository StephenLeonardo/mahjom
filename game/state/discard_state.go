package state

import (
	"fmt"
	"mahjom/game/domain"
)

const (
	DISCARD_STATE = "DISCARD"
)

type DiscardState struct {
	defaultState
	game *GameHandler
}

func (s *DiscardState) GetStateName() string {
	return DISCARD_STATE
}

func (s *DiscardState) Resolve() error {
	if s.game.currState.GetStateName() != s.GetStateName() {
		return domain.ErrRoundNotInDiscard
	}

	round := s.game.State.Round

	// TODO: prompt user which tile to discard
	player := s.game.State.GetCurrentPlayerState()
	tile, scores := chooseDiscardTile(player)
	if tile == nil {
		return domain.ErrNoTileToDiscard
	}

	removed, ok := player.RemoveFromHand(tile.ID)
	if !ok {
		return fmt.Errorf("%s, tile: %s,", domain.ErrTileNotInHand.Error(), tile.Describe())
	}
	tile = removed

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
	round.TileDiscardScores = scores

	s.game.currState = s.game.claimState
	return nil
}

const (
	discardSameTileWeight = 100
	discardAdjacentWeight = 20
	discardGapWeight      = 10
	discardSuitedBase     = 1
)

func chooseDiscardTile(player *domain.PlayerState) (*domain.Tile, map[string]int) {
	if player == nil || len(player.Hand) == 0 {
		return nil, nil
	}

	counts := make(map[domain.Suit]map[uint8]int)
	for _, tile := range player.Hand {
		if tile == nil {
			continue
		}

		if _, ok := counts[tile.Suit]; !ok {
			counts[tile.Suit] = make(map[uint8]int)
		}

		counts[tile.Suit][tile.Rank]++
	}

	scores := make(map[string]int)

	var best *domain.Tile
	bestScore := int(^uint(0) >> 1)

	for _, tile := range player.Hand {
		if tile == nil || tile.IsBonus() {
			continue
		}

		score := discardTileScore(tile, counts)
		scores[tile.GetRankSuit()] = score

		if best == nil ||
			score < bestScore ||
			(score == bestScore && tile.ID < best.ID) {
			best = tile
			bestScore = score
		}
	}

	return best, scores
}

func discardTileScore(tile *domain.Tile, counts map[domain.Suit]map[uint8]int) int {
	score := 0
	suitCounts := counts[tile.Suit]
	if suitCounts == nil {
		return score
	}

	if tile.IsSuitedForChow() {
		score += discardSuitedBase
	}

	if sameCount := suitCounts[tile.Rank] - 1; sameCount > 0 {
		score += sameCount * discardSameTileWeight
	}

	if !tile.IsSuitedForChow() {
		return score
	}

	rank := int(tile.Rank)

	if left := rank - 1; left >= 1 {
		score += suitCounts[uint8(left)] * discardAdjacentWeight
	}
	if right := rank + 1; right <= 9 {
		score += suitCounts[uint8(right)] * discardAdjacentWeight
	}
	if gapLeft := rank - 2; gapLeft >= 1 {
		score += suitCounts[uint8(gapLeft)] * discardGapWeight
	}
	if gapRight := rank + 2; gapRight <= 9 {
		score += suitCounts[uint8(gapRight)] * discardGapWeight
	}

	return score
}
