// newalgorithm.go — V2 generator: new clue windows plus a player-model
// solvability test.
//
// The production generator (GenerateSudokuPuzzle) only guarantees a unique
// solution, which does not guarantee the puzzle is solvable by a human using
// singles. V2 differs from production in two ways:
//
//  1. New clue windows (givens per difficulty):
//       easy      38–42
//       medium    34–38
//       hard      30–34
//       very hard 26–30
//
//  2. Every candidate is solved exactly like a player would (naked + hidden
//     singles only, no pencil marks). A puzzle the singles solver cannot
//     finish is discarded and regenerated — this is what eliminates "broken"
//     puzzles that require guessing. A returned puzzle is therefore always
//     completable with the techniques the game supports.
//
// Very hard never falls back: it retries until a candidate genuinely lands in
// its window, so a puzzle carrying that label is always the real thing. Other
// levels may return the closest singles-solvable candidate with Success=false
// if the attempt cap is reached, so the admin screen can see the window was
// missed while the app still gets a playable board.
//
// Clue count alone does not determine difficulty — HiddenSingles is the better
// signal, since it counts placements that need a unit scan rather than being
// visible from the cell alone. It is reported so the levels can be compared on
// what they actually demand of a player.
//
// Note that pushing very hard much below this window stops being viable: see
// TestMinimalClueFloor, which finds that stripping boards as far as uniqueness
// allows bottoms out around 22 clues and that boards near that floor are never
// singles-solvable. Few clues is precisely what forces advanced techniques.
//
// Production code paths are untouched; this file only adds new symbols.
package sudoku

// PlayerRating is the outcome of solving a puzzle the way a player would:
// place every naked single on sight, and when none remain, scan units for a
// hidden single. No pairs, box-line, chains, or guessing.
type PlayerRating struct {
	Solvable      bool
	NakedSingles  int
	HiddenSingles int
}

// V2Result carries everything the admin test screen needs to inspect a
// generated puzzle. HiddenSingles/NakedSingles are diagnostics: how many
// placements were slow finds (unit scans) versus instant finds.
type V2Result struct {
	Puzzle        Board
	Solution      Board
	Clues         int
	HiddenSingles int
	NakedSingles  int
	Attempts      int
	Success       bool
}

// v2Params returns the clue window (givens) and the attempt cap for a
// difficulty. Windows are contiguous, so a puzzle sitting exactly on a
// boundary (30, 34, 38) is valid for either neighbouring level; the label
// comes from what was requested, not from the clue count alone.
func v2Params(difficulty string) (cluesRange [2]int, maxAttempts int) {
	switch difficulty {
	case "very hard":
		return [2]int{26, 30}, 200
	case "hard":
		return [2]int{30, 34}, 60
	case "medium":
		return [2]int{34, 38}, 60
	default: // easy
		return [2]int{38, 42}, 60
	}
}

// allowsFallback reports whether a level may return a near-miss when its
// window cannot be hit. Very hard never does: it retries until it genuinely
// succeeds, or reports failure. Its higher attempt cap reflects that.
func allowsFallback(difficulty string) bool {
	return difficulty != "very hard"
}

// candidateMask returns a bitmask of digits (bit d set = digit d possible)
// that can legally go in the given empty cell.
func candidateMask(b *Board, row, col int) uint16 {
	var used uint16
	for i := 0; i < N; i++ {
		used |= 1 << b[row][i]
		used |= 1 << b[i][col]
	}
	startRow, startCol := (row/3)*3, (col/3)*3
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			used |= 1 << b[startRow+i][startCol+j]
		}
	}
	return ^used & 0b1111111110 // bits 1..9
}

func soleCandidate(mask uint16) (int, bool) {
	if mask == 0 || mask&(mask-1) != 0 {
		return 0, false
	}
	for d := 1; d <= 9; d++ {
		if mask&(1<<d) != 0 {
			return d, true
		}
	}
	return 0, false
}

// findHiddenSingle scans rows, then columns, then boxes for a digit that fits
// in exactly one empty cell of the unit. Scan order is fixed so ratings are
// deterministic for a given puzzle.
func findHiddenSingle(b *Board) (row, col, num int, found bool) {
	for d := 1; d <= 9; d++ {
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
				return r, cc, d, true
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
				return rr, c, d, true
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
				return rr, cc, d, true
			}
		}
	}
	return 0, 0, 0, false
}

// RatePlayerSolve solves the puzzle like a player: sweep for naked singles
// until none remain, then place one hidden single and sweep again. Returns
// whether the puzzle completes and how many placements of each kind it took.
func RatePlayerSolve(puzzle Board) PlayerRating {
	b := puzzle
	var rating PlayerRating
	for {
		progress := false
		for i := 0; i < N; i++ {
			for j := 0; j < N; j++ {
				if b[i][j] == 0 {
					if num, ok := soleCandidate(candidateMask(&b, i, j)); ok {
						b[i][j] = num
						rating.NakedSingles++
						progress = true
					}
				}
			}
		}
		if progress {
			continue
		}
		if r, c, num, ok := findHiddenSingle(&b); ok {
			b[r][c] = num
			rating.HiddenSingles++
			continue
		}
		break
	}
	rating.Solvable = countClues(b) == N*N
	return rating
}

func distToRange(v int, window [2]int) int {
	if v < window[0] {
		return window[0] - v
	}
	if v > window[1] {
		return v - window[1]
	}
	return 0
}

// GenerateSudokuPuzzleV2 generates a 9x9 puzzle with the production removal
// logic and the V2 clue windows, keeping only candidates a player can finish
// with singles. Success means the clue count landed inside the window.
//
// Very hard retries until it truly succeeds — no near-miss is ever returned
// under that label. Other levels may fall back to the closest singles-solvable
// candidate with Success=false if the cap is reached; that board is still
// guaranteed playable, since candidates requiring guessing are never kept,
// even as fallback.
func GenerateSudokuPuzzleV2(difficulty string) V2Result {
	cluesRange, maxAttempts := v2Params(difficulty)
	fallbackOK := allowsFallback(difficulty)

	var best V2Result
	haveFallback := false

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		solution := generateFullGrid()
		puzzle := cloneBoard(&solution)
		puzzle, inRange := removeClues(puzzle, cluesRange[0], cluesRange[1])
		rating := RatePlayerSolve(puzzle)
		if !rating.Solvable {
			continue
		}
		result := V2Result{
			Puzzle:        puzzle,
			Solution:      solution,
			Clues:         countClues(puzzle),
			HiddenSingles: rating.HiddenSingles,
			NakedSingles:  rating.NakedSingles,
			Attempts:      attempt,
		}
		if inRange {
			result.Success = true
			return result
		}
		if fallbackOK && (!haveFallback || distToRange(result.Clues, cluesRange) < distToRange(best.Clues, cluesRange)) {
			best = result
			haveFallback = true
		}
	}

	if haveFallback {
		best.Attempts = maxAttempts
		return best
	}
	// Either the level forbids fallback, or no singles-solvable board turned
	// up at all. Report an explicit failure rather than a puzzle that is the
	// wrong difficulty or might require guessing.
	return V2Result{Attempts: maxAttempts}
}
