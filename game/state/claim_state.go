package state

import (
	"mahjom/game/domain"
	"mahjom/game/utils"
	"sort"
)

const (
	CLAIM_STATE = "CLAIM"
)

type ClaimState struct {
	defaultState
	game *GameHandler
}

func (s *ClaimState) GetStateName() string {
	return CLAIM_STATE
}

func (s *ClaimState) Resolve() error {
	if s.game.currState.GetStateName() != s.GetStateName() {
		return domain.ErrTileNotClaim
	}

	// generate all possible claims from all players
	playersPossibleClaims := s.getAllEligibleClaims()

	selectedClaimsPerPlayer := make([]*domain.ClaimDecl, 0, 4)

	// TODO: prompt to user to select their eligible claims
	for _, playerClaims := range playersPossibleClaims {
		if len(playerClaims) == 0 {
			continue
		}

		// TODO: add some rng to simulate a pass action
		if len(playerClaims) == 1 {
			selectedClaimsPerPlayer = append(selectedClaimsPerPlayer, playerClaims[0])
			continue
		}

		// TODO: prompt user with goroutine
		selectedClaimsPerPlayer = append(selectedClaimsPerPlayer, playerClaims[0])
	}

	winningClaim := s.getWinningClaim(selectedClaimsPerPlayer)
	round := s.game.State.Round

	if winningClaim == nil {
		s.game.currState = s.game.drawState
		round.CurrentPlayer = round.GetNextSeatAfter(round.CurrentPlayer)
		return nil
	}

	round.ClaimV3 = winningClaim

	round.LastDiscard = nil
	round.LastDiscardBy = domain.SeatNone
	round.CurrentPlayer = winningClaim.Claimant

	player := s.game.State.Players[winningClaim.Claimant]
	player.AddToMelds(winningClaim)
	for _, t := range winningClaim.Tiles {
		player.RemoveFromHand(t.ID)
	}

	if winningClaim.ClaimType == domain.ClaimHu {
		s.game.currState = s.game.checkWinState
		return nil
	}

	s.game.currState = s.game.discardState
	return nil
}

func (s *ClaimState) getWinningClaim(claims []*domain.ClaimDecl) *domain.ClaimDecl {
	if len(claims) == 0 {
		return nil
	}
	discarder := int(claims[0].Discarder)
	n := 4
	sort.Slice(claims, func(i, j int) bool {
		if claims[i].ClaimType == claims[j].ClaimType {
			rankI := (int(claims[i].Claimant) - discarder + n) % n
			rankJ := (int(claims[j].Claimant) - discarder + n) % n
			return rankI < rankJ
		}

		return claims[i].ClaimType < claims[j].ClaimType
	})

	return claims[0]
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
	res := make([]*domain.ClaimDecl, 0, 8)

	matching := player.GetMatchingInHand(discarded)
	chows := s.claimChows(discarder, claimant, discarded, false)

	hasPong := len(matching) >= 3
	hasKong := len(matching) >= 4
	hasChow := len(chows) > 0

	for _, claimType := range domain.ClaimPriority {
		switch claimType {
		case domain.ClaimHu:
			// only allow if discarded tile makes a Hu
			player.AddTile(discarded)
			if !domain.IsSimpleWinningHand(player) {
				player.RemoveFromHand(discarded.ID)
				continue
			}
			player.RemoveFromHand(discarded.ID)

			if hasPong {
				res = append(res, s.generateClaimDecl(discarder, claimant, matching[:3], claimType, domain.MeldPong, domain.KongNone))
			}
			if hasKong {
				res = append(res, s.generateClaimDecl(discarder, claimant, matching[:4], claimType, domain.MeldKong, domain.KongExposed))
			}
			if hasChow {
				res = append(res, chows...)
			}
		case domain.ClaimPong:
			if hasPong {
				res = append(res, s.generateClaimDecl(discarder, claimant, matching[:3], claimType, domain.MeldPong, domain.KongNone))
			}
		case domain.ClaimKong:
			if hasKong {
				res = append(res, s.generateClaimDecl(discarder, claimant, matching[:4], claimType, domain.MeldKong, domain.KongExposed))
			}
		case domain.ClaimChow:
			if hasChow {
				res = append(res, chows...)
			}
		}
	}

	return res
}

func (s *ClaimState) claimChows(discarder, claimant domain.SeatIndex, discarded *domain.Tile, isHu bool) []*domain.ClaimDecl {
	var res []*domain.ClaimDecl
	// Chow is only valid on suited tiles (Wan/Bamboo/Dot).
	if !discarded.IsSuitedForChow() {
		return res
	}
	// Chow is only valid from the seat immediately after the
	// discarder if not Hu. Other seats can Pong or Kong the same tile, but
	// not Chow it.
	if !isHu && utils.SeatDistance(discarder, claimant) != 1 {
		return res
	}

	claimType := domain.ClaimChow
	if isHu {
		claimType = domain.ClaimHu
	}
	player := s.game.State.Players[claimant]
	chowTiles := s.getAllPossibleChowCombinations(player, discarded)
	for _, tiles := range chowTiles {
		res = append(res, s.generateClaimDecl(discarder, claimant, tiles, claimType, domain.MeldChow, domain.KongNone))
	}

	return res
}

func (s *ClaimState) generateClaimDecl(
	discarder,
	claimantIndex domain.SeatIndex,
	tiles []*domain.Tile,
	claimType domain.ClaimType,
	meldType domain.MeldType,
	kongType domain.KongType,
) *domain.ClaimDecl {
	claimDecl := &domain.ClaimDecl{
		Discarder: discarder,
		Claimant:  claimantIndex,
		ClaimType: claimType,
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
