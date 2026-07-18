package domain

import (
	"errors"
	"testing"
)

// Phase 4 tests: claims, melds, priority, and the no-robbing-the-Kong rule.
//
// All tests construct a Game struct directly (rather than going through
// NewGame) so each test can lay out the exact hand it needs. The pattern
// is:
//
//   1. Pick a discarder seat and a target tile for the discard.
//   2. Set PhasePlay, CurrentPlayer=discarder, and pre-populate each
//      player's hand with the tiles we want to assert against.
//   3. Put a few extra tiles in the wall so draws don't fail.
//   4. Call Discard to open the ClaimWindow, then drive the seats with
//      ActionClaim.

const claimTestIDStart = 0 // TileIDs we use are the first physical copies

// claimTestHelper hands out a fresh *Game pre-loaded with the given hands
// and wall, and a discarder who has just opened a ClaimWindow.
//
// callerDiscards, if non-nil, is the *Tile the discarder is discarding.
// The discarder must own this tile in their hand before Discard is called.
// A fake "previously drawn" tile is set on the round so Discard's
// NewlyDrawnTile precondition is satisfied.
func claimTestHelper(t *testing.T, hands map[Wind][]*Tile, wallTiles []*Tile, caller Wind, callerDiscards *Tile) *Game {
	t.Helper()
	g := &Game{
		ID:     "claim-test",
		Config: &GameConfig{},
		State: &GameState{
			Round: Round{
				Phase:         PhasePlay,
				CurrentPlayer: SeatIndex(caller),
			},
		},
	}
	// Lay out the requested hands.
	for seat, hand := range hands {
		g.State.Players[seat].Hand = append([]*Tile(nil), hand...)
	}
	// Add the discard tile to the discarder's hand if it's not there yet.
	if callerDiscards != nil {
		g.State.Players[caller].Hand = append(g.State.Players[caller].Hand, callerDiscards)
	}
	// Set a placeholder "newly drawn" tile on the discarder so the
	// Discard precondition passes. We use a high-id tile (TileID 145,
	// Cat) — the discarder won't actually be holding a meaningful draw;
	// the test is about the claim window, not the draw flow.
	g.State.Round.NewlyDrawnTile = &Tile{ID: 145, Suit: SuitAnimal, Rank: 0}
	// Set up a wall with the requested tiles plus a small safety margin
	// so replacement draws never go empty.
	g.State.Wall.DrawPile = append([]*Tile(nil), wallTiles...)
	// Push a long tail of dummy normal tiles so any draw that needs a
	// "next" tile can succeed.
	for i := range 32 {
		id := TileID(64 + i) // dot suit, ranks 5.., won't collide with our test tiles
		g.State.Wall.DrawPile = append(g.State.Wall.DrawPile, &Tile{ID: id, Suit: SuitDot, Rank: 5})
	}
	return g
}

// wanTile returns a fresh pointer to the n-th physical copy of rank r
// Wan. r is 1..9, copy is 0..3.
func wanTile(r uint8, copy int) *Tile {
	id := TileID(int(SuitCharacter)*36 + int(r-1)*4 + copy)
	return &Tile{ID: id, Suit: SuitCharacter, Rank: r}
}

// bambooTile returns a fresh pointer to the n-th physical copy of rank r
// Bamboo.
func bambooTile(r uint8, copy int) *Tile {
	id := TileID(int(SuitBamboo)*36 + int(r-1)*4 + copy)
	return &Tile{ID: id, Suit: SuitBamboo, Rank: r}
}

// windTile returns a fresh pointer to the n-th physical copy of the
// given wind rank (0=East, 1=South, 2=West, 3=North).
func windTile(rank uint8, copy int) *Tile {
	id := TileID(firstWindID + int(rank)*4 + copy)
	return &Tile{ID: id, Suit: SuitWind, Rank: rank}
}

// ----------------------------------------------------------------------------
// Pong
// ----------------------------------------------------------------------------

