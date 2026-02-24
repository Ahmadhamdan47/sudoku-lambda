package sudoku

import "math/rand"

const N6 = 6

// Board6 is a 6x6 sudoku board (2x3 boxes, numbers 1-6).
type Board6 [N6][N6]int

func isValid6(board *Board6, row, col, num int) bool {
	for i := 0; i < N6; i++ {
		if board[row][i] == num || board[i][col] == num {
			return false
		}
	}
	// 6x6 uses 2x3 boxes (2 rows, 3 cols per box)
	startRow, startCol := (row/2)*2, (col/3)*3
	for i := 0; i < 2; i++ {
		for j := 0; j < 3; j++ {
			if board[startRow+i][startCol+j] == num {
				return false
			}
		}
	}
	return true
}

func solve6(board *Board6) bool {
	for row := 0; row < N6; row++ {
		for col := 0; col < N6; col++ {
			if board[row][col] == 0 {
				nums := rand.Perm(N6)
				for _, v := range nums {
					num := v + 1
					if isValid6(board, row, col, num) {
						board[row][col] = num
						if solve6(board) {
							return true
						}
						board[row][col] = 0
					}
				}
				return false
			}
		}
	}
	return true
}

func generateFullGrid6() Board6 {
	var board Board6
	solve6(&board)
	return board
}

func cloneBoard6(board *Board6) Board6 {
	var newBoard Board6
	for i := 0; i < N6; i++ {
		for j := 0; j < N6; j++ {
			newBoard[i][j] = board[i][j]
		}
	}
	return newBoard
}

func countClues6(board Board6) int {
	count := 0
	for i := 0; i < N6; i++ {
		for j := 0; j < N6; j++ {
			if board[i][j] != 0 {
				count++
			}
		}
	}
	return count
}

func countSolutions6(board *Board6, limit int) int {
	count := 0
	var helper func(b *Board6)
	helper = func(b *Board6) {
		if count >= limit {
			return
		}
		found := false
		var row, col int
		for i := 0; i < N6 && !found; i++ {
			for j := 0; j < N6 && !found; j++ {
				if b[i][j] == 0 {
					row, col = i, j
					found = true
				}
			}
		}
		if !found {
			count++
			return
		}
		for num := 1; num <= N6; num++ {
			if isValid6(b, row, col, num) {
				b[row][col] = num
				helper(b)
				b[row][col] = 0
			}
		}
	}
	temp := cloneBoard6(board)
	helper(&temp)
	return count
}

func isUniqueSolution6(board *Board6) bool {
	return countSolutions6(board, 2) == 1
}

func removeClues6(board Board6, cluesMin, cluesMax int) (Board6, bool) {
	targetClues := cluesMin + rand.Intn(cluesMax-cluesMin+1)
	cluesCount := countClues6(board)

	positions := make([][2]int, 0, cluesCount)
	for i := 0; i < N6; i++ {
		for j := 0; j < N6; j++ {
			if board[i][j] != 0 {
				positions = append(positions, [2]int{i, j})
			}
		}
	}
	rand.Shuffle(len(positions), func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	for _, pos := range positions {
		if cluesCount <= targetClues {
			break
		}
		i, j := pos[0], pos[1]
		if board[i][j] == 0 {
			continue
		}

		removalType := rand.Intn(2)
		if removalType == 0 { // Symmetric removal
			symI, symJ := N6-1-i, N6-1-j
			if board[symI][symJ] != 0 {
				original := board[i][j]
				symOriginal := board[symI][symJ]

				board[i][j] = 0
				board[symI][symJ] = 0

				if isUniqueSolution6(&board) {
					cluesCount -= 2
				} else {
					board[i][j] = original
					board[symI][symJ] = symOriginal
				}
				continue
			}
		}

		// Asymmetric removal
		original := board[i][j]
		board[i][j] = 0

		if isUniqueSolution6(&board) {
			cluesCount--
		} else {
			board[i][j] = original
		}
	}

	return board, cluesCount >= cluesMin && cluesCount <= cluesMax
}

// GenerateSudoku6Puzzle generates a 6x6 sudoku puzzle with a unique solution.
// Total cells: 36. Clue ranges by difficulty:
//   - easy:      22-26 clues
//   - medium:    18-21 clues
//   - hard:      14-17 clues
//   - very hard: 10-13 clues
func GenerateSudoku6Puzzle(difficulty string) (Board6, Board6, int, bool) {
	cluesRange := [2]int{22, 26} // default: easy
	switch difficulty {
	case "medium":
		cluesRange = [2]int{18, 21}
	case "hard":
		cluesRange = [2]int{14, 17}
	case "very hard":
		cluesRange = [2]int{10, 13}
	}

	solution := generateFullGrid6()
	puzzle := cloneBoard6(&solution)
	puzzle, ok := removeClues6(puzzle, cluesRange[0], cluesRange[1])
	return puzzle, solution, countClues6(puzzle), ok
}

// Board6ToSlice converts a Board6 to a [][]int slice.
func Board6ToSlice(b Board6) [][]int {
	result := make([][]int, N6)
	for i := range b {
		result[i] = make([]int, N6)
		for j := range b[i] {
			result[i][j] = b[i][j]
		}
	}
	return result
}
