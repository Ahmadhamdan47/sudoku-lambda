package sudoku

import (
    "math/rand"
    "time"
    "fmt"
)

const N = 9

type Board [N][N]int

// ... (rest of the provided code)

func generateSudokuPuzzle(difficulty string) (Board, Board, int, bool) {
    cluesRange := map[string][2]int{
        "easy":      {36, 40},
        "medium":    {30, 35},
        "hard":      {25, 29},
        "very hard": {17, 24},
    }[difficulty]

    solution := generateFullGrid()
    puzzle := createMinimalPuzzle(solution)
    puzzle = applySymmetricRemoval(puzzle, cluesRange)

    return puzzle, solution, countClues(puzzle), true
}

func generateFullGrid() Board {
    var board Board
    board.solve()
    return board
}

func printBoard(b Board) {
    for i := 0; i < N; i++ {
        for j := 0; j < N; j++ {
            fmt.Printf("%2d ", b[i][j])
        }
        fmt.Println()
    }
}