func TestPongConsumesTwoHandTilesAndDiscard(t *testing.T) {
	// East discards 5-Wan. South has 5-Wan, 5-Wan in hand. South Pongs.
	discard := wanTile(5, 0)
	g := claimTestHelper(t, map[Wind][]*Tile{
		East:  {wanTile(1, 0), wanTile(2, 0), wanTile(3, 0), wanTile(4, 0)},
		South: {wanTile(5, 1), wanTile(5, 2), wanTile(9, 0), wanTile(9, 1)},
		West:  {wanTile(6, 0), wanTile(7, 0), wanTile(8, 0), wanTile(1, 1)},
		North: {wanTile(2, 1), wanTile(3, 1), wanTile(4, 1), wanTile(6, 1)},
	}, []*Tile{discard}, East, discard)

	if err := g.Discard(discard.ID); err != nil {
		t.Fatalf("discard: %v", err)
	}
	// South Pongs.
	if err := g.ActionClaim(SeatIndex(South), ClaimAction{Type: ClaimDeclare, Meld: MeldPong}); err != nil {
		t.Fatalf("South pong: %v", err)
	}
	// West passes (only South claimed; window should resolve immediately).
	if err := g.ActionClaim(SeatIndex(West), ClaimAction{Type: ClaimPass}); err != nil {
		t.Fatalf("West pass: %v", err)
	}
	// North passes.
	if err := g.ActionClaim(SeatIndex(North), ClaimAction{Type: ClaimPass}); err != nil {
		t.Fatalf("North pass: %v", err)
	}

	// After Pong: window closed, current player is South (the claimant),
	// no replacement draw, no newly-drawn tile.
	if g.State.Round.Phase != PhasePlay {
		t.Fatalf("phase = %v, want PhasePlay", g.State.Round.Phase)
	}
	if g.State.Round.Claim != nil {
		t.Fatal("claim window still open after pong")
	}
	if g.State.Round.CurrentPlayer != SeatIndex(South) {
		t.Fatalf("current player = %d, want South", g.State.Round.CurrentPlayer)
	}
	if g.State.Round.NewlyDrawnTile != nil {
		t.Fatal("Pong must not set a newly-drawn tile")
	}
	// South's hand should have lost the two 5-Wan; meld holds the 3.
	if got := len(g.State.Players[South].Hand); got != 2 {
		t.Fatalf("South hand size = %d, want 2", got)
	}
	for _, h := range g.State.Players[South].Hand {
		if h.Rank == 5 && h.Suit == SuitCharacter {
			t.Fatalf("South still holds a 5-Wan after pong: %v", h.Describe())
		}
	}
	if got := len(g.State.Players[South].Melds); got != 1 {
		t.Fatalf("South melds = %d, want 1", got)
	}
	meld := g.State.Players[South].Melds[0]
	if meld.Type != MeldPong {
		t.Fatalf("meld type = %v, want MeldPong", meld.Type)
	}
	if len(meld.Tiles) != 3 {
		t.Fatalf("meld tiles = %d, want 3", len(meld.Tiles))
	}
	for _, mt := range meld.Tiles {
		if !(mt.Suit == SuitCharacter && mt.Rank == 5) {
			t.Fatalf("meld contains non-5-Wan tile: %v", mt.Describe())
		}
	}
	// The discard must have been pulled from East's discard pile.
	if got := len(g.State.Players[East].Discards); got != 0 {
		t.Fatalf("East discards = %d, want 0 (the discard moved to the meld)", got)
	}
}

func TestPongRejectsInsufficientHandTiles(t *testing.T) {
	discard := wanTile(5, 0)
	g := claimTestHelper(t, map[Wind][]*Tile{
		East:  {wanTile(1, 0)},
		South: {wanTile(5, 1), wanTile(9, 0), wanTile(9, 1)},
	}, []*Tile{discard}, East, discard)
	if err := g.Discard(discard.ID); err != nil {
		t.Fatalf("discard: %v", err)
	}
	err := g.ActionClaim(SeatIndex(South), ClaimAction{Type: ClaimDeclare, Meld: MeldPong})
	if !errors.Is(err, ErrTileNotClaim) {
		t.Fatalf("pong with 1 in hand = %v, want ErrTileNotClaim", err)
	}
}

// ----------------------------------------------------------------------------
// Kong (exposed)
// ----------------------------------------------------------------------------

