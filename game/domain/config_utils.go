package domain

type tileKey struct {
	suit Suit
	rank uint8
}

// IsSimpleWinningHand reports whether the player's committed melds plus
// the tiles remaining in p.Hand form a standard winning hand:
//
//   - exactly 4 total melds
//   - exactly 1 pair
//
// A committed Kong counts as one meld even though it contains 4 physical
// tiles.
//
// The hand may therefore contain 14, 15, 16, ... physical tiles depending
// on the number of committed Kongs.
//
// Callers that have just drawn the winning tile should add it to p.Hand
// before calling.
func IsSimpleWinningHand(p *PlayerState) bool {
	committedMelds := len(p.Melds)

	// A standard hand has exactly 4 melds.
	if committedMelds > 4 {
		return false
	}

	// Number of additional melds required from the concealed hand.
	requiredMelds := 4 - committedMelds

	counts := tileCounts(p.Hand)

	// The remaining hand must be exactly:
	//
	//   requiredMelds * 3 + 2
	//
	// physical tiles.
	//
	// Committed Kongs are intentionally not included here because they
	// are already represented by p.Melds.
	expectedHandTiles := requiredMelds*3 + 2

	if totalTileCount(counts) != expectedHandTiles {
		return false
	}

	return canFormWinningHand(counts, requiredMelds)
}

// tileCounts groups physical hand tiles by tile kind.
func tileCounts(hand []*Tile) map[tileKey]int {
	counts := make(map[tileKey]int)

	for _, t := range hand {
		key := tileKey{
			suit: t.Suit,
			rank: t.Rank,
		}

		counts[key]++
	}

	return counts
}

// totalTileCount returns the number of physical tiles represented
// by the tile-count map.
func totalTileCount(counts map[tileKey]int) int {
	total := 0

	for _, count := range counts {
		total += count
	}

	return total
}

// canFormWinningHand determines whether the remaining hand can be
// decomposed into:
//
//   - requiredMelds melds
//   - exactly one pair
//
// It uses backtracking rather than greedy extraction because a tile
// can sometimes participate in multiple valid decompositions.
func canFormWinningHand(
	counts map[tileKey]int,
	requiredMelds int,
) bool {
	if requiredMelds == 0 {
		return hasPair(counts)
	}

	// Find the first remaining tile.
	key, ok := firstTile(counts)
	if !ok {
		return false
	}

	// ---------------------------------------------------------
	// Option 1: use this tile as a Pong
	// ---------------------------------------------------------

	if counts[key] >= 3 {
		counts[key] -= 3

		if canFormWinningHand(counts, requiredMelds-1) {
			counts[key] += 3
			return true
		}

		counts[key] += 3
	}

	// ---------------------------------------------------------
	// Option 2: use this tile as the beginning of a Chow
	// ---------------------------------------------------------

	if isSuited(key.suit) &&
		key.rank >= 1 &&
		key.rank <= 7 {

		second := tileKey{
			suit: key.suit,
			rank: key.rank + 1,
		}

		third := tileKey{
			suit: key.suit,
			rank: key.rank + 2,
		}

		if counts[second] > 0 && counts[third] > 0 {
			counts[key]--
			counts[second]--
			counts[third]--

			if canFormWinningHand(counts, requiredMelds-1) {
				counts[key]++
				counts[second]++
				counts[third]++
				return true
			}

			counts[key]++
			counts[second]++
			counts[third]++
		}
	}

	return false
}

// hasPair checks whether exactly one pair can be selected from
// the remaining tiles.
//
// At the point this function is called, all required melds have
// already been extracted.
func hasPair(counts map[tileKey]int) bool {
	for key, count := range counts {
		if count >= 2 {
			counts[key] -= 2

			// A valid winning hand must have no tiles left.
			if totalTileCount(counts) == 0 {
				counts[key] += 2
				return true
			}

			counts[key] += 2
		}
	}

	return false
}

// firstTile returns any remaining tile kind.
func firstTile(counts map[tileKey]int) (tileKey, bool) {
	for key, count := range counts {
		if count > 0 {
			return key, true
		}
	}

	return tileKey{}, false
}

// isSuited returns whether the tile can participate in a Chow.
func isSuited(suit Suit) bool {
	return suit >= SuitCharacter && suit <= SuitDot
}
