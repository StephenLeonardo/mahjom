package domain

type GameConfig struct {
	MinTai Tai
}

// ScoringElement represents a pattern or condition that grants Tai
type ScoringElement struct {
	ID       string
	Name     string
	TaiValue int
	// Check is a function that inspects the hand/game state to see if this applies
	Check func(p *PlayerState) bool
}

var ScoringRegistry = []ScoringElement{
	{ID: "SIMPLE_WIN", Name: "Simple Win", TaiValue: 1, Check: IsSimpleWinningHand},
	// {ID: "ALL_CHOW", Name: "All Chow", TaiValue: 4, Check: checkAllChow},
	// {ID: "ALL_PONG", Name: "All Pong", TaiValue: 2, Check: checkAllPong},
	// {ID: "NORMAL_WIN", Name: "Normal WIN", TaiValue: 1, Check: checkNormalWin},
	// {ID: "DRAGON_PUNG", Name: "Dragon Pung", TaiValue: 1, Check: checkDragonPung},
	// {ID: "HALF_FLUSH", Name: "Half Flush", TaiValue: 2, Check: checkHalfFlush},
	// {ID: "FULL_FLUSH", Name: "Full Flush", TaiValue: 4, Check: checkFullFlush},
	// {ID: "BIG_DRAGONS", Name: "Big Three Dragons", TaiValue: 5, Check: checkBigDragons},
}