func TestKongConsumesThreeHandTilesAndDiscardsReplacement(t *testing.T) {
	discard := wanTile(7, 0)
	// Wall: pre-load a known replacement tile (1-Wan copy 0) so we can
	// assert the draw matches. The helper appends dummy tiles after.
	g := claimTestHelper(t, map[Wind][]*Tile{
		East:  {wanTile(1, 1), wanTile(2, 0), wanTile(3, 0)},
		South: {wanTile(7, 1), wanTile(7, 2), wanTile(7, 3), wanTile(9, 0)},
	}, []*Tile{wanTile(1, 0)}, East, discard)
	if err := g.Discard(discard.ID); err != nil {
		t.Fatalf("discard: %v", err)
	}
	if err := g.ActionClaim(SeatIndex(South), ClaimAction{Type: ClaimDeclare, Meld: MeldKong, Kong: KongExposed}); err != nil {
		t.Fatalf("South kong: %v", err)
	}
	// West and North pass to close the window.
	if err := g.ActionClaim(SeatIndex(West), ClaimAction{Type: ClaimPass}); err != nil {
		t.Fatalf("West pass: %v", err)
	}
	if err := g.ActionClaim(SeatIndex(North), ClaimAction{Type: ClaimPass}); err != nil {
		t.Fatalf("North pass: %v", err)
	}
	// After Kong: window closed; South is current; one tile was drawn
	// (the replacement). The replacement came from the wall, so
	// NewlyDrawnTile should be the 1-Wan copy 0 we pre-loaded.
	if g.State.Round.Phase != PhasePlay {
		t.Fatalf("phase = %v, want PhasePlay", g.State.Round.Phase)
	}
	if g.State.Round.Claim != nil {
		t.Fatal("claim window still open after kong")
	}
	if g.State.Round.CurrentPlayer != SeatIndex(South) {
		t.Fatalf("current player = %d, want South", g.State.Round.CurrentPlayer)
	}
	if g.State.Round.NewlyDrawnTile == nil {
		t.Fatal("Kong must trigger a replacement draw (NewlyDrawnTile is nil)")
	}
	if g.State.Round.NewlyDrawnTile.ID != wanTile(1, 0).ID {
		t.Fatalf("replacement = %v, want 1-Wan copy 0", g.State.Round.NewlyDrawnTile.Describe())
	}
	// South's hand should hold the 9-Wan and the 1-Wan replacement:
	// the three 7-Wans are in the Kong, and the 1-Wan draw was added.
	if got := len(g.State.Players[South].Hand); got != 2 {
		t.Fatalf("South hand size = %d, want 2 (9-Wan + replacement)", got)
	}
	if got := len(g.State.Players[South].Melds); got != 1 {
		t.Fatalf("South melds = %d, want 1", got)
	}
	meld := g.State.Players[South].Melds[0]
	if meld.Type != MeldKong {
		t.Fatalf("meld type = %v, want MeldKong", meld.Type)
	}
	if meld.Kong != KongExposed {
		t.Fatalf("meld kong = %v, want KongExposed", meld.Kong)
	}
	if len(meld.Tiles) != 4 {
		t.Fatalf("meld tiles = %d, want 4", len(meld.Tiles))
	}
	for _, mt := range meld.Tiles {
		if !(mt.Suit == SuitCharacter && mt.Rank == 7) {
			t.Fatalf("meld contains non-7-Wan tile: %v", mt.Describe())
		}
	}
}

func TestKongRejectsInsufficientHandTiles(t *testing.T) {
	discard := wanTile(7, 0)
	g := claimTestHelper(t, map[Wind][]*Tile{
		East:  {wanTile(1, 0)},
		South: {wanTile(7, 1), wanTile(7, 2), wanTile(9, 0)}, // only 2 matching
	}, []*Tile{discard}, East, discard)
	if err := g.Discard(discard.ID); err != nil {
		t.Fatalf("discard: %v", err)
	}
	err := g.ActionClaim(SeatIndex(South), ClaimAction{Type: ClaimDeclare, Meld: MeldKong, Kong: KongExposed})
	if !errors.Is(err, ErrTileNotClaim) {
		t.Fatalf("kong with 2 in hand = %v, want ErrTileNotClaim", err)
	}
}

func TestKongRejectsConcealedOrAddedSubtypeOnClaim(t *testing.T) {
	// Only KongExposed is legal on a discard claim. KongConcealed and
	// KongAdded are self-actions during PhasePlay and must be rejected
	// here.
	discard := wanTile(7, 0)
	g := claimTestHelper(t, map[Wind][]*Tile{
		East:  {wanTile(1, 0)},
		South: {wanTile(7, 1), wanTile(7, 2), wanTile(7, 3), wanTile(9, 0)},
	}, []*Tile{discard}, East, discard)
	if err := g.Discard(discard.ID); err != nil {
		t.Fatalf("discard: %v", err)
	}
	err := g.ActionClaim(SeatIndex(South), ClaimAction{Type: ClaimDeclare, Meld: MeldKong, Kong: KongConcealed})
	if !errors.Is(err, ErrTileNotClaim) {
		t.Fatalf("kong with KongConcealed = %v, want ErrTileNotClaim", err)
	}
}

