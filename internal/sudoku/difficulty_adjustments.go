// difficulty_adjustments.go (place in the same package as the original code)
package sudoku

import (
    "math/rand"
)

// Override GenerateSudokuPuzzle to adjust easy/hard logic
func GenerateSudokuPuzzleWithAdjustments(difficulty string) (Board, Board, int, bool) {
    switch difficulty {
    case "easy", "hard":
        // Custom logic for easy/hard
        cluesRange := [2]int{38, 42}
        if difficulty == "hard" {
            cluesRange = [2]int{29, 36}
        }

        solution := generateFullGrid()
        puzzle := cloneBoard(&solution)
        puzzle, ok := removeCluesAdjusted(puzzle, cluesRange[0], cluesRange[1], difficulty)
        return puzzle, solution, countClues(puzzle), ok

    default:
        // Use original logic for medium/very hard
        return generateOriginalPuzzle(difficulty)
    }
}

// Original logic extracted for medium/very hard
func generateOriginalPuzzle(difficulty string) (Board, Board, int, bool) {
    cluesRange := [2]int{38, 42}
    switch difficulty {
    case "medium":
        cluesRange = [2]int{33, 37}
    case "very hard":
        cluesRange = [2]int{23, 30}
    }

    solution := generateFullGrid()
    puzzle := cloneBoard(&solution)
    puzzle, ok := removeClues(puzzle, cluesRange[0], cluesRange[1])
    return puzzle, solution, countClues(puzzle), ok
}

// Custom clue removal for easy/hard
func removeCluesAdjusted(board Board, cluesMin, cluesMax int, difficulty string) (Board, bool) {
    targetClues := cluesMin + rand.Intn(cluesMax-cluesMin+1)
    cluesCount := countClues(board)

    positions := make([][2]int, 0, cluesCount)
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

    for _, pos := range positions {
        if cluesCount <= targetClues {
            break
        }
        i, j := pos[0], pos[1]
        if board[i][j] == 0 {
            continue
        }

        // Adjust removal strategy based on difficulty
        removalType := rand.Intn(2)
        if difficulty == "hard" {
            removalType = 1 // Force asymmetric removal for harder puzzles
        } else if difficulty == "easy" {
            removalType = 0 // Prefer symmetric removal for easier puzzles
        }

        if removalType == 0 { // Symmetric removal
            symI, symJ := N-1-i, N-1-j
            if board[symI][symJ] != 0 {
                original := board[i][j]
                symOriginal := board[symI][symJ]

                board[i][j] = 0
                board[symI][symJ] = 0

                if isUniqueSolution(&board) {
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

        if isUniqueSolution(&board) {
            cluesCount--
        } else {
            board[i][j] = original
        }
    }

    return board, cluesCount >= cluesMin && cluesCount <= cluesMax
}