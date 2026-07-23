package sudoku

import (
	"sort"
	"testing"
	"time"
)

// TestProfileVeryHardCost splits very hard's generation cost between building
// a full solved grid and the removal/rating retry loop, to show which one
// drives the latency tail.
func TestProfileVeryHardCost(t *testing.T) {
	if testing.Short() {
		t.Skip("profiling only")
	}
	const samples = 300
	grid := make([]time.Duration, 0, samples)
	remove := make([]time.Duration, 0, samples)
	cluesRange, _ := v2Params("very hard")

	for i := 0; i < samples; i++ {
		start := time.Now()
		solution := generateFullGrid()
		grid = append(grid, time.Since(start))

		start = time.Now()
		puzzle := cloneBoard(&solution)
		puzzle, _ = removeClues(puzzle, cluesRange[0], cluesRange[1])
		RatePlayerSolve(puzzle)
		remove = append(remove, time.Since(start))
	}

	report := func(name string, d []time.Duration) {
		sort.Slice(d, func(i, j int) bool { return d[i] < d[j] })
		var sum time.Duration
		for _, v := range d {
			sum += v
		}
		t.Logf("%-18s median=%-10v p95=%-10v p99=%-10v max=%-10v mean=%v",
			name,
			d[len(d)/2].Round(time.Microsecond),
			d[len(d)*95/100].Round(time.Microsecond),
			d[len(d)*99/100].Round(time.Microsecond),
			d[len(d)-1].Round(time.Microsecond),
			(sum / time.Duration(len(d))).Round(time.Microsecond))
	}
	report("generateFullGrid", grid)
	report("remove+rate", remove)
}