// ----------------------------------------------------------------------------
// Chow
// ----------------------------------------------------------------------------

func TestChowLegalForNextPlayer(t *testing.T) {
	// East discards 5-Wan. South is next player. South has 4-Wan and
	// 6-Wan in hand. South chows 4-5-6.
	discard := wanTile(5, 0)
	g := claimTestHelper(t, map[Wind][]*Tile{
		East:  {wanTile(1, 0), wanTile(2, 0), wanTile(3, 0)},
		South: {wanTile(4, 0), wanTile(6, 0), wanTile(9, 0), wanTile(9, 1)},
	}, []*Tile{discard}, East, discard)
	if err := g.Discard(discard.ID); err != nil {
		t.Fatalf("discard: %v", err)
	}
	if err := g.ActionClaim(SeatIndex(South), ClaimAction{Type: ClaimDeclare, Meld: MeldChow}); err != nil {
		t.Fatalf("South chow: %v", err)
	}
	if err := g.ActionClaim(SeatIndex(West), ClaimAction{Type: ClaimPass}); err != nil {
		t.Fatalf("West pass: %v", err)
	}
	if err := g.ActionClaim(SeatIndex(North), ClaimAction{Type: ClaimPass}); err != nil {
		t.Fatalf("North pass: %v", err)
	}
	if g.State.Round.CurrentPlayer != SeatIndex(South) {
		t.Fatalf("current player = %d, want South", g.State.Round.CurrentPlayer)
	}
	if got := len(g.State.Players[South].Melds); got != 1 {
		t.Fatalf("South melds = %d, want 1", got)
	}
	meld := g.State.Players[South].Melds[0]
	if meld.Type != MeldChow {
		t.Fatalf("meld type = %v, want MeldChow", meld.Type)
	}
	if len(meld.Tiles) != 3 {
		t.Fatalf("meld tiles = %d, want 3", len(meld.Tiles))
	}
}

func TestChowRejectedForNonNextPlayer(t *testing.T) {
	// East discards 5-Wan. West is two seats away from East; West must
	// not be able to Chow even if West has 4-Wan and 6-Wan.
	discard := wanTile(5, 0)
	g := claimTestHelper(t, map[Wind][]*Tile{
		East:  {wanTile(1, 0), wanTile(2, 0), wanTile(3, 0)},
		South: {wanTile(1, 1), wanTile(2, 1), wanTile(3, 1)},
		West:  {wanTile(4, 0), wanTile(6, 0), wanTile(9, 0), wanTile(9, 1)},
		North: {wanTile(4, 1), wanTile(6, 1), wanTile(7, 0), wanTile(8, 0)},
	}, []*Tile{discard}, East, discard)
	if err := g.Discard(discard.ID); err != nil {
		t.Fatalf("discard: %v", err)
	}
	// West is not the PendingSeat on the first turn (South is). Let
	// South pass first.
	if err := g.ActionClaim(SeatIndex(South), ClaimAction{Type: ClaimPass}); err != nil {
		t.Fatalf("South pass: %v", err)
	}
	// Now West's turn. West tries to Chow; should be rejected because
	// West is not the next-after-East seat.
	err := g.ActionClaim(SeatIndex(West), ClaimAction{Type: ClaimDeclare, Meld: MeldChow})
	if !errors.Is(err, ErrTileNotClaim) {
		t.Fatalf("West chow = %v, want ErrTileNotClaim", err)
	}
}

func TestChowRejectedOnHonorTile(t *testing.T) {
	// East discards East wind. South cannot Chow (winds are not suited
	// for sequences).
	discard := windTile(0, 0) // East
	g := claimTestHelper(t, map[Wind][]*Tile{
		East:  {wanTile(1, 0), wanTile(2, 0), wanTile(3, 0)},
		South: {windTile(0, 1), windTile(0, 2), wanTile(9, 0), wanTile(9, 1)},
	}, []*Tile{discard}, East, discard)
	if err := g.Discard(discard.ID); err != nil {
		t.Fatalf("discard: %v", err)
	}
	err := g.ActionClaim(SeatIndex(South), ClaimAction{Type: ClaimDeclare, Meld: MeldChow})
	if !errors.Is(err, ErrTileNotClaim) {
		t.Fatalf("chow on wind = %v, want ErrTileNotClaim", err)
	}
}

