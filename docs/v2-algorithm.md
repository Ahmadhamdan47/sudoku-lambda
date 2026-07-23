# V2 Puzzle Generator — Integration Guide

The V2 generator is a **separate, opt-in code path**. The production generator is untouched: a request that does not ask for V2 runs exactly the code that runs today and returns exactly the fields it returns today. This is enforced by a test (`TestHandlerV1ShapeUnchanged`), so V2 cannot leak into production responses by accident.

To use V2, add `"algorithm": "v2"` to the request body. That is the whole integration.

---

## Why V2 exists

The production generator only checks that a puzzle has **exactly one solution**. That is not the same as the puzzle being **solvable by a human**. A puzzle can have a unique solution that is only reachable by guessing or chain logic, which is how the broken puzzles (74JYN0, 94648q, YDt0QA) shipped — all three had a unique solution.

V2 adds a player test. After generating a candidate, it solves the puzzle exactly the way a player does: naked singles and hidden singles only, no pencil marks, no pairs, no X-wing, no guessing. If that solver cannot finish, the puzzle is thrown away and a new one generated.

**Every puzzle V2 returns is guaranteed completable by number-checking alone.** That is the guarantee production does not have, and it is the fix for the broken-puzzle problem.

How much work this does varies by level. At 26–30 clues (very hard), **32% of generated candidates are not solvable by singles** — roughly one in three would have been a broken puzzle under production rules. At easy, essentially none are.

---

## Difficulty levels

| Level | Clue window | Window hit on first try | Falls back? |
|---|---|---|---|
| easy | 38–42 | 97% | yes |
| medium | 34–38 | 94% | yes |
| hard | 30–34 | 89% | yes |
| very hard | 26–30 | 68% | **never** |

Windows are contiguous, so a puzzle sitting exactly on a boundary (30, 34, 38) is valid for either neighbouring level. The label comes from what was requested, not inferred from the clue count.

**Very hard never falls back.** It retries until a candidate genuinely lands in its window, so a puzzle carrying that label is always the real thing. Its attempt cap is 200; at a 68% success rate per attempt, exhausting that is a statistical impossibility. Other levels return the closest singles-solvable candidate with `success: false` if their cap is reached — still fully playable, just slightly off-target.

### A caveat worth knowing: clue count is a weak difficulty signal

Measured hidden-single counts for puzzles actually served (`TestDifficultySeparation`, 60 samples each). Hidden singles are placements a player has to *hunt* for by scanning a row, column, or box; naked singles are visible from the cell alone:

| Level | median | mean | max |
|---|---|---|---|
| easy | 0 | 0 | 4 |
| medium | 0 | 0 | 9 |
| hard | 0 | 2 | 12 |
| very hard | 4 | 5 | 19 |

Easy and medium are statistically indistinguishable — both solve almost entirely with instant finds. Hard is only slightly harder. Only very hard separates clearly.

This confirms the original analysis: **clue count does not control difficulty.** Removing four clues does not reliably make a puzzle harder, because the remaining clues may still make every placement obvious. Search effort does, and that is what `hiddenSingles` measures.

The windows above are implemented as specified, and every puzzle is guaranteed solvable — the broken-puzzle fix works regardless. But if easy and medium should *feel* different in play, the clue window alone will not achieve it. The mechanism for that is a hidden-singles floor on top of the clue window (for example, medium requires ≥1, hard ≥3, very hard ≥6, rejecting candidates below it). That is a small, contained change to `v2Params` and the acceptance check in `GenerateSudokuPuzzleV2`. The `hiddenSingles` field is in the response so this can be judged from real puzzles first.

### Why very hard is not lower than 26

Pushing much below this window stops being viable. Measured directly (`TestMinimalClueFloor`): 60 grids were stripped as far as they could possibly go — repeated removal passes until no clue at all could be dropped without destroying the unique solution.

- Lowest clue count reached: **22**. Mean 24. Never below 22, not once in 60 runs.
- The sparsest boards (22 clues) were **never** solvable by singles.

Sub-20-clue puzzles do exist (17 is the proven minimum) but are found only by specialized search burning years of CPU time, and none are solvable by singles alone. Few clues is precisely what forces advanced techniques, so "very few clues" and "solvable by number-checking" pull against each other. 26–30 sits at a good point on that trade-off.

---

## Frontend integration

### Endpoint

Unchanged — the same API Gateway endpoint the app already calls. `POST`, JSON body.

### Request

