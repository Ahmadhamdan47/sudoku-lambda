package sudoku

import "testing"

// TestGenerateSudokuPuzzleV3 verifies every difficulty produces a puzzle
// that solves under the V3 player model within its naked-single budget, and
// that a success genuinely grades into the requested difficulty's table.
func TestGenerateSudokuPuzzleV3(t *testing.T) {
	difficulties := []string{"easy", "medium", "hard", "very hard"}
	for _, d := range difficulties {
		d := d
		t.Run(d, func(t *testing.T) {
			cluesRange, _, nakedBudget, _ := v3Params(d)
			for i := 0; i < 3; i++ {
				result := GenerateSudokuPuzzleV3(d)
				if result.Clues == 0 {
					t.Fatalf("run %d: no solvable puzzle produced in %d attempts", i, result.Attempts)
				}
				assertPlayableV3(t, i, d, result)
				if result.Success {
					rating := RatePlayerSolveV3(result.Puzzle, nakedBudget)
					if !v3GradeMatches(d, result.Clues, rating) {
						t.Errorf("run %d: success=true but grade does not match (clues=%d passes=%d naked=%d)",
							i, result.Clues, rating.Passes, rating.NakedSingles)
					}
				}
				t.Logf("run %d: clues=%d (window %v) passes=%d hidden=%d naked=%d success=%v attempts=%d",
					i, result.Clues, cluesRange, result.Passes,
					result.HiddenSingles, result.NakedSingles, result.Success, result.Attempts)
			}
		})
	}
}

// TestGenerateVeryHardV3NeverFallsBack checks the no-fallback contract: very
// hard must always report success with a board genuinely graded very hard.
func TestGenerateVeryHardV3NeverFallsBack(t *testing.T) {
	_, _, nakedBudget, _ := v3Params("very hard")
	for i := 0; i < 3; i++ {
		result := GenerateSudokuPuzzleV3("very hard")
		if !result.Success {
			t.Fatalf("run %d: very hard returned failure after %d attempts", i, result.Attempts)
		}
		rating := RatePlayerSolveV3(result.Puzzle, nakedBudget)
		if !v3GradeMatches("very hard", result.Clues, rating) {
			t.Errorf("run %d: grade does not match table (clues=%d passes=%d naked=%d) — fallback leaked",
				i, result.Clues, rating.Passes, rating.NakedSingles)
		}
		assertPlayableV3(t, i, "very hard", result)
		t.Logf("run %d: clues=%d passes=%d hidden=%d naked=%d attempts=%d",
			i, result.Clues, result.Passes, result.HiddenSingles, result.NakedSingles, result.Attempts)
	}
}

// TestRatePlayerSolveV3Deterministic pins that rating the same board twice
// gives identical numbers — the grade must be a property of the puzzle.
func TestRatePlayerSolveV3Deterministic(t *testing.T) {
	result := GenerateSudokuPuzzleV3("medium")
	if result.Clues == 0 {
		t.Fatal("no puzzle produced")
	}
	a := RatePlayerSolveV3(result.Puzzle, 0)
	b := RatePlayerSolveV3(result.Puzzle, 0)
	if a != b {
		t.Errorf("ratings differ across runs: %+v vs %+v", a, b)
	}
}

func assertPlayableV3(t *testing.T, run int, difficulty string, result V3Result) {
	t.Helper()
	_, _, nakedBudget, _ := v3Params(difficulty)
	if rating := RatePlayerSolveV3(result.Puzzle, nakedBudget); !rating.Solvable {
		t.Errorf("run %d: returned puzzle not solvable under the V3 model", run)
	}
	if !isUniqueSolution(&result.Puzzle) {
		t.Errorf("run %d: returned puzzle does not have a unique solution", run)
	}
	for r := 0; r < N; r++ {
		for c := 0; c < N; c++ {
			if result.Puzzle[r][c] != 0 && result.Puzzle[r][c] != result.Solution[r][c] {
				t.Errorf("run %d: clue at %d,%d disagrees with solution", run, r, c)
			}
		}
	}
}

// TestCalibrateV3 measures, per difficulty, how often a raw candidate in the
// clue window grades into the V3 table. It prints the observed pass/naked
// distributions so the tables and attempt caps can be judged from data.
// Measurement only — skipped under -short.
func TestCalibrateV3(t *testing.T) {
	if testing.Short() {
		t.Skip("calibration is a measurement, skipped in -short")
	}
	const samples = 300
	for _, d := range []string{"easy", "medium", "hard", "very hard"} {
		cluesRange, _, nakedBudget, _ := v3Params(d)
		matched, solvable := 0, 0
		passCount := map[int]int{}
		nakedCount := map[int]int{}
		for i := 0; i < samples; i++ {
			solution := generateFullGrid()
			puzzle := cloneBoard(&solution)
			puzzle, _ = removeClues(puzzle, cluesRange[0], cluesRange[1])
			rating := RatePlayerSolveV3(puzzle, nakedBudget)
			if !rating.Solvable {
				continue
			}
			solvable++
			passCount[rating.Passes]++
			nakedCount[rating.NakedSingles]++
			if v3GradeMatches(d, countClues(puzzle), rating) {
				matched++
			}
		}
		t.Logf("%-9s window=%v budget=%d  solvable=%d/%d  graded=%d/%d (%.1f%%)  passes=%v naked=%v",
			d, cluesRange, nakedBudget, solvable, samples, matched, samples,
			100*float64(matched)/float64(samples), passCount, nakedCount)
	}
}
