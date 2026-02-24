package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Ahmadhamdan47/sudoku-lambda/internal/sudoku"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Response struct {
	Puzzle     [][]int `json:"puzzle"`
	Solution   [][]int `json:"solution"`
	Clues      int     `json:"clues"`
	Difficulty string  `json:"difficulty"`
	Size       int     `json:"size"`
	Success    bool    `json:"success"`
}

type Request struct {
	Difficulty string `json:"difficulty"`
	Size       int    `json:"size"`
}

// boardToSlice converts the default 9x9 Board to a [][]int slice.
func boardToSlice(b sudoku.Board) [][]int {
	result := make([][]int, 9)
	for i := range b {
		result[i] = make([]int, 9)
		for j := range b[i] {
			result[i][j] = b[i][j]
		}
	}
	return result
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var parsedReq Request
	if err := json.Unmarshal([]byte(req.Body), &parsedReq); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf(`{"error": "invalid request body: %v"}`, err),
		}, nil
	}

	validDifficulties := map[string]bool{
		"easy":      true,
		"medium":    true,
		"hard":      true,
		"very hard": true,
	}

	if !validDifficulties[parsedReq.Difficulty] {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf(`{"error": "invalid difficulty level: %s"}`, parsedReq.Difficulty),
		}, nil
	}

	if parsedReq.Size == 0 {
		parsedReq.Size = 9
	}

	validSizes := map[int]bool{4: true, 6: true, 9: true}
	if !validSizes[parsedReq.Size] {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf(`{"error": "invalid size: %d, must be 4, 6, or 9"}`, parsedReq.Size),
		}, nil
	}

	var puzzleSlice, solutionSlice [][]int
	var clues int
	var success bool

	switch parsedReq.Size {
	case 4:
		puzzle4, solution4, c, ok := sudoku.GenerateSudoku4Puzzle(parsedReq.Difficulty)
		puzzleSlice = sudoku.Board4ToSlice(puzzle4)
		solutionSlice = sudoku.Board4ToSlice(solution4)
		clues, success = c, ok
	case 6:
		puzzle6, solution6, c, ok := sudoku.GenerateSudoku6Puzzle(parsedReq.Difficulty)
		puzzleSlice = sudoku.Board6ToSlice(puzzle6)
		solutionSlice = sudoku.Board6ToSlice(solution6)
		clues, success = c, ok
	default:
		puzzle9, solution9, c, ok := sudoku.GenerateSudokuPuzzle(parsedReq.Difficulty)
		puzzleSlice = boardToSlice(puzzle9)
		solutionSlice = boardToSlice(solution9)
		clues, success = c, ok
	}

	response := Response{
		Puzzle:     puzzleSlice,
		Solution:   solutionSlice,
		Clues:      clues,
		Difficulty: parsedReq.Difficulty,
		Size:       parsedReq.Size,
		Success:    success,
	}

	jsonResp, _ := json.Marshal(response)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonResp),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}