func TestChowRejectedWithoutSequencePartners(t *testing.T) {
	discard := wanTile(5, 0)
	g := claimTestHelper(t, map[Wind][]*Tile{
		East:  {wanTile(1, 0)},
		South: {wanTile(1, 1), wanTile(2, 1), wanTile(3, 1), wanTile(9, 0)},
	}, []*Tile{discard}, East, discard)
	if err := g.Discard(discard.ID); err != nil {
		t.Fatalf("discard: %v", err)
	}
	err := g.ActionClaim(SeatIndex(South), ClaimAction{Type: ClaimDeclare, Meld: MeldChow})
	if !errors.Is(err, ErrTileNotClaim) {
		t.Fatalf("chow without partners = %v, want ErrTileNotClaim", err)
	}
}

// ----------------------------------------------------------------------------
// Priority and tiebreak
// ----------------------------------------------------------------------------

func TestClaimWindowResolvePrefersKongOverPongOverChow(t *testing.T) {
	// Pure unit test on ClaimWindow.Resolve. A real discard can't have
	// both a Kong and a Pong declared (4 copies per rank, 1 used as
	// discard), but the priority rule is independent of that and is
	// what we want to lock in here.
	window := &ClaimWindow{
		FromSeat: SeatIndex(East),
		Declarations: []*ClaimDecl{
			{Claimant: SeatIndex(South), Type: MeldChow},
			{Claimant: SeatIndex(West), Type: MeldPong},
			{Claimant: SeatIndex(North), Type: MeldKong, Kong: KongExposed},
		},
	}
	winner := window.Resolve()
	if winner == nil {
		t.Fatal("Resolve returned nil")
	}
	if winner.Claimant != SeatIndex(North) {
		t.Fatalf("winner = %d, want North (kong)", winner.Claimant)
	}
	if winner.Type != MeldKong {
		t.Fatalf("winner type = %v, want MeldKong", winner.Type)
	}
}

func TestClaimWindowResolvePongOverChow(t *testing.T) {
	window := &ClaimWindow{
		FromSeat: SeatIndex(East),
		Declarations: []*ClaimDecl{
			{Claimant: SeatIndex(South), Type: MeldChow},
			{Claimant: SeatIndex(West), Type: MeldPong},
		},
	}
	winner := window.Resolve()
	if winner == nil || winner.Claimant != SeatIndex(West) {
		t.Fatalf("winner = %v, want West (pong)", winner)
	}
}

func TestClaimWindowResolveTiebreakByClosestNext(t *testing.T) {
	// Same type (Pong), different seats. South is closer to East.
	window := &ClaimWindow{
		FromSeat: SeatIndex(East),
		Declarations: []*ClaimDecl{
			{Claimant: SeatIndex(West), Type: MeldPong},
			{Claimant: SeatIndex(South), Type: MeldPong},
		},
	}
	winner := window.Resolve()
	if winner == nil || winner.Claimant != SeatIndex(South) {
		t.Fatalf("winner = %v, want South (closest-next on pong tie)", winner)
	}
}

func TestClaimWindowResolveInsertionOrderOnFullTie(t *testing.T) {
	// Same type, same distance (impossible in real play with a single
	// discard but tests the deterministic tiebreak). Earlier-declared
	// wins.
	window := &ClaimWindow{
		FromSeat: SeatIndex(East),
		Declarations: []*ClaimDecl{
			{Claimant: SeatIndex(South), Type: MeldChow},
			{Claimant: SeatIndex(South), Type: MeldChow}, // duplicate
		},
	}
	winner := window.Resolve()
	if winner == nil || winner.Claimant != SeatIndex(South) {
		t.Fatalf("winner = %v, want first-declared South", winner)
	}
}

func TestClaimWindowResolveEmpty(t *testing.T) {
	window := &ClaimWindow{FromSeat: SeatIndex(East)}
	if window.Resolve() != nil {
		t.Fatal("empty window must resolve to nil")
	}
}

