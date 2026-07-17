package domain

// Seat 0 is East and the dealer for the whole game; seat i wind is Wind(i).
type SeatIndex uint8

const SeatNone SeatIndex = 255

type Wind uint8

const (
	East Wind = iota
	South
	West
	North
)
