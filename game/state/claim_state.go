package state

import (
	"mahjom/game/domain"
	"mahjom/game/utils"
	"sort"
)

type ClaimState struct {
	defaultState
	game *GameHandler
}

func (s *ClaimState) GetStateName() string {
	return "claim state"
}

func (s *ClaimState) Resolve() error {
	if s.game.currState.GetStateName() != s.GetStateName() {
		return domain.ErrTileNotClaim
	}

	// generate all possible claims from all players
	playersPossibleClaims := s.getAllEligibleClaims()

	var selectedClaimsPerPlayer [4]*domain.ClaimDecl

	// TODO: prompt to user to select their eligible claims
	for i, playerClaims := range playersPossibleClaims {
		// TODO: add some rng to simulate a pass action
		if len(playerClaims) == 1 {
			selectedClaimsPerPlayer[i] = playerClaims[0]
			continue
		}

		// TODO: prompt user with goroutine
		selectedClaimsPerPlayer[i] = playerClaims[0]
	}

	return nil
}

func (s *ClaimState) getAllEligibleClaims() [][]*domain.ClaimDecl {

	res := make([][]*domain.ClaimDecl, 4)

	round := s.game.State.Round
	players := s.game.State.Players

	lastDiscardBy := round.LastDiscardBy
	lastDiscard := round.LastDiscard
	for playerIndex := range players {
		if domain.SeatIndex(playerIndex) == lastDiscardBy {
			res[playerIndex] = []*domain.ClaimDecl{}
			continue
		}
		currPlayerPossibleClaims := s.generateAllPossibleClaimsForPlayer(lastDiscardBy, domain.SeatIndex(playerIndex), lastDiscard)
		res[playerIndex] = currPlayerPossibleClaims
	}

	return res
}

func (s *ClaimState) generateAllPossibleClaimsForPlayer(discarder, claimant domain.SeatIndex, discarded *domain.Tile) []*domain.ClaimDecl {
	player := s.game.State.Players[claimant]
	res := []*domain.ClaimDecl{}

	for _, meldType := range []domain.MeldType{domain.MeldPong, domain.MeldKong, domain.MeldChow} {
		switch meldType {
		case domain.MeldPong:
			tiles := player.GetMatchingInHand(discarded)
			if len(tiles) >= 3 {
				res = append(res, s.generateClaimDecl(discarder, claimant, tiles[:3], domain.MeldPong, domain.KongNone))
			}
		case domain.MeldKong:
			tiles := player.GetMatchingInHand(discarded)
			if len(tiles) >= 4 {
				res = append(res, s.generateClaimDecl(discarder, claimant, tiles[:4], domain.MeldKong, domain.KongExposed))
			}
		case domain.MeldChow:
			// Chow is only valid on suited tiles (Wan/Bamboo/Dot).
			if !discarded.IsSuitedForChow() {
				continue
			}
			// Chow is only valid from the seat immediately after the
			// discarder. Other seats can Pong or Kong the same tile, but
			// not Chow it.
			if utils.SeatDistance(discarder, claimant) != 1 {
				continue
			}
			chowTiles := s.getAllPossibleChowCombinations(player, discarded)
			for _, tiles := range chowTiles {
				res = append(res, s.generateClaimDecl(discarder, claimant, tiles, domain.MeldChow, domain.KongNone))
			}
		}
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Type > res[j].Type
	})

	return res
}

func (s *ClaimState) generateClaimDecl(
	discarder,
	claimantIndex domain.SeatIndex,
	tiles []*domain.Tile,
	meldType domain.MeldType,
	kongType domain.KongType,
) *domain.ClaimDecl {
	claimDecl := &domain.ClaimDecl{
		Discarder: discarder,
		Claimant:  claimantIndex,
		Type:      meldType,
		Kong:      kongType,
		Tiles:     tiles,
	}

	return claimDecl
}

// getAllPossibleChowCombinations finds all possible 3-tile sequences (Chows) that can be
// formed using an opponent's discard and two tiles from your hand.
func (s *ClaimState) getAllPossibleChowCombinations(p *domain.PlayerState, discarded *domain.Tile) [][]*domain.Tile {
	if discarded == nil {
		return nil
	}

	var results [][]*domain.Tile

	// TODO: optimize
	// Helper to find all tiles in hand matching a specific suit and rank
	findTiles := func(suit domain.Suit, rank uint8) *domain.Tile {
		for _, h := range p.Hand {
			if h.Suit == suit && h.Rank == rank {
				return h
			}
		}
		return nil
	}

	// Sequences are (D-2, D-1, D), (D-1, D, D+1), or (D, D+1, D+2)
	// We check every possibility.
	// Case A: Discard is High (D-2, D-1, D)
	if discarded.Rank >= 3 {
		t1, t2 := findTiles(discarded.Suit, discarded.Rank-2), findTiles(discarded.Suit, discarded.Rank-1)
		if t1 != nil && t2 != nil {
			results = append(results, []*domain.Tile{t1, t2, discarded})
		}
	}
	// Case B: Discard is Middle (D-1, D, D+1)
	if discarded.Rank >= 2 && discarded.Rank <= 8 {
		t1, t3 := findTiles(discarded.Suit, discarded.Rank-1), findTiles(discarded.Suit, discarded.Rank+1)
		if t1 != nil && t3 != nil {
			results = append(results, []*domain.Tile{t1, discarded, t3})
		}
	}
	// Case C: Discard is Low (D, D+1, D+2)
	if discarded.Rank <= 7 {
		t2, t3 := findTiles(discarded.Suit, discarded.Rank+1), findTiles(discarded.Suit, discarded.Rank+2)
		if t2 != nil && t3 != nil {
			results = append(results, []*domain.Tile{discarded, t2, t3})
		}
	}

	return results
}