func TestPongBeatsChow(t *testing.T) {
	// East discards 5-Wan. South (next) can Chow with 4+6. West can
	// Pong with two 5-Wans. Priority: West (Pong) wins over Chow.
	discard := wanTile(5, 0)
	g := claimTestHelper(t, map[Wind][]*Tile{
		East:  {wanTile(1, 0), wanTile(2, 0), wanTile(3, 0)},
		South: {wanTile(4, 0), wanTile(6, 0), wanTile(9, 0), wanTile(9, 1)},
		West:  {wanTile(5, 1), wanTile(5, 2), wanTile(7, 0), wanTile(8, 0)},
		North: {wanTile(1, 1), wanTile(2, 1), wanTile(3, 1), wanTile(4, 1)},
	}, []*Tile{discard}, East, discard)
	if err := g.Discard(discard.ID); err != nil {
		t.Fatalf("discard: %v", err)
	}
	// South declares Chow.
	if err := g.ActionClaim(SeatIndex(South), ClaimAction{Type: ClaimDeclare, Meld: MeldChow}); err != nil {
		t.Fatalf("South chow: %v", err)
	}
	// West declares Pong.
	if err := g.ActionClaim(SeatIndex(West), ClaimAction{Type: ClaimDeclare, Meld: MeldPong}); err != nil {
		t.Fatalf("West pong: %v", err)
	}
	// North passes.
	if err := g.ActionClaim(SeatIndex(North), ClaimAction{Type: ClaimPass}); err != nil {
		t.Fatalf("North pass: %v", err)
	}
	if g.State.Round.Claim != nil {
		t.Fatal("claim window not closed after resolution")
	}
	if g.State.Round.CurrentPlayer != SeatIndex(West) {
		t.Fatalf("current player = %d, want West (pong beats chow)", g.State.Round.CurrentPlayer)
	}
	// South's chow should NOT have produced a meld; West's pong should.
	if got := len(g.State.Players[South].Melds); got != 0 {
		t.Fatalf("South melds = %d, want 0", got)
	}
	if got := len(g.State.Players[West].Melds); got != 1 || g.State.Players[West].Melds[0].Type != MeldPong {
		t.Fatalf("West melds = %v, want one Pong", g.State.Players[West].Melds)
	}
}

// Pong-vs-Pong tiebreaks (two seats both able to Pong the same discard)
// are not feasible in real play: the 4-copy-per-rank constraint means
// only one seat can hold 2+ matching tiles after a discard. The
// closest-next tiebreak is exercised at the unit level by
// TestClaimWindowResolveTiebreakByClosestNext above.

// ----------------------------------------------------------------------------
// No robbing the Kong
// ----------------------------------------------------------------------------

func TestConcealedKongDoesNotReopenClaimWindow(t *testing.T) {
	// East draws 5-Wan. East has 5-Wan, 5-Wan, 5-Wan in hand + the
	// drawn 5-Wan. East declares a concealed Kong. The replacement draw
	// should not be intercepted by a claim window.
	g := &Game{
		ID: "concealed-kong",
		State: &GameState{
			Round: Round{
				Phase:          PhasePlay,
				CurrentPlayer:  SeatIndex(East),
				NewlyDrawnTile: wanTile(5, 3),
			},
		},
	}
	g.State.Players[East].Hand = []*Tile{wanTile(5, 0), wanTile(5, 1), wanTile(5, 2), wanTile(1, 0)}
	// Wall with one normal tile for the replacement draw.
	g.State.Wall.DrawPile = []*Tile{wanTile(2, 0)}

	if err := g.DeclareConcealedKong(); err != nil {
		t.Fatalf("declare concealed kong: %v", err)
	}
	// After declaration: claim window must NOT be open, phase stays
	// PhasePlay, replacement tile is drawn.
	if g.State.Round.Phase != PhasePlay {
		t.Fatalf("phase = %v, want PhasePlay (no claim window for concealed kongs)", g.State.Round.Phase)
	}
	if g.State.Round.Claim != nil {
		t.Fatal("claim window opened after concealed kong; this enables robbing-the-kong")
	}
	if g.State.Round.NewlyDrawnTile == nil {
		t.Fatal("concealed kong did not trigger replacement draw")
	}
	if g.State.Round.NewlyDrawnTile.ID != wanTile(2, 0).ID {
		t.Fatalf("replacement tile = %v, want 2-Wan", g.State.Round.NewlyDrawnTile.Describe())
	}
	// The Kong should be in East's melds.
	var found *Meld
	for _, m := range g.State.Players[East].Melds {
		if m.Type == MeldKong && m.Kong == KongConcealed {
			found = m
			break
		}
	}
	if found == nil {
		t.Fatal("concealed Kong not recorded in melds")
	}
}

// ----------------------------------------------------------------------------
// Concealed-Kong chaining
// ----------------------------------------------------------------------------

