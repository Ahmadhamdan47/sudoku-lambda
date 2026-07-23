package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func invoke(t *testing.T, body string) (int, map[string]interface{}) {
	t.Helper()
	resp, err := handler(context.Background(), events.APIGatewayProxyRequest{Body: body})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Body), &parsed); err != nil {
		t.Fatalf("response body is not JSON: %v (%s)", err, resp.Body)
	}
	return resp.StatusCode, parsed
}

// TestHandlerV1ShapeUnchanged guards the production contract: a request
// without "algorithm" must return exactly the fields it always has, with no
// V2 diagnostics leaking in.
func TestHandlerV1ShapeUnchanged(t *testing.T) {
	status, body := invoke(t, `{"difficulty":"medium","size":9}`)
	if status != 200 {
		t.Fatalf("status = %d, want 200", status)
	}
	want := []string{"puzzle", "solution", "clues", "difficulty", "size", "success"}
	for _, k := range want {
		if _, ok := body[k]; !ok {
			t.Errorf("v1 response missing %q", k)
		}
	}
	for _, k := range []string{"algorithm", "hiddenSingles", "nakedSingles", "attempts"} {
		if _, ok := body[k]; ok {
			t.Errorf("v1 response should not contain V2 field %q", k)
		}
	}
	if len(body) != len(want) {
		t.Errorf("v1 response has %d fields, want %d: %v", len(body), len(want), body)
	}
}

// TestHandlerV2Shape pins the documented V2 response.
func TestHandlerV2Shape(t *testing.T) {
	for _, d := range []string{"easy", "medium", "hard", "very hard"} {
		d := d
		t.Run(d, func(t *testing.T) {
			status, body := invoke(t, `{"difficulty":"`+d+`","size":9,"algorithm":"v2"}`)
			if status != 200 {
				t.Fatalf("status = %d, want 200", status)
			}
			for _, k := range []string{
				"puzzle", "solution", "clues", "difficulty", "size", "success",
				"algorithm", "hiddenSingles", "nakedSingles", "attempts",
			} {
				if _, ok := body[k]; !ok {
					t.Errorf("v2 response missing %q", k)
				}
			}
			if body["algorithm"] != "v2" {
				t.Errorf("algorithm = %v, want v2", body["algorithm"])
			}
			if body["size"].(float64) != 9 {
				t.Errorf("size = %v, want 9", body["size"])
			}
			puzzle, ok := body["puzzle"].([]interface{})
			if !ok || len(puzzle) != 9 {
				t.Fatalf("puzzle is not a 9-row array: %v", body["puzzle"])
			}
			if row, ok := puzzle[0].([]interface{}); !ok || len(row) != 9 {
				t.Errorf("puzzle row is not 9 wide: %v", puzzle[0])
			}
			if d == "very hard" && body["success"] != true {
				t.Errorf("very hard must always succeed, got success=%v", body["success"])
			}
		})
	}
}

func TestHandlerRejectsBadInput(t *testing.T) {
	cases := map[string]string{
		"unknown difficulty": `{"difficulty":"impossible","size":9}`,
		"bad size":           `{"difficulty":"easy","size":5}`,
		"unknown algorithm":  `{"difficulty":"easy","size":9,"algorithm":"v3"}`,
		"v2 with size 4":     `{"difficulty":"easy","size":4,"algorithm":"v2"}`,
		"malformed json":     `{"difficulty":`,
	}
	for name, body := range cases {
		name, body := name, body
		t.Run(name, func(t *testing.T) {
			status, _ := invoke(t, body)
			if status != 400 {
				t.Errorf("status = %d, want 400", status)
			}
		})
	}
}

// TestHandlerV1SmallSizes confirms 4x4 and 6x6 still route to the production
// generators and are unaffected by the V2 work.
func TestHandlerV1SmallSizes(t *testing.T) {
	for _, size := range []string{"4", "6"} {
		status, body := invoke(t, `{"difficulty":"easy","size":`+size+`}`)
		if status != 200 {
			t.Fatalf("size %s: status = %d, want 200", size, status)
		}
		if _, ok := body["algorithm"]; ok {
			t.Errorf("size %s: unexpected algorithm field", size)
		}
	}
}
