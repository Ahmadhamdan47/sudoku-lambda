# Sudoku Puzzle Generator

This project is an AWS Lambda function that generates Sudoku puzzles. It utilizes Go for implementation and can be invoked via Amazon API Gateway.

## Project Structure

```
sudoku-lambda
├── cmd
│   └── main.go          # Entry point for the AWS Lambda function
├── internal
│   ├── sudoku
│   │   └── sudoku.go    # Contains the Sudoku generation and solving logic
├── go.mod                # Go module definition
├── go.sum                # Checksums for module dependencies
└── README.md             # Project documentation
```

## Setup Instructions

1. **Clone the Repository**
   ```bash
   git clone <repository-url>
   cd sudoku-lambda
   ```

2. **Install Dependencies**
   Ensure you have Go installed, then run:
   ```bash
   go mod tidy
   ```

3. **Deploy to AWS Lambda**
   Use the AWS CLI or your preferred deployment method to deploy the Lambda function defined in `cmd/main.go`.

4. **Set Up API Gateway**
   Create an API Gateway that triggers the Lambda function. Configure the necessary routes to call the Sudoku puzzle generation endpoint.

## Usage

Once deployed, you can invoke the Lambda function via the API Gateway endpoint. The function will return a generated Sudoku puzzle along with its solution.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.