package domain

type MeldType uint8

const (
	MeldChow MeldType = iota
	MeldPong
	MeldKong
)

var MeldTypes = []MeldType{
	MeldKong,
	MeldPong,
	MeldChow,
}

type KongType uint8

const (
	KongNone KongType = iota
	KongExposed
	KongConcealed
	KongAdded
)

type Meld struct {
	Type     MeldType
	Kong     KongType
	Tiles    []*Tile
	FromSeat SeatIndex

	// For logging
	FromPosition string // E | S | W | N
}
