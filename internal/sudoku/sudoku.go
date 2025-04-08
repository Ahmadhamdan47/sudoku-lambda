package sudoku

import (
    "fmt"
    "math/rand"

)

const N = 9

type Board [N][N]int

func isValid(board *Board, row, col, num int) bool {
    for i := 0; i < N; i++ {
        if board[row][i] == num || board[i][col] == num {
            return false
        }
    }
    startRow, startCol := (row/3)*3, (col/3)*3
    for i := 0; i < 3; i++ {
        for j := 0; j < 3; j++ {
            if board[startRow+i][startCol+j] == num {
                return false
            }
        }
    }
    return true
}

func solve(board *Board) bool {
    for row := 0; row < N; row++ {
        for col := 0; col < N; col++ {
            if board[row][col] == 0 {
                nums := rand.Perm(N)
                for _, v := range nums {
                    num := v + 1
                    if isValid(board, row, col, num) {
                        board[row][col] = num
                        if solve(board) {
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

func generateFullGrid() Board {
    var board Board
    solve(&board)
    return board
}

func cloneBoard(board *Board) Board {
    var newBoard Board
    for i := 0; i < N; i++ {
        for j := 0; j < N; j++ {
            newBoard[i][j] = board[i][j]
        }
    }
    return newBoard
}

func countClues(board Board) int {
    count := 0
    for i := 0; i < N; i++ {
        for j := 0; j < N; j++ {
            if board[i][j] != 0 {
                count++
            }
        }
    }
    return count
}

func countSolutions(board *Board, limit int) int {
    count := 0
    var helper func(b *Board)
    helper = func(b *Board) {
        if count >= limit {
            return
        }
        found := false
        var row, col int
        for i := 0; i < N && !found; i++ {
            for j := 0; j < N && !found; j++ {
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
        for num := 1; num <= 9; num++ {
            if isValid(b, row, col, num) {
                b[row][col] = num
                helper(b)
                b[row][col] = 0
            }
        }
    }
    temp := cloneBoard(board)
    helper(&temp)
    return count
}

func isUniqueSolution(board *Board) bool {
    return countSolutions(board, 2) == 1
}

func removeClues(board Board, cluesMin, cluesMax int) (Board, bool) {
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

        removalType := rand.Intn(2)
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

func GenerateSudokuPuzzle(difficulty string) (Board, Board, int, bool) {
    cluesRange := [2]int{38, 42} // Default to easy
    switch difficulty {
    case "medium":
        cluesRange = [2]int{33, 37}
    case "hard":
        cluesRange = [2]int{29, 36}
    case "very hard":
        cluesRange = [2]int{23, 30}
    }

    solution := generateFullGrid()
    puzzle := cloneBoard(&solution)
    puzzle, ok := removeClues(puzzle, cluesRange[0], cluesRange[1])
    return puzzle, solution, countClues(puzzle), ok
}

func printBoard(board Board) {
    for i := 0; i < N; i++ {
        for j := 0; j < N; j++ {
            fmt.Printf("%2d ", board[i][j])
        }
        fmt.Println()
    }
}

