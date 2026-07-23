package sudoku

import (
	"sort"
	"testing"
	"time"
)

// TestV2Latency measures wall-clock generation time per difficulty so the
// Lambda timeout and memory size can be chosen from data rather than guessed.
func TestV2Latency(t *testing.T) {
	if testing.Short() {
		t.Skip("latency measurement only")
	}
	for _, d := range []string{"easy", "medium", "hard", "very hard"} {
		samples := 30
		if d == "very hard" {
			samples = 25
		}
		durations := make([]time.Duration, 0, samples)
		maxAttempts := 0
		failures := 0
		for i := 0; i < samples; i++ {
			start := time.Now()
			result := GenerateSudokuPuzzleV2(d)
			durations = append(durations, time.Since(start))
			if result.Attempts > maxAttempts {
				maxAttempts = result.Attempts
			}
			if !result.Success {
				failures++
			}
		}
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		p50 := durations[len(durations)/2]
		p95 := durations[len(durations)*95/100]
		worst := durations[len(durations)-1]
		t.Logf("%-10s n=%d  p50=%-8v p95=%-8v max=%-8v  maxAttempts=%d  failures=%d",
			d, samples, p50.Round(time.Millisecond), p95.Round(time.Millisecond),
			worst.Round(time.Millisecond), maxAttempts, failures)
	}
}
