// Package domain defines value types and aggregates for a single Singapore Mahjong table.
//
// Invariants:
//   - One Game is self-contained; no cross-game or match history lives here.
//   - Players[0] is East and the dealer; seat wind for seat i is Wind(i).
//   - The tile set has TileSetSize (148) physical tiles; see tile.go for the catalog.
//   - Do not store fields derivable in O(1) from seat index or meld type (e.g. SeatWind, DealerSeat, Meld.Open).
package domain
