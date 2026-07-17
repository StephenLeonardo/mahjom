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

- [ ] Add `Game.DrawForCurrentPlayer() error` in `game/domain/game.go`.
- [ ] Draw a normal tile from the wall into the current player's hand.
- [ ] If the tile is a flower or animal, expose it and continue drawing a replacement.
- [ ] Update `Round.NewlyDrawnTile` with the final normal tile.
- [ ] Add `Game.Discard(player SeatIndex, id TileID) error`.
- [ ] Reject discards when it is not the player's turn, the round is not in `PhasePlay`, or the tile is absent.
- [ ] Move the discarded tile to `PlayerState.Discards`; set `Round.LastDiscard` and `Round.LastDiscardBy`.
- [ ] Clear `NewlyDrawnTile` after a discard.
- [ ] Advance to the next seat only if no player claims the discard.
- [ ] Add turn-flow tests for East's opening discard and a later player's draw/discard cycle.

## Phase 4: Claims and Melds

Goal: let players claim a discard legally and represent the resulting meld.

- [ ] Define precise `Meld` and `KongType` invariants in `game/domain/meld.go`.
- [ ] Add helpers to count matching tiles and select required hand tiles.
- [ ] Open a `ClaimWindow` after every discard.
- [ ] Implement claim priority and seat ordering according to the selected Singapore Mahjong ruleset.
- [ ] Implement `Pong`: consume two matching hand tiles plus the discard.
- [ ] Implement `Chow`: allow only the next player to take a suited-tile sequence, if permitted by the ruleset.
- [ ] Implement exposed and concealed kongs, including replacement draws.
- [ ] Close a claim window on pass or once the highest-priority legal claim resolves.
- [ ] Add tests for legal and illegal claims, multiple claimants, and priority resolution.

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

## Phase 7: Round Progression and End Conditions

Goal: finish non-winning rounds and support repeated rounds.

- [ ] Define when the wall is exhausted and whether a final-tile draw is allowed.
- [ ] End the round as a draw when no legal draw remains.
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
