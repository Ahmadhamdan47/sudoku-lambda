package sudoku

import (
	"math/rand"
	"testing"
)

// minimizeBoard removes clues in repeated passes until no single clue can be
// dropped without losing uniqueness — minimal in the classic sense, ignoring
// whether a player could actually solve the result. Used only to measure the
// clue floor; the generator uses minimizeForSingles instead.
func minimizeBoard(board Board) Board {
	for {
		positions := make([][2]int, 0, N*N)
		for i := 0; i < N; i++ {
			for j := 0; j < N; j++ {
				if board[i][j] != 0 {
					positions = append(positions, [2]int{i, j})
				}
			}
		}
		rand.Shuffle(len(positions), func(i, j int) {
			positions[i], positions[j] = positions[j], positions[i]
		})
		removed := false
		for _, pos := range positions {
			i, j := pos[0], pos[1]
			if board[i][j] == 0 {
				continue
			}
			original := board[i][j]
			board[i][j] = 0
			if isUniqueSolution(&board) {
				removed = true
			} else {
				board[i][j] = original
			}
		}
		if !removed {
			return board
		}
	}
}

// TestMinimalClueFloor measures the lowest clue count reachable by exhaustive
// removal, and how often such minimal puzzles are singles-solvable. This is
// the evidence that a <=18-clue very-hard target is unreachable: minimization
// bottoms out around 22 clues, and boards at the floor are never solvable by
// singles alone.
func TestMinimalClueFloor(t *testing.T) {
	if testing.Short() {
		t.Skip("calibration only")
	}
	const samples = 60
	hist := map[int]int{}
	solvableAt := map[int]int{}
	min, max := 81, 0
	atOrBelow18, atOrBelow18Solvable := 0, 0

	for i := 0; i < samples; i++ {
		solution := generateFullGrid()
		puzzle := minimizeBoard(cloneBoard(&solution))
		clues := countClues(puzzle)
		hist[clues]++
		if clues < min {
			min = clues
		}
		if clues > max {
			max = clues
		}
		rating := RatePlayerSolve(puzzle)
		if rating.Solvable {
			solvableAt[clues]++
		}
		if clues <= 18 {
			atOrBelow18++
			if rating.Solvable {
				atOrBelow18Solvable++
			}
		}
	}

	t.Logf("minimal clue counts over %d exhaustively-minimized grids: min=%d max=%d", samples, min, max)
	for c := min; c <= max; c++ {
		if hist[c] > 0 {
			t.Logf("  clues=%2d  count=%2d  singlesSolvable=%d", c, hist[c], solvableAt[c])
		}
	}
	t.Logf("reached <=18 clues: %d/%d   of those singles-solvable: %d", atOrBelow18, samples, atOrBelow18Solvable)
}
