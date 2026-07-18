package domain

// tileKey identifies a tile by (suit, rank), independent of physical
// TileID. Used by helpers that group tiles by kind for meld/pair checks.
type tileKey struct {
	suit Suit
	rank uint8
}

func countMeldsIncludingHand(p *PlayerState) int {
	// Start with the player's already-committed melds (Pong, Kong, Chow).
	count := len(p.Melds)

	// Build a working copy of the hand so we don't mutate player state
	// while counting. Group tiles by (suit, rank) so we can find Pongs
	// (3+ of a kind) and Chows (3 consecutive suited ranks).
	byKind := make(map[tileKey][]*Tile)
	for _, t := range p.Hand {
		k := tileKey{t.Suit, t.Rank}
		byKind[k] = append(byKind[k], t)
	}

	// Step 1: Pongs (3+ of any suit+rank, including honors and 4-of-a-
	// kind). Each group of 3 counts as one meld; the 4th tile in a
	// 4-of-a-kind is left in the bucket for any later pair check.
	for k, group := range byKind {
		count += len(group) / 3
		byKind[k] = group[:len(group)%3] // keep the 0, 1, or 2 leftover tiles
	}

	// Step 2: Chows (3 consecutive suited ranks). Honors cannot form
	// Chows, so skip them here.
	byRank := make(map[Suit][10]int) // index 1..9 used
	for k, leftover := range byKind {
		if k.suit == SuitWind || k.suit == SuitDragon {
			continue
		}
		if int(k.rank) < 1 || int(k.rank) > 9 {
			continue
		}
		existing := byRank[k.suit]
		existing[k.rank] = len(leftover)
		byRank[k.suit] = existing
	}
	for suit := SuitCharacter; suit <= SuitDot; suit++ {
		ranks, ok := byRank[suit]
		if !ok {
			continue
		}
		// Walk ranks 1..7; a Chow starting at r uses r, r+1, r+2.
		for r := 1; r <= 7; r++ {
			for ranks[r] > 0 && ranks[r+1] > 0 && ranks[r+2] > 0 {
				ranks[r]--
				ranks[r+1]--
				ranks[r+2]--
				count++
			}
		}
	}

	// When the hand is the full 14 tiles (13 in p.Hand + 1 winning
	// tile added by the caller) and count is 4, the remaining 2 tiles
	// just need to form a pair for a winning hand.
	return count
}

// IsSimpleWinningHand reports whether the player's committed melds plus
// the tiles remaining in p.Hand together form a standard winning hand:
// exactly 4 melds (Pong, Kong, or Chow) and 1 pair.
//
// A "simple" hand is the basic 4-melds-and-a-pair structure. Special
// hands (7 pairs, 13 wonders, etc.) are not considered here.
//
// Callers that have just drawn the winning tile should add it
// to p.Hand before calling.
func IsSimpleWinningHand(p *PlayerState) bool {
	// // Standard winning hand is 14 tiles. Committed melds account for
	// // 3 each; the rest must be in p.Hand.
	// totalTiles := len(p.Melds)*3 + len(p.Hand)
	// if totalTiles != 14 {
	// 	return false
	// }

	if countMeldsIncludingHand(p) != 4 {
		return false
	}

	// The leftover tiles after extracting the 4 melds must form a pair.
	// We re-run the greedy extraction on tile counts and inspect what's
	// left.
	leftovers := leftoverTileCounts(p)
	if len(leftovers) != 1 {
		// Either 0 leftovers (impossible: 14 - 4*3 = 2) or 2+ distinct
		// kinds remain, neither of which is a pair.
		return false
	}
	for _, n := range leftovers {
		if n != 2 {
			return false
		}
	}
	return true
}

// leftoverTileCounts returns the number of tiles left over after
// greedily extracting Pongs and Chows from p.Hand, grouped by
// (suit, rank). Used by IsSimpleWinningHand to detect a pair.
func leftoverTileCounts(p *PlayerState) map[tileKey]int {
	counts := make(map[tileKey]int)
	for _, t := range p.Hand {
		counts[tileKey{t.Suit, t.Rank}]++
	}

	// Step 1: Pongs (3+ of any kind). Each group of 3 becomes a meld;
	// the 0/1/2 remainder is leftover.
	for k, n := range counts {
		counts[k] = n % 3
	}

	// Step 2: Chows (3 consecutive suited ranks). Honors can't form
	// Chows, so they stay as leftovers untouched.
	bySuitRank := make(map[Suit][10]int) // index 1..9 used
	for k, n := range counts {
		if k.suit == SuitWind || k.suit == SuitDragon {
			continue
		}
		if int(k.rank) < 1 || int(k.rank) > 9 {
			continue
		}
		existing := bySuitRank[k.suit]
		existing[k.rank] = n
		bySuitRank[k.suit] = existing
	}
	for suit := SuitCharacter; suit <= SuitDot; suit++ {
		ranks, ok := bySuitRank[suit]
		if !ok {
			continue
		}
		// Walk ranks 1..7; a Chow starting at r uses r, r+1, r+2.
		for r := 1; r <= 7; r++ {
			for ranks[r] > 0 && ranks[r+1] > 0 && ranks[r+2] > 0 {
				ranks[r]--
				ranks[r+1]--
				ranks[r+2]--
			}
		}
		// Write the remaining counts back to the main map.
		for r := 1; r <= 9; r++ {
			if ranks[r] > 0 {
				counts[tileKey{suit, uint8(r)}] = ranks[r]
			}
		}
	}

	return counts
}
