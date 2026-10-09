package domain

// CheckHonorPungs calculates Tai for dragon pungs/kongs and seat wind pungs/kongs
func CheckHonorPungs(tiles []*Tile, player *PlayerState) int64 {
	counts := make(map[tileKey]int)
	for _, t := range tiles {
		counts[tileKey{Suit: t.Suit, Rank: t.Rank}]++
	}

	var totalTai int64 = 0

	for key, count := range counts {
		if count >= 3 {
			if isDragonSuit(key.Suit) {
				// Each unique dragon pung/kong is worth 1 Tai
				totalTai += 1
			} else if isWindSuit(key.Suit) {
				// Check Seat Wind match (1 Tai)
				if key.Rank == player.GetPositionRank() {
					totalTai += 1
				}
			}
		}
	}

	return totalTai
}

func isDragonSuit(s Suit) bool {
	return s == SuitDragon
}

func isWindSuit(s Suit) bool {
	return s == SuitWind
}