```jsonc
{
  "difficulty": "very hard",  // "easy" | "medium" | "hard" | "very hard"
  "size": 9,                  // 4 | 6 | 9 — V2 requires 9
  "algorithm": "v2"           // omit or "v1" for production; "v2" for the new generator
}
```

`algorithm` is optional. Omitting it is identical to today's behavior.

### Response — production (`v1`, or omitted)

Unchanged from today:

```json
{
  "puzzle":     [[0,0,2, ...], ...],
  "solution":   [[7,6,2, ...], ...],
  "clues":      36,
  "difficulty": "hard",
  "size":       9,
  "success":    true
}
```

### Response — V2

Same fields, plus diagnostics:

```json
{
  "puzzle":        [[9,0,2, ...], ...],
  "solution":      [[9,3,2, ...], ...],
  "clues":         28,
  "difficulty":    "very hard",
  "size":          9,
  "success":       true,
  "algorithm":     "v2",
  "hiddenSingles": 4,
  "nakedSingles":  49,
  "attempts":      1
}
```

### Field reference

| Field | Type | Meaning |
|---|---|---|
| `puzzle` | `int[9][9]` | The board to render. `0` = empty cell for the player to fill. |
| `solution` | `int[9][9]` | The completed grid. Used for validation / reveal. |
| `clues` | `int` | Count of non-zero cells in `puzzle`. |
| `difficulty` | `string` | Echoes the request. |
| `size` | `int` | Always `9` for V2. |
| `success` | `bool` | See below — **not** a "did it work" flag. |
| `algorithm` | `string` | `"v2"`. Absent on production responses. |
| `hiddenSingles` | `int` | Placements needing a unit scan. **The real difficulty signal.** |
| `nakedSingles` | `int` | Placements visible from the cell alone (instant finds). |
| `attempts` | `int` | Candidates generated before one was accepted. Tuning diagnostic. |

### Reading `success` correctly

This is the one field with a subtlety worth getting right.

`success` does **not** mean "the puzzle is valid." Every puzzle V2 returns — regardless of `success` — is verified to have a unique solution and to be completable with singles. It is always safe to hand to a player.

`success: false` means only that the clue count landed outside the requested window. It is possible for easy, medium and hard; **very hard always returns `true`**.

**Player-facing app:** ignore `success` and render the puzzle. Handle HTTP status codes instead.
**Admin test screen:** surface `success` — it tells you a window needs tuning.

### Example

```js
async function fetchPuzzle(difficulty, { useV2 = false } = {}) {
  const res = await fetch(ENDPOINT, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      difficulty,
      size: 9,
      ...(useV2 && { algorithm: 'v2' }),
    }),
  });

  if (!res.ok) {
    const { error } = await res.json().catch(() => ({}));
    throw new Error(error ?? `Puzzle request failed (${res.status})`);
  }

  return res.json(); // { puzzle, solution, clues, ... }
}
```

For the admin screen, the extra fields give a difficulty readout reflecting what the puzzle actually demands:

```js
const p = await fetchPuzzle('very hard', { useV2: true });

console.log(`${p.clues} clues`);
console.log(`${p.hiddenSingles} hidden singles (slow finds)`);
console.log(`${p.nakedSingles} naked singles (instant)`);
console.log(`generated in ${p.attempts} attempt(s)`);
if (!p.success) console.warn('clue count outside target window');
```

A good admin comparison is requesting the same difficulty from `v1` and `v2` side by side. Clue counts will look similar; the difference is that every V2 puzzle is guaranteed finishable, and `hiddenSingles` quantifies how much hunting it takes.

### Error responses

All return HTTP `400` with `{"error": "..."}`:

| Cause | Example |
|---|---|
| Malformed JSON body | `{"difficulty":` |
| Unknown difficulty | `{"difficulty": "impossible"}` |
| Size not 4, 6, or 9 | `{"size": 5}` |
| Unknown algorithm | `{"algorithm": "v3"}` |
| V2 with size 4 or 6 | `{"size": 4, "algorithm": "v2"}` |

V2 is 9x9 only. The 4x4 and 6x6 generators are untouched and still served by the production path.

---

## AWS deployment

### The short version

**No AWS configuration changes are required.** No new IAM permissions, no environment variables, no API Gateway changes — the integration is a Lambda proxy, so the new `algorithm` field passes through in the request body automatically. This is a code deploy only.

### Build

This project uses the `provided.al2` custom runtime (the deployment artifact is a binary named `bootstrap`), so the binary must be built for Linux and named exactly `bootstrap`:

