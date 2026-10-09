package domain

// CheckPengPengHu checks for All Pungs / Triplets (碰碰胡), which awards 2 points (Tai) in Singaporean Mahjong.
func CheckPengPengHu(tiles []*Tile, _ *PlayerState) int64 {
	if len(tiles) != 14 {
		return 0
	}

	// Count frequencies of each unique tile across the entire hand
	counts := make(map[tileKey]int)
	for _, t := range tiles {
		counts[tileKey{Suit: t.Suit, Rank: t.Rank}]++
	}

	// Test every unique tile as the candidate for the pair (eyes)
	for pairKey, count := range counts {
		if count >= 2 {
			// Clone the frequency map to test this pair configuration
			tempCounts := cloneTileCounts(counts)
			tempCounts[pairKey] -= 2
			if tempCounts[pairKey] == 0 {
				delete(tempCounts, pairKey)
			}

			// Verify that all remaining tiles are grouped into triplets (3) or kongs (4)
			isAllTriplets := true
			for _, c := range tempCounts {
				if c != 3 && c != 4 {
					isAllTriplets = false
					break
				}
			}

			if isAllTriplets {
				return 2 // 2 points for Peng Peng Hu
			}
		}
	}

	return 0
}

func cloneTileCounts(original map[tileKey]int) map[tileKey]int {
	clone := make(map[tileKey]int, len(original))
	for k, v := range original {
		clone[k] = v
	}
	return clone
}
