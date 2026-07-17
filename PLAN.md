# Singapore Mahjong Implementation Plan

This plan builds a playable single-round Singapore Mahjong engine from the
current domain model. Complete each phase in order; later phases rely on the
state and invariants established by earlier ones.

## Current Foundation

- [x] Define the 148-tile catalog: suited tiles, winds, dragons, flowers, and animals.
- [x] Provide human-readable tile descriptions and flower/animal classifiers.
- [x] Model a wall with deterministic shuffling, drawing, and remaining-tile count.
- [x] Deal initial hands: 13 normal tiles per player and 14 for East.
- [x] Reveal flowers and animals during the initial deal and draw replacements.

## Phase 1: Game Setup

Goal: create a ready-to-play `Game` without setup code in `main.go`.

- [x] Add `NewGame(id string, config GameConfig, seed uint64) *Game` in `game/domain/game.go`.
- [x] Copy `constants.Catalog` into a new wall slice before shuffling; never shuffle the canonical catalog itself.
- [x] Initialize the round to `PhaseDealing`, with no discard and no open claim window.
- [x] Shuffle the wall and call `GameState.DealInitialHands()`.
- [x] Set the round to `PhasePlay`, with East as `CurrentPlayer` and East's fourteenth tile as `NewlyDrawnTile`.
- [x] Keep `game/main.go` limited to creating and displaying a game.
- [x] Add setup tests: four hands contain 13/14 normal tiles, bonus tiles are exposed, and the wall count is correct.

## Phase 2: Player Hand Operations

Goal: make hand mutation safe, reusable, and predictable.

- [x] Add a stable tile sort order in `game/domain/tile.go` or `player.go`: Wan, Bamboo, Dot, winds, dragons.
- [x] Add `PlayerState.SortHand()` in IPlayerState interface and implement in `game/domain/player.go`.
- [x] Add a helper to find a tile in a hand by `TileID`.
- [x] Add `PlayerState.RemoveFromHand(id TileID) (*Tile, bool)`.
- [x] Add helpers for adding a normal tile, flower, or animal to the proper player collection.
- [x] Add tests for sorting, duplicate-looking physical tiles, removal, and missing tile IDs.

## Phase 3: Normal Turn Flow

Goal: support draw then discard, starting from East's opening discard.

- [x] Add `Game.DrawForCurrentPlayer() error` in `game/domain/game.go`.
- [x] Draw a normal tile from the wall into the current player's hand.
- [x] If the tile is a flower or animal, expose it and continue drawing a replacement.
- [x] Update `Round.NewlyDrawnTile` with the final normal tile.
- [x] Add `Game.Discard(player SeatIndex, id TileID) error`.
- [x] Reject discards when it is not the player's turn, the round is not in `PhasePlay`, or the tile is absent.
- [x] Move the discarded tile to `PlayerState.Discards`; set `Round.LastDiscard` and `Round.LastDiscardBy`.
- [x] Clear `NewlyDrawnTile` after a discard.
- [x] Advance to the next seat only if no player claims the discard.
- [x] Add turn-flow tests for East's opening discard and a later player's draw/discard cycle.

## Phase 4: Claims and Melds

Goal: let players claim a discard legally and represent the resulting meld.

House rules pinned before coding this phase (recorded here so future phases
can rely on them):

- **Chow:** allowed. Only the player seated immediately after the discarder
  (the next player in turn order) may claim a Chow. The two players seated
  before the discarder cannot Chow. Chow is only valid on suited tiles and
  only on a discard (not on a self-drawn tile).
- **Pong:** any player may Pong a discard of a tile they hold two copies of.
  Pong consumes 2 hand tiles + 1 discard and exposes a triplet.
- **Kong (exposed):** any player may Kong a discard of a tile they hold three
  copies of. Forms an exposed Kong of 4 tiles (3 hand + 1 discard).
- **Kong (concealed):** a player holding a concealed pung may declare Kong
  when they self-draw the fourth matching tile. Forms a concealed Kong of 4
  tiles.
- **Robbing the Kong (抢杠):** not allowed. A concealed Kong cannot be
  intercepted for a win. The discard claim window closes as soon as a Kong
  is declared; no second window opens on a concealed-Kong upgrade.
- **Kong turn flow (algorithm view):** declaring a Kong transitions the
  current player to the Kong claimant, who then performs a normal
  draw+discard cycle before the turn passes on. The replacement draw is
  mandatory and happens immediately; if it is a bonus tile it is exposed
  in `PlayerState.Flowers` / `PlayerState.Animals` and the player keeps
  drawing until a normal tile is drawn, then discards. Self-draw wins off
  the Kong replacement tile are evaluated as normal self-drawn wins.
- **Post-claim turn flow (Kong vs. Pong/Chow):** after a successful Kong,
  the claimant draws a replacement tile (with normal bonus-tile
  replacement) and then discards. After a successful Pong or Chow, the
  claimant does **not** draw — the state machine goes straight to
  discard. In all three cases the claimant becomes the new
  `CurrentPlayer`. Kong is therefore the only claim type that consumes a
  wall tile.
- **Concealed-Kong chaining:** a player who draws a tile that completes
  a 4-of-a-kind may declare another concealed Kong immediately (no cap on
  the number of chained kongs per turn). Each declaration triggers
  another replacement draw. The chain ends when the player draws a tile
  that does not complete a Kong, at which point they must discard.
- **Upgrade exposed Pong to Kong:** a player holding an exposed Pong may
  self-draw the fourth matching tile and upgrade the existing Meld to a
  Kong (treated as `KongExposed`; `KongAdded` is reserved but not
  semantically distinct in this implementation). The fourth tile is
  added to the existing Meld's `Tiles`, and the player draws a
  replacement.
