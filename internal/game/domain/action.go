package domain

type ActionType uint8

const (
	ActionDraw ActionType = iota

	ActionDiscard

	ActionChow
	ActionPong
	ActionKong

	ActionWin

	ActionPass
)

type Action struct {
	Player int

	Type ActionType

	Tile *Tile
}