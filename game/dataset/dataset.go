package dataset

import "mahjom/game/domain"

type TrainingSample struct {
	Seed      int    `json:"seed"`
	GameID    string `json:"game_id"`
	StateName string `json:"state_name"`

	DecisionID int `json:"decision_id"`

	Player int `json:"player"`

	State *GameState `json:"state"`

	Action *Action `json:"action"`

	Outcome *GameOutcome `json:"outcome"`
}

type GameState struct {
	Round *RoundState `json:"round"`

	MainPlayer *MainPlayerState `json:"main_player"`

	Opponents [3]*OpponentState `json:"opponents"`

	LastDiscard *domain.Tile `json:"last_discard"`

	LastDiscardedBy *int `json:"last_discarded_by"`
}

type RoundState struct {
	Turn int `json:"turn"`

	CurrentPlayer int `json:"current_player"`

	RemainingWall int `json:"remaining_wall"`
}

type MainPlayerState struct {
	Hand []*domain.Tile `json:"hand"`

	Melds []*domain.Meld `json:"melds"`

	Flowers []*domain.Tile `json:"flowers"`

	Animals []*domain.Tile `json:"animals"`

	Discards []*domain.Tile `json:"discards"`

	Score int `json:"score"`

	SeatWind domain.Wind `json:"seat_wind"`
}

type OpponentState struct {
	Melds []*domain.Meld `json:"melds"`

	Flowers []*domain.Tile `json:"flowers"`

	Animals []*domain.Tile `json:"animals"`

	Discards []*domain.Tile `json:"discards"`

	Score int `json:"score"`

	SeatWind domain.Wind `json:"seat_wind"`
}

type ActionType uint8

const (
	ActionDiscard ActionType = iota

	ActionPong

	ActionChow

	ActionKong

	ActionWin

	ActionPass
)

type Action struct {
	Type ActionType `json:"type"`

	Tile *domain.Tile `json:"tile,omitempty"`
}

type GameOutcome struct {
	Winner int `json:"winner"`

	WinningTai int `json:"winning_tai"`

	FinalScores [4]int `json:"final_scores"`

	TurnsToFinish int `json:"turns_to_finish"`
}
