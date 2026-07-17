package domain

type PlayerState struct {
	ID int

	SeatWind Wind

	Hand []Tile

	Melds []Meld

	Flowers []Tile
	Animals []Tile

	Discards []Tile

	Score int
}