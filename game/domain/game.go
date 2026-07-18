package domain

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

var (
	ErrRoundNotInPlay    = errors.New("round is not in play")
	ErrRoundNotInDiscard = errors.New("round is not in discard")
	ErrMustDraw          = errors.New("player must draw before discarding")
	ErrMustDiscard       = errors.New("player must discard before drawing")
	ErrTileNotInHand     = errors.New("tile is not in the player's hand")
	ErrTileNotClaim      = errors.New("tile is not claimable for the requested meld")
	ErrNotClaimSeat      = errors.New("it is not this seat's turn to declare or pass")
	ErrAlreadyActed      = errors.New("seat has already declared or passed in this claim window")
	ErrClaimClosed       = errors.New("claim window is not open")
	ErrWallEmpty         = errors.New("wall is empty")
	ErrCheckWinClosed    = errors.New("not the time to check win")
)

type Game struct {
	ID     string
	Config *GameConfig
	State  *GameState
}

// NewGame creates and deals a new round using a deterministic shuffle seed.
func NewGame(id string, config *GameConfig, seed uint64) *Game {
	wallTiles := append([]*Tile(nil), Catalog...)
	game := &Game{
		ID:     id,
		Config: config,
		State: &GameState{
			Round: Round{
				Phase:         PhaseDealing,
				CurrentPlayer: SeatIndex(West),
				LastDiscardBy: SeatNone,
			},
			Players: [4]*PlayerState{
				{
					ID: "1",
				},
				{
					ID: "2",
				},
				{
					ID: "3",
				},
				{
					ID: "4",
				},
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

	tile, ok := g.State.drawHandTile(g.State.Players[round.CurrentPlayer])
	if !ok {
		return ErrWallEmpty
	}
	round.NewlyDrawnTile = tile
	round.Phase = PhaseCheckWin
	return nil
}

func (g *Game) ResolveWin() error {
	round := &g.State.Round
	if round.Phase != PhaseCheckWin {
		return ErrCheckWinClosed
	}
	if IsSimpleWinningHand(g.State.Players[round.CurrentPlayer]) {
		round.Phase = PhaseWin
		return nil
	}
	round.Phase = PhaseDiscard
	return nil
}

// Discard removes a tile from the current player's hand and opens a
// ClaimWindow. The current player is the discarder; the discarder does not
// advance. Resolution of the ClaimWindow drives the next CurrentPlayer:
//
//   - Kong: claimant draws a replacement, then discards.
//   - Pong/Chow: claimant does not draw, discards immediately.
//   - All pass: next seat after the discarder becomes CurrentPlayer and
//     must draw.
func (g *Game) Discard(id TileID) error {
	round := &g.State.Round
	if round.Phase != PhaseDiscard {
		return ErrRoundNotInDiscard
	}
	// if round.NewlyDrawnTile == nil {
	// 	return ErrMustDraw
	// }

	player := round.CurrentPlayer
	tile, ok := g.State.Players[player].RemoveFromHand(id)
	if !ok {
		return ErrTileNotInHand
	}
	if tile.IsBonus() {
		// Defensive: a bonus tile should never be in the hand, but if a
		// caller manages to construct one, refuse to discard it.
		g.State.Players[player].Hand = append(g.State.Players[player].Hand, tile)
		return ErrTileNotInHand
	}

	g.State.Players[player].Discards = append(g.State.Players[player].Discards, tile)
	round.LastDiscard = tile
	round.LastDiscardBy = player
	round.NewlyDrawnTile = nil

	// Open the claim window. The discarder does not advance here; the
	// claim flow or the all-pass path drives CurrentPlayer. The discarder
	// is pre-marked as acted since they never get a claim turn.
	acted := [4]bool{}
	acted[player] = true
	window := &ClaimWindow{
		Discard:     tile,
		FromSeat:    player,
		PendingSeat: nextSeatAfter(player),
		Acted:       acted,
	}
	windowV2 := g.generateClaimWindowV2(round)
	round.Claim = window
	round.ClaimV2 = windowV2
	round.Phase = PhaseClaim
	return nil
}

func (g *Game) generateClaimWindowV2(round *Round) *ClaimWindowV2 {
	claimWindowV2 := &ClaimWindowV2{
		Discard:                 round.LastDiscard,
		FromSeat:                round.LastDiscardBy,
		AllPossibleDeclarations: g.getAllEligibleClaims(g.State, round),
		DeadlineUnix:            time.Now().Add(5 * time.Second).Unix(),
	}
	return claimWindowV2
}

func (g *Game) getAllEligibleClaims(state *GameState, round *Round) [][]*ClaimDecl {

	res := make([][]*ClaimDecl, 4)

	lastDiscardBy := round.LastDiscardBy
	lastDiscard := round.LastDiscard
	for playerIndex := range state.Players {
		if SeatIndex(playerIndex) == lastDiscardBy {
			res[playerIndex] = []*ClaimDecl{}
			continue
		}
		currPlayerPossibleClaims := g.generateAllPossibleClaimsForPlayer(lastDiscardBy, SeatIndex(playerIndex), lastDiscard)
		res[playerIndex] = currPlayerPossibleClaims
	}

	return res
}

func (g *Game) convertTilesToClaimDecl(discarder, claimantIndex SeatIndex, tiles []*Tile, meldType MeldType, kongType KongType) *ClaimDecl {
	claimDecl := &ClaimDecl{
		Discarder: discarder,
		Claimant:  claimantIndex,
		Type:      meldType,
		Kong:      kongType,
		Tiles:     tiles,
	}

	return claimDecl
}

func (g *Game) generateAllPossibleClaimsForPlayer(discarder, playerIndex SeatIndex, discard *Tile) []*ClaimDecl {
	player := g.State.Players[playerIndex]
	res := []*ClaimDecl{}

	for _, meldType := range []MeldType{MeldPong, MeldKong, MeldChow} {
		switch meldType {
		case MeldPong:
			tiles := getMatchingInHand(player, discard)
			if len(tiles) >= 3 {
				res = append(res, g.convertTilesToClaimDecl(discarder, playerIndex, tiles[:3], MeldPong, KongNone))
			}
		case MeldKong:
			tiles := getMatchingInHand(player, discard)
			if len(tiles) >= 4 {
				res = append(res, g.convertTilesToClaimDecl(discarder, playerIndex, tiles[:4], MeldKong, KongExposed))
			}
		case MeldChow:
			// Chow is only valid on suited tiles (Wan/Bamboo/Dot).
			if !discard.IsSuitedForChow() {
				continue
			}
			// Chow is only valid from the seat immediately after the
			// discarder. Other seats can Pong or Kong the same tile, but
			// not Chow it.
			if seatDistance(discarder, playerIndex) != 1 {
				continue
			}
			chowTiles := getAllPossibleChowCombinations(player, discard)
			for _, tiles := range chowTiles {
				res = append(res, g.convertTilesToClaimDecl(discarder, playerIndex, tiles, MeldChow, KongNone))
			}
		}
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Type > res[j].Type
	})

	return res
}

// getAllPossibleChowCombinations finds all possible 3-tile sequences (Chows) that can be
// formed using an opponent's discard and two tiles from your hand.
func getAllPossibleChowCombinations(p *PlayerState, discard *Tile) [][]*Tile {
	if discard == nil {
		return nil
	}

	var results [][]*Tile

	// TODO: optimize
	// Helper to find all tiles in hand matching a specific suit and rank
	findTiles := func(suit Suit, rank uint8) *Tile {
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
	if discard.Rank >= 3 {
		t1, t2 := findTiles(discard.Suit, discard.Rank-2), findTiles(discard.Suit, discard.Rank-1)
		if t1 != nil && t2 != nil {
			results = append(results, []*Tile{t1, t2, discard})
		}
	}
	// Case B: Discard is Middle (D-1, D, D+1)
	if discard.Rank >= 2 && discard.Rank <= 8 {
		t1, t3 := findTiles(discard.Suit, discard.Rank-1), findTiles(discard.Suit, discard.Rank+1)
		if t1 != nil && t3 != nil {
			results = append(results, []*Tile{t1, discard, t3})
		}
	}
	// Case C: Discard is Low (D, D+1, D+2)
	if discard.Rank <= 7 {
		t2, t3 := findTiles(discard.Suit, discard.Rank+1), findTiles(discard.Suit, discard.Rank+2)
		if t2 != nil && t3 != nil {
			results = append(results, []*Tile{discard, t2, t3})
		}
	}

	return results
}

func (g *Game) ResolveClaim() error {
	claimDecl := g.State.Round.ClaimV2.Winner
	round := &g.State.Round
	if round.Phase != PhaseClaim {
		return ErrClaimClosed
	}
	round.Phase = PhaseCheckWin
	round.Claim = nil
	round.ClaimV2 = nil

	if claimDecl == nil {
		round.CurrentPlayer = nextSeatAfter(round.CurrentPlayer)
		round.Phase = PhasePlay
		return nil
	}
	round.CurrentPlayer = claimDecl.Claimant

	claimant := g.State.Players[claimDecl.Claimant]
	for _, tile := range claimDecl.Tiles {
		if tile.ID == round.LastDiscard.ID {
			continue
		}
		if _, ok := claimant.RemoveFromHand(tile.ID); !ok {
			fmt.Printf("!!!!!!! Can not find tile rank: %d, suit: %d from player: %d !!!!!!!\n", tile.Rank, tile.Suit, claimDecl.Claimant)
		}
	}
	claimant.AddToMelds(claimDecl)

	if claimDecl.Type == MeldKong {
		tile, ok := g.State.drawHandTile(claimant)
		if !ok {
			return ErrWallEmpty
		}
		round.NewlyDrawnTile = tile
	}

	return nil
}

// ActionClaim processes a ClaimAction from the seat whose turn it is in
// the open ClaimWindow. On Pass, the seat is recorded as acted and
// PendingSeat advances. On Declare, the seat's declaration is validated
// and appended. Once all four seats have acted, Resolve runs and the
// window's winner (if any) is applied to the game state.
//
// Caller is expected to drive seat-by-seat (one action per seat, in turn
// order starting at the seat immediately after the discarder). The
// engine does not advance the turn until the window resolves.
// DEPRECATED
func (g *Game) ActionClaim(seat SeatIndex, action ClaimAction) error {
	round := &g.State.Round
	if round.Phase != PhaseClaim {
		return ErrClaimClosed
	}
	window := round.Claim
	if window == nil {
		return ErrClaimClosed
	}
	if seat != window.PendingSeat {
		return ErrNotClaimSeat
	}
	if window.Acted[seat] {
		return ErrAlreadyActed
	}
	window.Acted[seat] = true

	if action.Type == ClaimPass {
		window.PendingSeat = nextSeatAfter(seat)
		if window.AllSeatsActed() {
			return g.resolveClaimWindow()
		}
		return nil
	}

	// ClaimDeclare: validate and append.
	player := g.State.Players[seat]
	if err := validateClaim(seat, window.FromSeat, player, window.Discard, action.Meld, action.Kong); err != nil {
		// Roll back the Acted flag so the seat can retry with a valid
		// declaration? No — the caller asked to declare and failed
		// validation, so the seat has effectively passed. This matches
		// the "you only get one shot" rule for claim windows.
		window.PendingSeat = nextSeatAfter(seat)
		if window.AllSeatsActed() {
			g.resolveClaimWindow()
		}
		return err
	}

	window.Declarations = append(window.Declarations, &ClaimDecl{
		Claimant: seat,
		Type:     action.Meld,
		Kong:     action.Kong,
	})
	window.PendingSeat = nextSeatAfter(seat)
	if window.AllSeatsActed() {
		return g.resolveClaimWindow()
	}
	return nil
}

// resolveClaimWindow applies the winning claim (if any) to game state and
// closes the window. Called by ActionClaim when the last seat has acted.
func (g *Game) resolveClaimWindow() error {
	round := &g.State.Round
	window := round.Claim
	winner := window.Resolve()
	window.Winner = winner
	window.Resolved = true
	round.Claim = nil
	round.Phase = PhaseDiscard

	if winner == nil {
		// Everyone passed. The discarder does not get a draw, so the
		// turn advances to the next seat, which must draw.
		round.CurrentPlayer = nextSeatAfter(window.FromSeat)
		return nil
	}

	// Apply the winning claim.
	claimant := g.State.Players[winner.Claimant]
	discarder := g.State.Players[window.FromSeat]
	meld := buildMeld(claimant, discarder, window.FromSeat, window.Discard, winner.Type, winner.Kong)
	claimant.Melds = append(claimant.Melds, meld)
	round.CurrentPlayer = winner.Claimant

	// Pong/Chow: no draw. The claimant must discard.
	// Kong: claimant draws a replacement tile.
	if winner.Type == MeldKong {
		tile, ok := g.State.drawHandTile(claimant)
		if !ok {
			return ErrWallEmpty
		}
		round.NewlyDrawnTile = tile
	}
	return nil
}

// buildMeld constructs the Meld from the claim and removes the consumed
// hand tiles from the claimant. The discard tile is moved from the
// discarder's discards into the meld. discarderSeat is needed to fill
// Meld.FromSeat.
func buildMeld(claimant, discarder *PlayerState, discarderSeat SeatIndex, discard *Tile, meldType MeldType, kong KongType) *Meld {
	// Pull the discard out of the discarder's discards.
	discarder.removeFromDiscards(discard.ID)

	switch meldType {
	case MeldPong:
		// Two matching tiles from hand + the discard.
		removed := claimant.removeMatchingFromHand(discard, 2)
		return &Meld{
			Type:     MeldPong,
			Tiles:    tilesToValues(append([]*Tile{discard}, removed...)),
			FromSeat: discarderSeat,
		}
	case MeldKong:
		// Three matching tiles from hand + the discard.
		removed := claimant.removeMatchingFromHand(discard, 3)
		return &Meld{
			Type:     MeldKong,
			Kong:     kong,
			Tiles:    tilesToValues(append([]*Tile{discard}, removed...)),
			FromSeat: discarderSeat,
		}
	case MeldChow:
		// Two sequence tiles from hand + the discard.
		removed := claimant.removeChowPartnersFromHand(discard, 2)
		return &Meld{
			Type:     MeldChow,
			Tiles:    tilesToValues(append([]*Tile{discard}, removed...)),
			FromSeat: discarderSeat,
		}
	}
	return nil
}

// validateClaim checks that a seat can legally declare the requested
// claim type against the discard tile, given the tiles in the seat's
// hand.
func validateClaim(seat, discarder SeatIndex, p *PlayerState, discard *Tile, meldType MeldType, kong KongType) error {
	if discard == nil {
		return ErrTileNotInHand
	}
	if discard.IsBonus() {
		// Bonus tiles should never be the discard, but guard anyway.
		return ErrTileNotClaim
	}
	switch meldType {
	case MeldPong:
		if countMatchingInHand(p, discard) < 2 {
			return ErrTileNotClaim
		}
	case MeldKong:
		if kong != KongExposed {
			return ErrTileNotClaim
		}
		if countMatchingInHand(p, discard) < 3 {
			return ErrTileNotClaim
		}
	case MeldChow:
		// Chow is only valid on suited tiles (Wan/Bamboo/Dot).
		if !discard.IsSuitedForChow() {
			return ErrTileNotClaim
		}
		// Chow is only valid from the seat immediately after the
		// discarder. Other seats can Pong or Kong the same tile, but
		// not Chow it.
		if seatDistance(discarder, seat) != 1 {
			return ErrTileNotClaim
		}
		// Need two sequence-partner tiles in hand.
		if !canCompleteChow(p, discard) {
			return ErrTileNotClaim
		}
	}
	return nil
}

func getMatchingInHand(p *PlayerState, t *Tile) []*Tile {
	res := []*Tile{t}
	for _, h := range p.Hand {
		if sameKind(h, t) {
			res = append(res, h)
		}
	}
	return res
}

func countMatchingInHand(p *PlayerState, t *Tile) int {
	n := 0
	for _, h := range p.Hand {
		if sameKind(h, t) {
			n++
		}
	}
	return n
}

// sameKind reports whether two tiles match for triplet/pong purposes:
// same suit, same rank, regardless of physical TileID.
func sameKind(a, b *Tile) bool {
	if a == nil || b == nil {
		return false
	}
	return a.Suit == b.Suit && a.Rank == b.Rank
}

// canCompleteChow checks whether the hand has tiles to complete a Chow
// with the given discard as one of the three.
func canCompleteChow(p *PlayerState, discard *Tile) bool {
	if discard == nil {
		return false
	}
	for _, pos := range []int{-1, 0, +1} {
		if chowHelper(p, discard, pos) {
			return true
		}
	}
	return false
}

// chowHelper checks whether, treating `discard` as the tile at offset
// `pos` (-1, 0, or +1) within a 3-tile sequence, the hand contains the
// other two sequence tiles.
//
// pos = -1: discard is the low end of the sequence. The sequence is
//
//	discard.Rank, discard.Rank+1, discard.Rank+2. We need the
//	+1 and +2 ranks in hand. Valid only when discard.Rank <= 7.
//
// pos = 0: discard is the middle of the sequence. The sequence is
//
//	discard.Rank-1, discard.Rank, discard.Rank+1. We need the
//	-1 and +1 ranks in hand. Valid only when 2 <= discard.Rank <= 8.
//
// pos = +1: discard is the high end of the sequence. The sequence is
//
//	discard.Rank-2, discard.Rank-1, discard.Rank. We need the
//	-2 and -1 ranks in hand. Valid only when discard.Rank >= 3.
func chowHelper(p *PlayerState, discard *Tile, pos int) bool {
	if discard == nil {
		return false
	}
	var otherA, otherB uint8
	switch pos {
	case -1:
		if discard.Rank > 7 {
			return false
		}
		otherA, otherB = discard.Rank+1, discard.Rank+2
	case 0:
		if discard.Rank < 2 || discard.Rank > 8 {
			return false
		}
		otherA, otherB = discard.Rank-1, discard.Rank+1
	case +1:
		if discard.Rank < 3 {
			return false
		}
		otherA, otherB = discard.Rank-2, discard.Rank-1
	default:
		return false
	}
	haveA, haveB := false, false
	for _, h := range p.Hand {
		if h.Suit != discard.Suit {
			continue
		}
		switch h.Rank {
		case otherA:
			haveA = true
		case otherB:
			haveB = true
		}
	}
	return haveA && haveB
}

// removeMatchingFromHand removes up to `n` tiles matching the given tile
// from the hand and returns them. Used to build Pong and Kong melds.
func (p *PlayerState) removeMatchingFromHand(t *Tile, n int) []*Tile {
	out := make([]*Tile, 0, n)
	kept := p.Hand[:0]
	for _, h := range p.Hand {
		if len(out) < n && sameKind(h, t) {
			out = append(out, h)
			continue
		}
		kept = append(kept, h)
	}
	p.Hand = append(p.Hand[:0], kept...)
	return out
}

// removeChowPartnersFromHand removes the two sequence-partner tiles
// needed to complete a Chow with the given discard. We pick the lowest
// matching sequence; if multiple are legal we take the first one found.
func (p *PlayerState) removeChowPartnersFromHand(discard *Tile, n int) []*Tile {
	// TODO: make the player decide which tiles to declare if has multiple options
	// Try each legal position (-1, 0, +1) in order; first one that has
	// both partners in hand wins.
	for _, pos := range []int{-1, 0, +1} {
		if !chowHelper(p, discard, pos) {
			continue
		}
		var otherA, otherB uint8
		switch pos {
		case -1:
			otherA, otherB = discard.Rank+1, discard.Rank+2
		case 0:
			otherA, otherB = discard.Rank-1, discard.Rank+1
		case +1:
			otherA, otherB = discard.Rank-2, discard.Rank-1
		}
		// Remove one of each.
		out := make([]*Tile, 0, n)
		kept := p.Hand[:0]
		gotA, gotB := false, false
		for _, h := range p.Hand {
			if !gotA && h.Suit == discard.Suit && h.Rank == otherA {
				out = append(out, h)
				gotA = true
				continue
			}
			if !gotB && h.Suit == discard.Suit && h.Rank == otherB {
				out = append(out, h)
				gotB = true
				continue
			}
			kept = append(kept, h)
		}
		p.Hand = append(p.Hand[:0], kept...)
		return out
	}
	return nil
}

// removeFromDiscards removes a tile with the given ID from a player's
// discard pile. Used to take the discarded tile back when a claim wins.
func (p *PlayerState) removeFromDiscards(id TileID) bool {
	for i, t := range p.Discards {
		if t.ID == id {
			p.Discards = append(p.Discards[:i], p.Discards[i+1:]...)
			return true
		}
	}
	return false
}

// DeclareConcealedKong is a self-action during PhasePlay on the active
// player's turn. It looks for a 4-of-a-kind in the active player's hand
// (using the most recently drawn tile plus three matching hand tiles) and
// forms a concealed Kong. After declaration, the player draws a
// replacement tile. The player can call this method repeatedly while
// successive drawn tiles complete further Kongs; once the draw no
// longer completes a Kong, the player should discard.
func (g *Game) DeclareConcealedKong() error {
	round := &g.State.Round
	if round.Phase != PhasePlay {
		return ErrRoundNotInPlay
	}
	if round.NewlyDrawnTile == nil {
		return ErrMustDraw
	}

	player := g.State.Players[round.CurrentPlayer]
	drawn := round.NewlyDrawnTile
	if drawn.IsBonus() {
		return ErrMustDraw // can't Kong a bonus tile (it shouldn't be in hand anyway)
	}

	// Need 3 matching hand tiles to complete a 4-of-a-kind.
	if countMatchingInHandExcluding(player, drawn, drawn) < 3 {
		return ErrTileNotClaim
	}

	// Build the Kong: drawn tile + 3 matching from hand.
	removed := player.removeMatchingFromHandExcluding(drawn, drawn, 3)
	meld := &Meld{
		Type:     MeldKong,
		Kong:     KongConcealed,
		Tiles:    tilesToValues(append([]*Tile{drawn}, removed...)),
		FromSeat: round.CurrentPlayer,
	}
	player.Melds = append(player.Melds, meld)
	round.NewlyDrawnTile = nil

	// Draw a replacement (with bonus-tile replacement).
	tile, ok := g.State.drawHandTile(player)
	if !ok {
		return ErrWallEmpty
	}
	round.NewlyDrawnTile = tile
	return nil
}

// tilesToValues converts a slice of *Tile to a slice of Tile by copying
// the underlying value. Meld.Tiles is []Tile (value) for compatibility
// with the rest of the domain types; pointers are easier to work with
// elsewhere in the engine.
func tilesToValues(pts []*Tile) []*Tile {
	out := make([]*Tile, len(pts))
	for i, t := range pts {
		if t != nil {
			out[i] = t
		}
	}
	return out
}

// countMatchingInHandExcluding counts tiles in the hand that match
// `proto` but are not the same physical tile as `exclude`. Used for
// concealed-Kong validation, where `proto == exclude == the drawn tile`
// and we want to count *other* matching tiles in the hand.
func countMatchingInHandExcluding(p *PlayerState, proto, exclude *Tile) int {
	n := 0
	for _, h := range p.Hand {
		if h == exclude {
			continue
		}
		if sameKind(h, proto) {
			n++
		}
	}
	return n
}

func (p *PlayerState) removeMatchingFromHandExcluding(proto, exclude *Tile, n int) []*Tile {
	out := make([]*Tile, 0, n)
	kept := p.Hand[:0]
	for _, h := range p.Hand {
		if h == exclude {
			kept = append(kept, h)
			continue
		}
		if len(out) < n && sameKind(h, proto) {
			out = append(out, h)
			continue
		}
		kept = append(kept, h)
	}
	p.Hand = append(p.Hand[:0], kept...)
	return out
}

// UpgradePongToKong promotes an existing exposed Pong to an exposed Kong
// when the active player self-draws the fourth matching tile. The fourth
// tile is moved from the drawn hand tile into the existing Meld. The
// player then draws a replacement tile.
func (g *Game) UpgradePongToKong(meldIndex int) error {
	round := &g.State.Round
	if round.Phase != PhasePlay {
		return ErrRoundNotInPlay
	}
	if round.NewlyDrawnTile == nil {
		return ErrMustDraw
	}

	player := g.State.Players[round.CurrentPlayer]
	if meldIndex < 0 || meldIndex >= len(player.Melds) {
		return ErrTileNotClaim
	}
	meld := player.Melds[meldIndex]
	if meld.Type != MeldPong {
		return ErrTileNotClaim
	}
	if meld.Kong != KongNone {
		return ErrTileNotClaim
	}

	drawn := round.NewlyDrawnTile
	if drawn.IsBonus() {
		return ErrMustDraw
	}
	// Drawn tile must match the Pong's tiles by suit+rank.
	if len(meld.Tiles) == 0 {
		return ErrTileNotClaim
	}
	first := meld.Tiles[0]
	if !sameKind(first, drawn) {
		return ErrTileNotClaim
	}

	meld.Type = MeldKong
	meld.Kong = KongExposed
	meld.Tiles = append(meld.Tiles, drawn)
	round.NewlyDrawnTile = nil

	tile, ok := g.State.drawHandTile(player)
	if !ok {
		return ErrWallEmpty
	}
	round.NewlyDrawnTile = tile
	return nil
}

func (g *Game) GetPhaseHandlers(tileIDToDiscard TileID) map[RoundPhase]func() error {
	return map[RoundPhase]func() error{
		PhasePlay: func() error {
			return g.DrawForCurrentPlayer()
		},
		PhaseDiscard: func() error {
			return g.Discard(tileIDToDiscard)
		},
		PhaseClaim: func() error {
			return g.ResolveClaim()
		},
		PhaseCheckWin: func() error {
			return g.ResolveWin()
		},
	}
}
