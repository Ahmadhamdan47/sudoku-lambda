// v3algorithm.go — V3 generator: pass-graded solvability on top of the V2
// clue windows.
//
// V2 answered "can a player finish this with singles?". V3 additionally asks
// "how long does it take?", using a stricter player model:
//
//   - The solver holds one digit at a time, places every hidden single that
//     digit has (re-scanning until the digit is exhausted), then moves to the
//     next digit. A full 1→9 cycle is one PASS.
//   - Naked singles are never used freely. They are an escape hatch only:
//     when a full pass places nothing (the player is stuck), exactly one
//     naked single may be placed — if the level's budget allows — and
//     hidden-single passes resume. Easy/medium budget 0, hard 1, very hard 2.
//   - A stuck pass (a full 1→9 scan that finds nothing) still counts as a
//     pass: the player did the work of scanning even though nothing placed.
//
// Every candidate is graded on (clues, passes, naked singles used) against
// the per-difficulty tables in v3GradeMatches. No match → discard and
// regenerate. A puzzle that gets stuck with no naked single available, or
// with the budget spent, is disregarded immediately.
//
//   easy       38–42 clues, no naked:  40–42 → 2 passes, 38–39 → 2–3
//   medium     34–38 clues, no naked:  36–38 → 3 passes, 34–35 → 3–4
//   hard       30–34 clues:            32–34 → 4–6 passes (no naked)
//                                      30–31 → 5–6 passes (no naked)
//                                      30–34 → 5–6 passes with exactly 1 naked
//   very hard  26–30 clues:            28–30 → 6–9 passes (no naked)
//                                      26–27 → 7–10 passes (no naked)
//                                      26–30 → 7–10 passes with 1–2 naked
//
// As with V2, very hard never falls back: it retries until a candidate
// genuinely grades into its table. Other levels may return the closest
// solvable candidate with Success=false when the attempt cap is reached —
// that board is still guaranteed finishable under the level's naked budget.
//
// Production and V2 code paths are untouched; this file only adds new symbols.
package sudoku

// V3Rating is the outcome of the V3 player model: digit-cycling hidden
// singles, with naked singles only as a stuck-pass escape hatch.
type V3Rating struct {
	Solvable      bool
	Passes        int // full 1→9 cycles, including stuck ones
	HiddenSingles int
	NakedSingles  int // naked singles actually used (≤ budget)
	StuckPasses   int // passes that placed nothing and forced a naked single
}

// V3Result carries everything the admin test screen needs to inspect a
// generated puzzle. Passes is the difficulty grade; the singles counts are
// diagnostics.
type V3Result struct {
	Puzzle        Board
	Solution      Board
	Clues         int
	Passes        int
	HiddenSingles int
	NakedSingles  int
	Attempts      int
	Success       bool
}

// v3Params returns, per difficulty: the clue window (same windows as V2),
// the overall pass range (union of the sub-ranges in the grade table, used
// only to rank fallback candidates), the naked-single budget, and the
// attempt cap. The fine-grained clue↔pass mapping lives in v3GradeMatches.
func v3Params(difficulty string) (cluesRange, passRange [2]int, nakedBudget, maxAttempts int) {
	switch difficulty {
	case "very hard":
		return [2]int{26, 30}, [2]int{6, 10}, 2, 300
	case "hard":
		return [2]int{30, 34}, [2]int{4, 6}, 1, 100
	case "medium":
		return [2]int{34, 38}, [2]int{3, 4}, 0, 60
	default: // easy
		return [2]int{38, 42}, [2]int{2, 3}, 0, 60
	}
}

// findHiddenSingleForDigit scans rows, then columns, then boxes for a unit
// where the given digit fits in exactly one empty cell. Scan order is fixed
// so ratings are deterministic for a given puzzle.
func findHiddenSingleForDigit(b *Board, d int) (row, col int, found bool) {
	bit := uint16(1) << d
	for r := 0; r < N; r++ {
		count, cc := 0, -1
		for c := 0; c < N; c++ {
			if b[r][c] == 0 && candidateMask(b, r, c)&bit != 0 {
				count++
				cc = c
			}
		}
		if count == 1 {
			return r, cc, true
		}
	}
	for c := 0; c < N; c++ {
		count, rr := 0, -1
		for r := 0; r < N; r++ {
			if b[r][c] == 0 && candidateMask(b, r, c)&bit != 0 {
				count++
				rr = r
			}
		}
		if count == 1 {
			return rr, c, true
		}
	}
	for box := 0; box < N; box++ {
		startRow, startCol := (box/3)*3, (box%3)*3
		count, rr, cc := 0, -1, -1
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				r, c := startRow+i, startCol+j
				if b[r][c] == 0 && candidateMask(b, r, c)&bit != 0 {
					count++
					rr, cc = r, c
				}
			}
		}
		if count == 1 {
			return rr, cc, true
		}
	}
	return 0, 0, false
}

