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
	Type         MeldType  `json:"type"`
	Kong         KongType  `json:"kong"`
	Tiles        []*Tile   `json:"tiles"`
	FromSeat     SeatIndex `json:"-"`
	FromPosition string    `json:"fromPosition"` // E | S | W | N
}