```bash
# From the repo root. Match GOARCH to the function's configured architecture.
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap ./cmd
zip bootstrap.zip bootstrap
```

On Windows PowerShell:

```powershell
$env:GOOS = 'linux'; $env:GOARCH = 'amd64'; $env:CGO_ENABLED = '0'
go build -tags lambda.norpc -o bootstrap ./cmd
Compress-Archive -Path bootstrap -DestinationPath bootstrap.zip -Force
```

If the function runs on Graviton, use `GOARCH=arm64`. A mismatch here produces a runtime error on invoke, not a build error, so it is worth confirming against the function's `Architectures` setting.

### Deploy

```bash
aws lambda update-function-code \
  --function-name <your-function-name> \
  --zip-file fileb://bootstrap.zip
```

### Lambda configuration

Measured generation time on a developer machine (`TestV2Latency`):

| Level | p50 | p95 | max |
|---|---|---|---|
| easy | <1 ms | 1 ms | 1 ms |
| medium | 1 ms | 2 ms | 2 ms |
| hard | 1 ms | 4 ms | 10 ms |
| very hard | 7 ms | 72 ms | 276 ms |

Very hard is slowest because it discards roughly a third of candidates and retries — that is the broken-puzzle filter doing its job. Even so, the worst case is well under a third of a second.

Lambda CPU scales with memory: at 128 MB you get a small fraction of a vCPU, and a full vCPU arrives around 1,769 MB. These numbers come from a dev machine, so a low-memory function will be several times slower.

- **Memory: 512 MB is comfortable**, 1024 MB if you want headroom. Because Lambda bills GB-seconds, more memory finishing proportionally faster often costs the same or less — worth measuring against real traffic.
- **Timeout: 30 seconds** is more than sufficient. If called through API Gateway, note that API Gateway enforces its own hard **29-second** integration limit that cannot be raised. Nothing measured comes remotely close to either.

### Verify after deploying

```bash
# Production path must be unchanged
aws lambda invoke --function-name <name> \
  --payload '{"body":"{\"difficulty\":\"hard\",\"size\":9}"}' \
  --cli-binary-format raw-in-base64-out out.json && cat out.json

# V2 path
aws lambda invoke --function-name <name> \
  --payload '{"body":"{\"difficulty\":\"very hard\",\"size\":9,\"algorithm\":\"v2\"}"}' \
  --cli-binary-format raw-in-base64-out out.json && cat out.json
```

The second response should include `"algorithm":"v2"` and `"success":true`.

### Rollout

Because V2 is opt-in per request, rollout is controlled entirely from the client — no infrastructure work, no separate deployment, no traffic splitting:

1. Deploy the new code. Nothing changes for existing users; V2 is dormant until requested.
2. Have the admin exercise V2 through the app's normal puzzle flow and compare against production.
3. When satisfied, flip the app to send `"algorithm": "v2"`, or promote V2 to the default with a one-line change in the handler.

**Rollback** is stopping sending `algorithm: "v2"`. No redeploy needed.

---

## Running the tests locally

```bash
go test ./...                                             # everything
go test ./cmd/ -v                                         # API contract, incl. v1-unchanged guard
go test ./internal/sudoku/ -run TestGenerateVeryHard -v    # very hard no-fallback guarantee
go test ./internal/sudoku/ -run TestCalibrateWindows -v    # acceptance rates per window
go test ./internal/sudoku/ -run TestDifficultySeparation -v # hidden-single spread per level
go test ./internal/sudoku/ -run TestV2Latency -v           # timing table above
go test ./internal/sudoku/ -run TestMinimalClueFloor -v    # the clue-floor evidence
```

The calibration and latency tests are measurements rather than pass/fail assertions — they print tables and are skipped under `-short`. **Re-run `TestCalibrateWindows` after changing any clue window**, since acceptance rate drives generation time.

## Where the code lives

| File | Role |
|---|---|
| [internal/sudoku/newalgorithm.go](../internal/sudoku/newalgorithm.go) | The entire V2 generator. New symbols only — nothing existing is modified. |
| [internal/sudoku/sudoku.go](../internal/sudoku/sudoku.go) | Production 9x9 generator. Unchanged. |
| [cmd/main.go](../cmd/main.go) | Request routing; V2 branch is additive. |

Key functions: `GenerateSudokuPuzzleV2` (entry point), `RatePlayerSolve` (the player model — the core of the fix), `v2Params` (clue windows).

To change a clue window, edit `v2Params` — it is the single source of truth, and the tests read from it rather than hardcoding values.