// firstNakedSingle returns the first cell (row-major order) whose candidate
// mask has exactly one digit.
func firstNakedSingle(b *Board) (row, col, num int, found bool) {
	for r := 0; r < N; r++ {
		for c := 0; c < N; c++ {
			if b[r][c] == 0 {
				if d, ok := soleCandidate(candidateMask(b, r, c)); ok {
					return r, c, d, true
				}
			}
		}
	}
	return 0, 0, 0, false
}

// RatePlayerSolveV3 solves the puzzle with the V3 player model: hold digit 1,
// place every hidden single it has until exhausted, move to digit 2, … 9 —
// that cycle is one pass. When a full pass places nothing, spend one naked
// single from the budget (if any remains and one exists) and resume passes;
// otherwise the puzzle is unsolvable under this model and Solvable is false.
func RatePlayerSolveV3(puzzle Board, nakedBudget int) V3Rating {
	b := puzzle
	var rating V3Rating
	for countClues(b) < N*N {
		placed := 0
		for d := 1; d <= 9; d++ {
			for {
				r, c, ok := findHiddenSingleForDigit(&b, d)
				if !ok {
					break
				}
				b[r][c] = d
				rating.HiddenSingles++
				placed++
			}
		}
		rating.Passes++
		if countClues(b) == N*N {
			break
		}
		if placed == 0 {
			if rating.NakedSingles >= nakedBudget {
				return rating
			}
			r, c, d, ok := firstNakedSingle(&b)
			if !ok {
				return rating
			}
			b[r][c] = d
			rating.NakedSingles++
			rating.StuckPasses++
		}
	}
	rating.Solvable = true
	return rating
}

// v3GradeMatches implements the V3 acceptance tables: given the requested
// difficulty, the clue count, and the solve rating, it reports whether the
// candidate is the right difficulty. Anything outside these tables is a
// wrong number → discard and regenerate.
func v3GradeMatches(difficulty string, clues int, r V3Rating) bool {
	if !r.Solvable {
		return false
	}
	p := r.Passes
	switch difficulty {
	case "easy":
		if r.NakedSingles != 0 {
			return false
		}
		switch {
		case clues >= 40 && clues <= 42:
			return p == 2
		case clues >= 38 && clues <= 39:
			return p >= 2 && p <= 3
		}
	case "medium":
		if r.NakedSingles != 0 {
			return false
		}
		switch {
		case clues >= 36 && clues <= 38:
			return p == 3
		case clues >= 34 && clues <= 35:
			return p >= 3 && p <= 4
		}
	case "hard":
		switch r.NakedSingles {
		case 0:
			switch {
			case clues >= 32 && clues <= 34:
				return p >= 4 && p <= 6
			case clues >= 30 && clues <= 31:
				return p >= 5 && p <= 6
			}
		case 1:
			return clues >= 30 && clues <= 34 && p >= 5 && p <= 6
		}
	case "very hard":
		switch {
		case r.NakedSingles == 0:
			switch {
			case clues >= 28 && clues <= 30:
				return p >= 6 && p <= 9
			case clues >= 26 && clues <= 27:
				return p >= 7 && p <= 10
			}
		case r.NakedSingles <= 2:
			return clues >= 26 && clues <= 30 && p >= 7 && p <= 10
		}
	}
	return false
}

// GenerateSudokuPuzzleV3 generates a 9x9 puzzle with the production removal
// logic and the V2 clue windows, keeping only candidates whose V3 solve
// grades into the requested difficulty's table. Success means the grade
// matched exactly.
//
// Very hard retries until it truly succeeds — no near-miss is ever returned
// under that label. Other levels may fall back to the closest solvable
// candidate with Success=false if the cap is reached; that board is still
// guaranteed finishable under the level's naked-single budget.
func GenerateSudokuPuzzleV3(difficulty string) V3Result {
	cluesRange, passRange, nakedBudget, maxAttempts := v3Params(difficulty)
	fallbackOK := allowsFallback(difficulty)

	var best V3Result
	bestDist := 0
	haveFallback := false

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		solution := generateFullGrid()
		puzzle := cloneBoard(&solution)
		puzzle, _ = removeClues(puzzle, cluesRange[0], cluesRange[1])
		rating := RatePlayerSolveV3(puzzle, nakedBudget)
		if !rating.Solvable {
			continue
		}
		result := V3Result{
			Puzzle:        puzzle,
			Solution:      solution,
			Clues:         countClues(puzzle),
			Passes:        rating.Passes,
			HiddenSingles: rating.HiddenSingles,
			NakedSingles:  rating.NakedSingles,
			Attempts:      attempt,
		}
		if v3GradeMatches(difficulty, result.Clues, rating) {
			result.Success = true
			return result
		}
		if fallbackOK {
			// Rank near-misses: being outside the clue window is worse
			// than being off on passes, since the windows are the spec.
			dist := distToRange(result.Clues, cluesRange)*10 + distToRange(result.Passes, passRange)
			if !haveFallback || dist < bestDist {
				best, bestDist = result, dist
				haveFallback = true
			}
		}
	}

	if haveFallback {
		best.Attempts = maxAttempts
		return best
	}
	// Either the level forbids fallback, or no solvable board turned up at
	// all. Report an explicit failure rather than a puzzle that is the
	// wrong difficulty.
	return V3Result{Attempts: maxAttempts}
}
