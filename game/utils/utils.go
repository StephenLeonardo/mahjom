package utils

import "mahjom/game/domain"

// SeatDistance returns the number of seats from `from` to `to` going
// clockwise in turn order. The result is in [0,3]; 0 means same seat,
// 1 is the next player, 2 is opposite, 3 is previous.
//
// This is the canonical tiebreak primitive: when two claims of the same
// type compete, the smaller distance wins.
func SeatDistance(from, to domain.SeatIndex) int {
	return (int(to) - int(from) + 4) % 4
}