func TestConcealedKongChainsAcrossSuccessiveDraws(t *testing.T) {
	// Set up East with a hand that can form two consecutive concealed
	// Kongs. East's first draw completes a 4-of-a-kind of 1-Wan. The
	// replacement draw completes a 4-of-a-kind of 2-Wan. The third draw
	// does not complete a Kong.
	g := &Game{
		ID: "kong-chain",
		State: &GameState{
			Round: Round{
				Phase:          PhasePlay,
				CurrentPlayer:  SeatIndex(East),
				NewlyDrawnTile: wanTile(1, 3),
			},
		},
	}
	g.State.Players[East].Hand = []*Tile{
		wanTile(1, 0), wanTile(1, 1), wanTile(1, 2),
		wanTile(2, 0), wanTile(2, 1), wanTile(2, 2),
		wanTile(9, 0),
	}
	// Wall: 2-Wan copy 3 first (completes the second kong), then 3-Wan
	// (does not complete anything).
	g.State.Wall.DrawPile = []*Tile{wanTile(2, 3), wanTile(3, 0)}

	if err := g.DeclareConcealedKong(); err != nil {
		t.Fatalf("first concealed kong: %v", err)
	}
	// After first kong: 1-Wan Kong in melds, replacement draw is the
	// 2-Wan copy 3, which completes the second 4-of-a-kind.
	if g.State.Round.NewlyDrawnTile == nil || g.State.Round.NewlyDrawnTile.ID != wanTile(2, 3).ID {
		t.Fatalf("replacement = %v, want 2-Wan", g.State.Round.NewlyDrawnTile.Describe())
	}
	if err := g.DeclareConcealedKong(); err != nil {
		t.Fatalf("second concealed kong: %v", err)
	}
	// After second kong: 2-Wan Kong in melds, replacement draw is the
	// 3-Wan, which does NOT complete a Kong. So the player must now
	// discard.
	if g.State.Round.NewlyDrawnTile == nil || g.State.Round.NewlyDrawnTile.ID != wanTile(3, 0).ID {
		t.Fatalf("second replacement = %v, want 3-Wan", g.State.Round.NewlyDrawnTile.Describe())
	}
	// Try to declare another Kong — should fail (no fourth 3-Wan in hand).
	if err := g.DeclareConcealedKong(); !errors.Is(err, ErrTileNotClaim) {
		t.Fatalf("third concealed kong = %v, want ErrTileNotClaim", err)
	}
	// Both Kongs are in melds.
	kongCount := 0
	for _, m := range g.State.Players[East].Melds {
		if m.Type == MeldKong && m.Kong == KongConcealed {
			kongCount++
		}
	}
	if kongCount != 2 {
		t.Fatalf("concealed kongs = %d, want 2", kongCount)
	}
}

// ----------------------------------------------------------------------------
// Pong upgrade to Kong
// ----------------------------------------------------------------------------

func TestPongUpgradeToKongOnSelfDrawnFourthTile(t *testing.T) {
	// East holds an exposed Pong of 5-Wan and self-draws the fourth
	// 5-Wan. East upgrades the Pong to a Kong and draws a replacement.
	g := &Game{
		ID: "pong-upgrade",
		State: &GameState{
			Round: Round{
				Phase:          PhasePlay,
				CurrentPlayer:  SeatIndex(East),
				NewlyDrawnTile: wanTile(5, 3),
			},
		},
	}
	g.State.Players[East].Hand = []*Tile{wanTile(9, 0), wanTile(9, 1)}
	g.State.Players[East].Melds = []*Meld{
		{
			Type:     MeldPong,
			Tiles:    []*Tile{wanTile(5, 0), wanTile(5, 1), wanTile(5, 2)},
			FromSeat: SeatIndex(South),
		},
	}
	g.State.Wall.DrawPile = []*Tile{wanTile(1, 0)}

	// Find the pong's meld index.
	pongIdx := -1
	for i, m := range g.State.Players[East].Melds {
		if m.Type == MeldPong {
			pongIdx = i
			break
		}
	}
	if pongIdx < 0 {
		t.Fatal("setup: no pong to upgrade")
	}

	if err := g.UpgradePongToKong(pongIdx); err != nil {
		t.Fatalf("upgrade: %v", err)
	}
	meld := g.State.Players[East].Melds[pongIdx]
	if meld.Type != MeldKong {
		t.Fatalf("meld type = %v, want MeldKong", meld.Type)
	}
	if meld.Kong != KongExposed {
		t.Fatalf("meld kong = %v, want KongExposed", meld.Kong)
	}
	if len(meld.Tiles) != 4 {
		t.Fatalf("meld tiles = %d, want 4", len(meld.Tiles))
	}
	if g.State.Round.NewlyDrawnTile == nil || g.State.Round.NewlyDrawnTile.ID != wanTile(1, 0).ID {
		t.Fatalf("replacement = %v, want 1-Wan", g.State.Round.NewlyDrawnTile.Describe())
	}
}

