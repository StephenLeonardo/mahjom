# mahjom

`mahjom` is a Go implementation of the domain logic for a single Singapore Mahjong table.

The code in this repository focuses on the game engine: tiles, players, wall handling, round state, turn flow, and the data needed for later claim and scoring logic. It does not currently try to model a full user interface or multiplayer transport.

## Current scope

The domain package models one self-contained game:

- 4 players, fixed seats
- one shuffled wall
- one active round at a time
- deterministic setup from a seed
- draw and discard flow
- bonus tile exposure for flowers and animals

The code is organized so that each major concept has a focused type:

- `Game` owns the top-level flow and error states
- `GameState` holds the wall, players, and round
- `Round` tracks the current phase and turn metadata
- `Wall` manages the draw pile
- `PlayerState` stores hand tiles, exposed bonuses, melds, and discards
- `Tile` represents a physical tile in the set

## Rules implemented so far

The current rules are intentionally minimal and are based on what the engine already enforces:

### Tile set

The game uses a 148-tile Singapore Mahjong set:

- 108 suited tiles
  - Characters
  - Bamboo
  - Dots
  - 1 to 9 in each suit, 4 copies each
- 16 winds
  - East, South, West, North
  - 4 copies each
- 12 dragons
  - Hong Zhong
  - Fa Cai
  - Bai Ban
  - 4 copies each
- 8 flowers
  - Plum
  - Orchid
  - Chrysanthemum
  - Bamboo
  - 2 copies of each flower
- 4 animals
  - Cat
  - Mouse
  - Rooster
  - Centipede

Each physical tile has a unique `TileID` from `0` to `147`.

### Setup

`domain.NewGame(id, config, seed)` creates a game, shuffles the wall with the provided seed, and deals the opening hands.

Initial dealing currently follows the standard turn structure used in the engine:

- East starts
- each player receives 13 tiles
- East receives one extra tile and starts with 14
- the round starts in `PhasePlay` after the deal is complete

### Bonus tiles

Flowers and animals are treated as bonus tiles.

When a player draws from the wall:

- if the tile is a flower, it is moved to that player’s flower area
- if the tile is an animal, it is moved to that player’s animal area
- the player keeps drawing replacements until a normal hand tile is drawn

That means bonus tiles never remain in the hand.

### Turn flow

The current turn flow is simple:

- a player must draw before discarding
- a player may not draw again until they have discarded
- discarding always applies to the current player
- after discard, the turn advances to the next seat in rotation

The engine records:

- `Round.NewlyDrawnTile` for the most recent hand tile drawn
- `Round.LastDiscard` for the most recent discard
- `Round.LastDiscardBy` for the seat that discarded

### Draw sites and bonus tile handling

Bonus tile handling is consistent across every place a tile leaves the wall
and enters player state:

- initial deal
- normal turn draw
- Kong replacement draw
- last-tile draw (when the wall has exactly one tile left)

In each case, if the drawn tile is a flower or animal it is exposed in
`PlayerState.Flowers` or `PlayerState.Animals` and the player keeps drawing
until a normal hand tile is drawn. The hand never holds a bonus tile. As a
consequence, bonus tiles never appear in `Discards` and cannot be claimed
for a meld.

### Claim rules (house rules, to be implemented in Phase 4)

These are the rules pinned for Phase 4 and later phases. The engine should
not invent conflicting behavior elsewhere.

- **Chow:** allowed. Only the player seated immediately after the discarder
  (the next player in turn order) may claim a Chow. Chow is only valid on
  suited tiles and only on a discard (not on a self-drawn tile).
- **Pong:** any player may Pong a discard of a tile they hold two copies of.
  Pong consumes 2 hand tiles + 1 discard and exposes a triplet.
- **Kong (exposed):** any player may Kong a discard of a tile they hold
  three copies of. Forms an exposed Kong of 4 tiles (3 hand + 1 discard).
- **Kong (concealed):** a player holding a concealed pung may declare Kong
  when they self-draw the fourth matching tile. Forms a concealed Kong of
  4 tiles.
- **Robbing the Kong (抢杠):** not allowed. A concealed Kong cannot be
  intercepted for a win. The discard claim window closes as soon as a Kong
  is declared; no second window opens on a concealed-Kong upgrade.
- **Kong turn flow:** declaring a Kong transitions the current player to
  the Kong claimant, who then performs a normal draw+discard cycle before
  the turn passes on. The replacement draw is mandatory and happens
  immediately; if it is a bonus tile it is exposed and the player keeps
  drawing until a normal tile is drawn, then discards. Self-draw wins off
  the Kong replacement tile are evaluated as normal self-drawn wins.
- **Claim priority:** when multiple players can legally claim the same
  discard, the seat closest-next to the discarder in turn order wins, with
  type priority Kong > Pong > Chow. The discarder never claims their own
  discard.
- **Last-tile self-draw win:** legal. A draw that exhausts the wall can
  still be a winning tile if the resulting hand satisfies the winning
  pattern.

### Errors

The game currently returns a small set of domain errors:

- `ErrRoundNotInPlay`
- `ErrMustDraw`
- `ErrMustDiscard`
- `ErrTileNotInHand`
- `ErrWallEmpty`

## Tile naming

`Tile.Describe()` returns a human-readable name for a tile.

Examples:

- `1 Wan`
- `9 Bamboo`
- `East Wind`
- `Hong Zhong`
- `Plum`
- `Cat`

This is primarily used for debugging and console output.

## Repository layout

- `game/main.go` prints a sample dealt game to stdout
- `game/domain/` contains the core game model and rules
- `PLAN.md` tracks the implementation phases already completed or planned

## Development notes

- The wall shuffle is deterministic for a given seed, which makes test runs reproducible.
- Player hands are sorted with `SortHand()`.
- Drawing and discarding operate on the current player; the caller does not need to pass a seat into `Discard()`.
- Bonus tile handling is already integrated into draw and deal logic.

## What is not implemented yet

The engine still needs the rest of the Mahjong round system:

- claim resolution after a discard (Pong, Kong, next-player-only Chow)
- meld creation and validation, including the no-robbing-the-Kong rule
- round completion detection, including last-tile self-draw wins
- scoring / tai evaluation
- win handling and end-of-round settlement

## Running the sample program

The sample entry point is `game/main.go`.

It creates a new game with the current time as the shuffle seed and prints each player’s hand and bonus counts.

## License

No license file is currently present in this repository.
