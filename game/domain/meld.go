package domain

type MeldType uint8

const (
	MeldChow MeldType = iota
	MeldPong
	MeldKong
)

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
	Tiles    []Tile
	FromSeat SeatIndex
}