func TestPongUpgradeRejectsMismatchedDrawnTile(t *testing.T) {
	g := &Game{
		ID: "pong-upgrade-mismatch",
		State: &GameState{
			Round: Round{
				Phase:          PhasePlay,
				CurrentPlayer:  SeatIndex(East),
				NewlyDrawnTile: wanTile(7, 0), // wrong suit/rank
			},
		},
	}
	g.State.Players[East].Hand = []*Tile{wanTile(9, 0)}
	g.State.Players[East].Melds = []*Meld{
		{
			Type:  MeldPong,
			Tiles: []*Tile{wanTile(5, 0), wanTile(5, 1), wanTile(5, 2)},
		},
	}
	err := g.UpgradePongToKong(0)
	if !errors.Is(err, ErrTileNotClaim) {
		t.Fatalf("upgrade with mismatched tile = %v, want ErrTileNotClaim", err)
	}
}

// ----------------------------------------------------------------------------
// All-pass turn advancement
// ----------------------------------------------------------------------------

func TestAllPassAdvancesToNextSeat(t *testing.T) {
	discard := wanTile(5, 0)
	g := claimTestHelper(t, map[Wind][]*Tile{
		East:  {wanTile(1, 0)},
		South: {wanTile(9, 0)},
		West:  {wanTile(9, 1)},
		North: {wanTile(9, 2)},
	}, []*Tile{discard}, East, discard)
	if err := g.Discard(discard.ID); err != nil {
		t.Fatalf("discard: %v", err)
	}
	for _, w := range []Wind{South, West, North} {
		if err := g.ActionClaim(SeatIndex(w), ClaimAction{Type: ClaimPass}); err != nil {
			t.Fatalf("seat %d pass: %v", w, err)
		}
	}
	if g.State.Round.Phase != PhasePlay {
		t.Fatalf("phase = %v, want PhasePlay", g.State.Round.Phase)
	}
	if g.State.Round.CurrentPlayer != SeatIndex(South) {
		t.Fatalf("current player = %d, want South", g.State.Round.CurrentPlayer)
	}
	if g.State.Round.NewlyDrawnTile != nil {
		t.Fatal("all-pass should not pre-draw a tile")
	}
}

// ----------------------------------------------------------------------------
// Displayer never claims their own discard
// ----------------------------------------------------------------------------

func TestDiscarderCannotClaimOwnDiscard(t *testing.T) {
	// East discards 5-Wan. East has 5-Wan, 5-Wan, 5-Wan in hand but is
	// excluded from the claim window by pre-acting. The discarder
	// cannot claim, period.
	discard := wanTile(5, 0)
	g := claimTestHelper(t, map[Wind][]*Tile{
		East:  {wanTile(5, 1), wanTile(5, 2), wanTile(5, 3)},
		South: {wanTile(9, 0)},
		West:  {wanTile(9, 1)},
		North: {wanTile(9, 2)},
	}, []*Tile{discard}, East, discard)
	if err := g.Discard(discard.ID); err != nil {
		t.Fatalf("discard: %v", err)
	}
	// The discarder was pre-marked as acted; trying to act again
	// returns ErrAlreadyActed (and East is not the PendingSeat anyway).
	if g.State.Round.Claim == nil || !g.State.Round.Claim.Acted[East] {
		t.Fatal("discarder should be pre-marked as acted")
	}
	if g.State.Round.Claim.PendingSeat == SeatIndex(East) {
		t.Fatal("PendingSeat should not be the discarder")
	}
	// Drives all to pass; turn should advance to South.
	for _, w := range []Wind{South, West, North} {
		if err := g.ActionClaim(SeatIndex(w), ClaimAction{Type: ClaimPass}); err != nil {
			t.Fatalf("seat %d pass: %v", w, err)
		}
	}
	if g.State.Round.CurrentPlayer != SeatIndex(South) {
		t.Fatalf("current player = %d, want South", g.State.Round.CurrentPlayer)
	}
}
