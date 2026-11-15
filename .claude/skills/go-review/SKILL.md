---
name: go-review
description: Review the user's uncommitted backend Go changes in this repo (advice only, no edits). Use when the user asks to "review", "review again", or "how does it look now" for code they wrote in backend/.
---

# Go review (wealth-management backend)

The user writes this code themselves and wants a reviewer, not a co-author. **Do not edit their files.** Show
snippets in the reply instead; only change code if they explicitly ask.

## 1. Scope: read beyond the diff

1. `git status` and `git diff` (plus untracked files) under `backend/` to find what changed.
2. For every changed function, also read, even though they're unchanged:
   - its **callers** (handlers, other packages): how they build the arguments (nil vs zero value, `time.Now()`);
   - the **repository queries** that feed it: `ORDER BY` direction, `Scan` targets, `rows.Close()`;
   - its **tests**: do they still compile, and do they cover the new branches?
3. When library behaviour matters (e.g. which `apd` methods dereference the receiver), read the source in
   `$(go list -m -f '{{.Dir}}' <module>)` rather than guessing.

## 2. Run, don't just read

From `backend/`: `go build ./...`, `go vet ./...`, `go test ./...` (scoped to the changed package if the full suite
hits live sites). Report compile or test failures first, with the exact error.

## 3. Re-review mode

If this conversation already contains a review of the same code, open with what changed since then:
**fixed**, **regressed**, and **still open**. Don't re-explain points already fixed.

## 4. Checklist (issues that have actually occurred in this repo)

- **nil `*apd.Decimal`:** `String()`, `Float64()`, `Text()` dereference the receiver, so they panic on nil, including
  inside `MarshalJSON`. Nullable decimals must become `null` (a `*json.Number` helper), not be formatted.
- **Value vs pointer marshalling:** `apd.Decimal.MarshalText` has a pointer receiver; a struct passed by value to
  `json.Marshal`/`context.JSON` dumps `Form/Coeff`. Types with decimals need a value-receiver `MarshalJSON`.
- **"Unknown" encoded as 0:** a missing price, rate or ratio must be `nil` → JSON `null`, not a zero value. Watch map
  lookups without comma-ok and `&zeroValue`, which make a non-nil pointer to 0.
- **Ordering assumptions:** `txns[0]` or `dividends[0]` assumed to be oldest; repo queries are often `ORDER BY ... desc`.
- **Purity:** calculation functions (`overview.go`, `scraper.go` logic) must not call `time.Now()` or the repo; take
  `now` and data as parameters (see CLAUDE.md "Backend testing conventions").
- **N+1:** no per-item queries inside list loops; batch-load and group in Go.
- **`database/sql`:** `defer rows.Close()` after the `Query` error check, and `rows.Err()` after the loop.
- **`err` shadowing:** `:=` inside an inner block that should assign an outer variable.
- **Decimal arithmetic:** multiply before divide, and avoid the division entirely when algebra allows; fresh
  `new(apd.Decimal)` per loop iteration (CLAUDE.md "apd footgun"); `float64` only for estimates like XIRR.
- **Domain formulas:** check denominators and definitions against `docs/` and CLAUDE.md (e.g. ROI and payback divide
  by total invested, not the remaining cost basis; `TotalAmount` is cash moved, fee +/− by txn type).
- **Slices:** `make([]T, n)` followed by `append` (should be `make([]T, 0, n)`); nil slice → JSON `null` on list endpoints.

## 5. Output

- Open with one or two lines on what's good or newly fixed.
- Findings ordered by severity: 🔴 broken (crash, wrong numbers, doesn't compile) / 🟠 likely bug or misleading
  output / 🟡 design, testability / minor.
- Each finding: `file:line`, what's wrong, a **concrete failure case** (input → wrong output), and the fix as a
  short snippet.
- End with suggested test cases for the new branches when relevant, and the order to fix things in.
