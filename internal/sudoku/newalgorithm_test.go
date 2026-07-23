package sudoku

import "testing"

// TestGenerateSudokuPuzzleV2 verifies every difficulty produces a
// singles-solvable, unique-solution puzzle, and that levels with a clue
// window land inside it when they report success.
func TestGenerateSudokuPuzzleV2(t *testing.T) {
	difficulties := []string{"easy", "medium", "hard", "very hard"}
	for _, d := range difficulties {
		d := d
		t.Run(d, func(t *testing.T) {
			cluesRange, _ := v2Params(d)
			for i := 0; i < 5; i++ {
				result := GenerateSudokuPuzzleV2(d)
				if result.Clues == 0 {
					t.Fatalf("run %d: no singles-solvable puzzle produced in %d attempts", i, result.Attempts)
				}
				assertPlayable(t, i, result)
				if result.Success && (result.Clues < cluesRange[0] || result.Clues > cluesRange[1]) {
					t.Errorf("run %d: success=true but clues=%d outside window %v", i, result.Clues, cluesRange)
				}
				t.Logf("run %d: clues=%d (window %v) success=%v hidden=%d naked=%d attempts=%d",
					i, result.Clues, cluesRange, result.Success,
					result.HiddenSingles, result.NakedSingles, result.Attempts)
			}
		})
	}
}

// TestGenerateVeryHardV2NeverFallsBack checks the no-fallback contract: very
// hard must always report success with a board genuinely inside its window,
// never a near miss.
func TestGenerateVeryHardV2NeverFallsBack(t *testing.T) {
	cluesRange, _ := v2Params("very hard")
	for i := 0; i < 8; i++ {
		result := GenerateSudokuPuzzleV2("very hard")
		if !result.Success {
			t.Fatalf("run %d: very hard returned failure after %d attempts", i, result.Attempts)
		}
		if result.Clues < cluesRange[0] || result.Clues > cluesRange[1] {
			t.Errorf("run %d: clues=%d outside window %v — fallback leaked", i, result.Clues, cluesRange)
		}
		assertPlayable(t, i, result)
		t.Logf("run %d: clues=%d hidden=%d naked=%d attempts=%d",
			i, result.Clues, result.HiddenSingles, result.NakedSingles, result.Attempts)
	}
}

func assertPlayable(t *testing.T, run int, result V2Result) {
	t.Helper()
	if rating := RatePlayerSolve(result.Puzzle); !rating.Solvable {
		t.Errorf("run %d: returned puzzle not singles-solvable", run)
	}
	if !isUniqueSolution(&result.Puzzle) {
		t.Errorf("run %d: returned puzzle does not have a unique solution", run)
	}
	filled := result.Puzzle
	for r := 0; r < N; r++ {
		for c := 0; c < N; c++ {
			if filled[r][c] != 0 && filled[r][c] != result.Solution[r][c] {
				t.Errorf("run %d: clue at %d,%d disagrees with solution", run, r, c)
			}
		}
	}
}
