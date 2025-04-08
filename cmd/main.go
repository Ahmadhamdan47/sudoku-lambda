package main

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/Ahmadhamdan47/sudoku-lambda/internal/sudoku"
)

type Response struct {
    Puzzle     sudoku.Board `json:"puzzle"`
    Solution   sudoku.Board `json:"solution"`
    Clues      int          `json:"clues"`
    Difficulty string       `json:"difficulty"`
    Success    bool         `json:"success"`
}

type Request struct {
    Difficulty string `json:"difficulty"`
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

    puzzle, solution, clues, success := sudoku.GenerateSudokuPuzzle(parsedReq.Difficulty)

    response := Response{
        Puzzle:     puzzle,
        Solution:   solution,
        Clues:      clues,
        Difficulty: parsedReq.Difficulty,
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
