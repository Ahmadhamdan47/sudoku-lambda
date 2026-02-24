package sudoku

import "math/rand"

const N4 = 4

// Board4 is a 4x4 sudoku board (2x2 boxes, numbers 1-4).
type Board4 [N4][N4]int

func isValid4(board *Board4, row, col, num int) bool {
	for i := 0; i < N4; i++ {
		if board[row][i] == num || board[i][col] == num {
			return false
		}
	}
	// 4x4 uses 2x2 boxes
	startRow, startCol := (row/2)*2, (col/2)*2
	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			if board[startRow+i][startCol+j] == num {
				return false
			}
		}
	}
	return true
}

func solve4(board *Board4) bool {
	for row := 0; row < N4; row++ {
		for col := 0; col < N4; col++ {
			if board[row][col] == 0 {
				nums := rand.Perm(N4)
				for _, v := range nums {
					num := v + 1
					if isValid4(board, row, col, num) {
						board[row][col] = num
						if solve4(board) {
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

func generateFullGrid4() Board4 {
	var board Board4
	solve4(&board)
	return board
}

func cloneBoard4(board *Board4) Board4 {
	var newBoard Board4
	for i := 0; i < N4; i++ {
		for j := 0; j < N4; j++ {
			newBoard[i][j] = board[i][j]
		}
	}
	return newBoard
}

func countClues4(board Board4) int {
	count := 0
	for i := 0; i < N4; i++ {
		for j := 0; j < N4; j++ {
			if board[i][j] != 0 {
				count++
			}
		}
	}
	return count
}

func countSolutions4(board *Board4, limit int) int {
	count := 0
	var helper func(b *Board4)
	helper = func(b *Board4) {
		if count >= limit {
			return
		}
		found := false
		var row, col int
		for i := 0; i < N4 && !found; i++ {
			for j := 0; j < N4 && !found; j++ {
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
		for num := 1; num <= N4; num++ {
			if isValid4(b, row, col, num) {
				b[row][col] = num
				helper(b)
				b[row][col] = 0
			}
		}
	}
	temp := cloneBoard4(board)
	helper(&temp)
	return count
}

func isUniqueSolution4(board *Board4) bool {
	return countSolutions4(board, 2) == 1
}

func removeClues4(board Board4, cluesMin, cluesMax int) (Board4, bool) {
	targetClues := cluesMin + rand.Intn(cluesMax-cluesMin+1)
	cluesCount := countClues4(board)

	positions := make([][2]int, 0, cluesCount)
	for i := 0; i < N4; i++ {
		for j := 0; j < N4; j++ {
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
			symI, symJ := N4-1-i, N4-1-j
			if board[symI][symJ] != 0 {
				original := board[i][j]
				symOriginal := board[symI][symJ]

				board[i][j] = 0
				board[symI][symJ] = 0

				if isUniqueSolution4(&board) {
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

		if isUniqueSolution4(&board) {
			cluesCount--
		} else {
			board[i][j] = original
		}
	}

	return board, cluesCount >= cluesMin && cluesCount <= cluesMax
}

// GenerateSudoku4Puzzle generates a 4x4 sudoku puzzle with a unique solution.
// Total cells: 16. Clue ranges by difficulty:
//   - easy:      10-12 clues
//   - medium:     8-9  clues
//   - hard:       6-7  clues
//   - very hard:  5-6  clues
func GenerateSudoku4Puzzle(difficulty string) (Board4, Board4, int, bool) {
	cluesRange := [2]int{10, 12} // default: easy
	switch difficulty {
	case "medium":
		cluesRange = [2]int{8, 9}
	case "hard":
		cluesRange = [2]int{6, 7}
	case "very hard":
		cluesRange = [2]int{5, 6}
	}

	solution := generateFullGrid4()
	puzzle := cloneBoard4(&solution)
	puzzle, ok := removeClues4(puzzle, cluesRange[0], cluesRange[1])
	return puzzle, solution, countClues4(puzzle), ok
}

// Board4ToSlice converts a Board4 to a [][]int slice.
func Board4ToSlice(b Board4) [][]int {
	result := make([][]int, N4)
	for i := range b {
		result[i] = make([]int, N4)
		for j := range b[i] {
			result[i][j] = b[i][j]
		}
	}
	return result
}
