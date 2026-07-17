package domain

type MeldType uint8

const (
	MeldChow MeldType = iota
	MeldPong
	MeldKong
)

type Meld struct {
	Type       MeldType
	Tiles      []Tile
	FromPlayer int  // -1 if concealed
	Concealed  bool
}

