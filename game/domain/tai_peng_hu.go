package domain

// 4 points
// CheckPengHu checks for Big Ping Hu (平胡), which awards 4 points (Tai) in Singaporean Mahjong.
func CheckPengHu(tiles []*Tile, _ *PlayerState) int64 {
	if len(tiles) != 14 {
		return 0
	}

	// Group tile counts by Suit and Rank (1-9)
	suitRanks := make(map[Suit][10]int)

	for _, t := range tiles {
		// Ping Hu strictly requires suited tiles (ranks 1-9).
		// If Rank is outside 1-9, it contains honors, flowers, or animals.
		if t.Rank < 1 || t.Rank > 9 {
			return 0
		}
		counts := suitRanks[t.Suit]
		counts[t.Rank]++
		suitRanks[t.Suit] = counts
	}

	// Try every possible suit and rank as the pair (eyes)
	foundPair := false
	for pairSuit, pairCounts := range suitRanks {
		for r := uint8(1); r <= 9; r++ {
			if pairCounts[r] >= 2 {
				// Clone the map and remove the pair
				copiedSuitRanks := cloneSuitRanks(suitRanks)
				cCounts := copiedSuitRanks[pairSuit]
				cCounts[r] -= 2
				copiedSuitRanks[pairSuit] = cCounts

				// Check if all remaining tiles across all suits can form valid chows
				if validateAllSuitsChows(copiedSuitRanks) {
					foundPair = true
					break
				}
			}
		}
		if foundPair {
			break
		}
	}

	if foundPair {
		return 4 // 4 points for Big Ping Hu
	}
	return 0
}

func cloneSuitRanks(original map[Suit][10]int) map[Suit][10]int {
	clone := make(map[Suit][10]int, len(original))
	for k, v := range original {
		clone[k] = v
	}
	return clone
}

// validateAllSuitsChows verifies that every suit's remaining tiles decompose into chows.
func validateAllSuitsChows(suitRanks map[Suit][10]int) bool {
	for _, counts := range suitRanks {
		if !canFormChows(counts) {
			return false
		}
	}
	return true
}

// canFormChows recursively checks if a 1-9 rank frequency array forms valid sequences of 3.
func canFormChows(counts [10]int) bool {
	var r uint8 = 1
	for r <= 9 && counts[r] == 0 {
		r++
	}
	if r > 9 {
		return true // All tiles successfully accounted for
	}

	count := counts[r]
	if count == 0 {
		return true
	}

	// A chow starting at r requires r, r+1, r+2. Cannot start a chow at 8 or 9.
	if r > 7 {
		return false
	}

	if counts[r+1] < count || counts[r+2] < count {
		return false
	}

	// Consume the tiles for these sequences
	counts[r] -= count
	counts[r+1] -= count
	counts[r+2] -= count

	return canFormChows(counts)
}
