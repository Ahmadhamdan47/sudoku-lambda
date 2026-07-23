package sudoku

import (
	"sort"
	"testing"
)

// TestCalibrateWindows measures, for each V2 clue window, how often raw
// generation lands in the window, how often the result is singles-solvable,
// and both at once — i.e. the real acceptance rate of GenerateSudokuPuzzleV2.
func TestCalibrateWindows(t *testing.T) {
	if testing.Short() {
		t.Skip("calibration only")
	}
	levels := []string{"easy", "medium", "hard", "very hard"}
	t.Log("inRange = clue window hit; solvable = completable with singles; accepted = both")
	const samples = 150
	for _, name := range levels {
		cluesRange, _ := v2Params(name)
		inRange, solvable, both := 0, 0, 0
		minClues, maxClues := 81, 0
		for i := 0; i < samples; i++ {
			solution := generateFullGrid()
			puzzle := cloneBoard(&solution)
			puzzle, ok := removeClues(puzzle, cluesRange[0], cluesRange[1])
			clues := countClues(puzzle)
			if clues < minClues {
				minClues = clues
			}
			if clues > maxClues {
				maxClues = clues
			}
			rating := RatePlayerSolve(puzzle)
			if ok {
				inRange++
			}
			if rating.Solvable {
				solvable++
			}
			if ok && rating.Solvable {
				both++
			}
		}
		t.Logf("%-10s window=%v  cluesReached=%d-%d  inRange=%d%%  solvable=%d%%  accepted=%d%%",
			name, cluesRange, minClues, maxClues,
			inRange*100/samples, solvable*100/samples, both*100/samples)
	}
}

// TestDifficultySeparation reports the hidden-single distribution actually
// served per level. Hidden singles are the placements a player has to hunt
// for, so this — not clue count — is what makes the levels feel different.
func TestDifficultySeparation(t *testing.T) {
	if testing.Short() {
		t.Skip("calibration only")
	}
	const samples = 60
	for _, name := range []string{"easy", "medium", "hard", "very hard"} {
		hidden := make([]int, 0, samples)
		for i := 0; i < samples; i++ {
			hidden = append(hidden, GenerateSudokuPuzzleV2(name).HiddenSingles)
		}
		sort.Ints(hidden)
		mean := 0
		for _, v := range hidden {
			mean += v
		}
		t.Logf("%-10s hiddenSingles  min=%-3d p25=%-3d median=%-3d p75=%-3d max=%-3d mean=%d",
			name, hidden[0], hidden[samples/4], hidden[samples/2],
			hidden[samples*3/4], hidden[samples-1], mean/samples)
	}
}