- **Action type exclusivity per declaration:** during a `ClaimWindow`,
  each seat declares at most one action (Pong, Kong, Chow, or Pass). A
  seat cannot declare two claim types on the same discard. Self-actions
  such as concealed-Kong declarations and Pong-upgrade-to-Kong only
  happen during `PhasePlay` on the active player's turn and are not
  routed through the claim window.
- **Tie-breaking on competing claims:** when multiple players can legally
  claim the same discard (e.g. Kong vs. Pong, or two Pongs), the seat
  closest-next to the discarder in turn order wins. The discarder never
  claims their own discard.
- **Last-tile self-draw win:** legal. A draw that exhausts the wall can
  still be a winning tile if the resulting hand satisfies the winning
  pattern.
- **Priority order when types differ:** Kong > Pong > Chow. Chow is only
  considered for the next-player seat; Kong and Pong compete across all
  seats using the closest-next tiebreak.
- **Bonus tiles in claims:** flowers and animals are exposed on draw and
  never appear in `Discards`, so they cannot be claimed for any meld.

Tasks:

- [ ] Define precise `Meld` and `KongType` invariants in `game/domain/meld.go`.
- [ ] Add helpers to count matching tiles and select required hand tiles.
- [ ] Open a `ClaimWindow` after every discard (one window per discard, no
      re-opens for concealed-Kong upgrades).
- [ ] Implement claim priority and seat ordering per the house rules above
      (Kong > Pong > Chow; closest-next seat wins ties; discarder excluded).
- [ ] Implement `Pong`: consume two matching hand tiles plus the discard.
- [ ] Implement `Chow`: next-player-only, suited-tile sequences only, on
      discard only.
- [ ] Implement exposed and concealed kongs, including replacement draws
      and the draw+discard cycle that follows a Kong declaration.
- [ ] Support upgrade of an exposed Pong to a Kong on a self-drawn fourth
      tile, including the replacement draw.
- [ ] Support concealed-Kong chaining across successive replacement draws
      until a non-completing tile is drawn.
- [ ] Close a claim window on pass or once the highest-priority legal claim
      resolves; the turn then proceeds per the claim type's flow.
- [ ] Add tests for legal and illegal claims, multiple claimants, priority
      resolution, and the no-robbing-the-Kong rule.

## Phase 5: Winning-Hand Validation

Goal: determine whether a player has a legal winning hand.

- [ ] Define the supported winning patterns and any house-rule options in `GameConfig`.
- [ ] Build a hand-normalization helper that separates bonus tiles, fixed melds, and concealed tiles.
- [ ] Implement the standard four-melds-and-a-pair solver.
- [ ] Support sequences only in suited tiles; honors can form triplets/pairs but not sequences.
- [ ] Add special-hand recognition only after the standard solver is fully tested.
- [ ] Expose a method such as `CanWin(hand []*Tile, melds []*Meld, winningTile *Tile) bool`.
- [ ] Add table-driven tests for valid hands, near misses, honors, and duplicate physical tiles.

## Phase 6: Scoring and Round Completion

Goal: score a declared win and end the round consistently.

- [ ] Define `WinResult` in `game/domain/scoring.go`: winner, source of winning tile, winning method, tai, and payments.
- [ ] Implement the base Singapore Mahjong scoring rules selected for this project.
- [ ] Count flowers, animals, seat/round wind, and meld-related bonuses as applicable.
- [ ] Enforce `GameConfig.MinTai` before accepting a win.
- [ ] Apply score changes to all four players.
- [ ] Transition the round to `PhaseRoundEnd` and prevent further actions.
- [ ] Add score tests for self-draw, discard wins, minimum tai, and bonus tiles.
- [ ] Include a self-draw win off the last wall tile in scoring tests, since
      it is a legal win path under the house rules.

## Phase 7: Round Progression and End Conditions

Goal: finish non-winning rounds and support repeated rounds.

- [ ] Define when the wall is exhausted and whether a final-tile draw is allowed.
- [ ] End the round as a draw when no legal draw remains. Note: the last
      legal draw is still a win path (self-draw off the final wall tile);
      round draws only happen when a player must draw but the wall is empty
      *and* the resulting hand is not a winning hand.
- [ ] Record dealer retention and round-wind progression rules.
- [ ] Add a method to reset per-round state while preserving player scores.
- [ ] Test dealer wins, dealer losses, drawn rounds, and consecutive rounds.

## Phase 8: Interface and Integration

Goal: make the domain engine usable from a CLI, server, or UI.

- [ ] Replace direct domain-slice access in `game/main.go` with the game setup API.
- [ ] Add a small CLI view of player hands, exposed bonus tiles, discards, and remaining wall tiles.
- [ ] Keep input parsing and display code outside `game/domain`.
- [ ] Add serialization-friendly DTOs only at the application boundary, not inside core rule logic.
- [ ] Add an end-to-end deterministic simulation using a fixed seed.

## Cross-Cutting Standards

- [ ] Do not mutate `constants.Catalog`; every game owns its own copied wall slice.
- [ ] Use `TileID` to distinguish physical copies of otherwise identical tiles.
- [ ] Use a supplied seed or `rand.Rand` for reproducible tests.
- [ ] Avoid printing from domain methods; return values or errors instead.
- [ ] Add table-driven tests beside the domain code as each feature is introduced.
- [ ] Run `gofmt` and `go test ./...` after each completed phase.
