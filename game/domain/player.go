package domain

type PlayerState struct {
	ID string

	Hand     []*Tile
	Melds    []*Meld
	Flowers  []*Tile
	Animals  []*Tile
	Discards []*Tile

	Score int
}
