package domain

type Tai uint8

type WinType uint8

const (
	WinSelfDraw WinType = iota
	WinOnDiscard
)

type TaiEntry struct {
	Name  string
	Value Tai
}

type WinResult struct {
	Winner       SeatIndex
	Type         WinType
	TotalTai     Tai
	TaiBreakdown []TaiEntry
}
