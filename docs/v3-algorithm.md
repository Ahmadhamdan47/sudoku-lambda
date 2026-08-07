# V3 Puzzle Generator — Integration Guide

V3 is a **separate, opt-in code path**, exactly like V2. Production (`v1`) and V2 are untouched; a request that does not ask for V3 runs exactly the code it runs today. To use it, add `"algorithm": "v3"` to the request body. That is the whole integration.

---

## Why V3 exists

V2 guarantees a puzzle is *finishable* with singles. V3 additionally grades *how much work* the solve takes, using a stricter player model, and only accepts puzzles whose workload matches the requested difficulty. This is the mechanism the V2 doc flagged as missing: under V2, easy and medium were statistically indistinguishable in play. V3 fixes that by grading on passes.

### The V3 player model

- The solver **holds one digit at a time**: it places every hidden single that digit has (re-scanning until the digit is exhausted), then moves to the next digit. A full 1→9 cycle is **one pass**.
- **Naked singles are never used freely.** They are an escape hatch only: when a full pass places nothing (the player is stuck), exactly one naked single may be placed — if the level's budget allows — and hidden-single passes resume.
- A stuck pass (a full 1→9 scan that finds nothing) still counts as a pass.
- Naked-single budgets: easy **0**, medium **0**, hard **1**, very hard **2**.

If the solver gets stuck with no naked single available, or with the budget already spent, the candidate is discarded immediately.

### The grading tables

Every candidate is graded on (clues, passes, naked singles used). No match → discard and regenerate.

| Level | Clues | Passes | Naked singles |
|---|---|---|---|
| easy | 40–42 | exactly 2 | 0 |
| easy | 38–39 | 2–3 | 0 |
| medium | 36–38 | exactly 3 | 0 |
| medium | 34–35 | 3–4 | 0 |
| hard | 32–34 | 4–6 | 0 |
| hard | 30–31 | 5–6 | 0 |
| hard | 30–34 | 5–6 | exactly 1 (only when stuck) |
| very hard | 28–30 | 6–9 | 0 |
| very hard | 26–27 | 7–10 | 0 |
| very hard | 26–30 | 7–10 | 1–2 (only when stuck) |

Clue windows are the same as V2: easy 38–42, medium 34–38, hard 30–34, very hard 26–30.

**Very hard never falls back** — same contract as V2. It retries until a candidate genuinely grades into its table. Other levels may return the closest solvable candidate with `success: false` if the attempt cap is reached; that board is still guaranteed finishable under the level's naked budget.

### Measured acceptance rates (`TestCalibrateV3`, 300 samples per level)

| Level | Raw candidates that grade in | Attempt cap |
|---|---|---|
| easy | 83% | 60 |
| medium | 51% | 60 |
| hard | 26% | 100 |
| very hard | 18% | 300 |

Expected attempts are single digits everywhere; at 18% per attempt, exhausting very hard's cap of 300 is a statistical impossibility. Generation stays well within V2's latency envelope.

---

## Frontend integration

### Request

```jsonc
{
  "difficulty": "very hard",  // "easy" | "medium" | "hard" | "very hard"
  "size": 9,                  // V3 requires 9
  "algorithm": "v3"
}
```

### Response

Same shape as V2, plus `passes`:

```json
{
  "puzzle":        [[9,0,2, "..."]],
  "solution":      [[9,3,2, "..."]],
  "clues":         28,
  "difficulty":    "very hard",
  "size":          9,
  "success":       true,
  "algorithm":     "v3",
  "passes":        8,
  "hiddenSingles": 51,
  "nakedSingles":  2,
  "attempts":      3
}
```

| Field | Meaning |
|---|---|
| `passes` | Full 1→9 hidden-single cycles the solve took. **The difficulty grade.** |
| `hiddenSingles` | Total hidden singles placed. |
| `nakedSingles` | Naked singles spent on stuck passes (0 unless hard/very hard). |
| `attempts` | Candidates generated before one graded in. Tuning diagnostic. |
| `success` | Grade matched the table. As with V2, a `false` (easy/medium/hard fallback) board is still finishable — render it; surface the flag on the admin screen only. Very hard is always `true`. |

Error handling is identical to V2: unknown algorithm, or `v3` with size 4/6, returns HTTP 400.

---

## AWS deployment

Identical to V2 — code deploy only, no configuration changes. Build and deploy per the [V2 guide](v2-algorithm.md#aws-deployment). Verify with:

```bash
aws lambda invoke --function-name <name> \
  --payload '{"body":"{\"difficulty\":\"very hard\",\"size\":9,\"algorithm\":\"v3\"}"}' \
  --cli-binary-format raw-in-base64-out out.json && cat out.json
```

The response should include `"algorithm":"v3"`, `"passes"`, and `"success":true`.

---

## Tests

```bash
go test ./...                                              # everything
go test ./cmd/ -run TestHandlerV3Shape -v                  # API contract
go test ./internal/sudoku/ -run TestGenerateSudokuPuzzleV3 -v
go test ./internal/sudoku/ -run TestGenerateVeryHardV3 -v  # no-fallback guarantee
go test ./internal/sudoku/ -run TestCalibrateV3 -v         # acceptance rates (skipped in -short)
```

**Re-run `TestCalibrateV3` after changing any window or grade table**, since acceptance rate drives generation time.

## Where the code lives

| File | Role |
|---|---|
| [internal/sudoku/v3algorithm.go](../internal/sudoku/v3algorithm.go) | The entire V3 generator. New symbols only. |
| [internal/sudoku/newalgorithm.go](../internal/sudoku/newalgorithm.go) | V2. Unchanged (V3 reuses its `candidateMask`/`soleCandidate` helpers). |
| [cmd/main.go](../cmd/main.go) | Request routing; V3 branch is additive. |

Key functions: `GenerateSudokuPuzzleV3` (entry point), `RatePlayerSolveV3` (the digit-cycling player model), `v3GradeMatches` (the acceptance tables), `v3Params` (windows, budgets, caps).
