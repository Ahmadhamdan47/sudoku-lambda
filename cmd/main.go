package main

import (
    "context"
    "encoding/json"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/Ahmadhamdan47/sudoku-lambda/internal/sudoku"
)

type Response struct {
    Puzzle  sudoku.Board `json:"puzzle"`
    Solution sudoku.Board `json:"solution"`
    Clues   int          `json:"clues"`
    Difficulty string     `json:"difficulty"`
}

func handler(ctx context.Context, difficulty string) (Response, error) {
    puzzle, solution, clues, _ := sudoku.GenerateSudokuPuzzle(difficulty)
    return Response{
        Puzzle:    puzzle,
        Solution:  solution,
        Clues:     clues,
        Difficulty: difficulty,
    }, nil
}

func main() {
    lambda.Start(handler)
}