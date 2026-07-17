package domain

type Round struct {
	Number int

	CurrentPlayer int

	NewlyDrawnTile *Tile

	LastDiscardedTile *Tile
	LastDiscardedBy   int
}
