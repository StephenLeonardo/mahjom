package domain

type Suit uint8

const (
	SuitCharacter Suit = iota
	SuitWan
	SuitBamboo
	SuitDot
	SuitWind
	SuitDragon
	SuitFlower
	SuitAnimal
)

type Wind uint8

const (
	East Wind = iota
	South
	West
	North
)

type Tile struct {
	ID   uint8 // Unique tile ID (0-43, etc.)
	Suit Suit
	Rank uint8
}

