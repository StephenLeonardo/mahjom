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
	Actor SeatIndex
	Type  ActionType

	Discard Tile
	Take    *Tile
	Tiles   []Tile
	Kong    KongType
	Win     *WinResult
}
