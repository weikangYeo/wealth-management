# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

A personal wealth-management app tracking Gold, Stocks, and Unit Trust Funds. Backend is Go (Gin + MySQL), frontend is
Angular + Material + Tailwind. A separate scraper binary pulls prices/dividends/NAV from external sources (Bursa, KLSE
Screener, fund houses) using chromedp (headless Chrome) since some sources have bot detection.

## Commands

### First-time setup

- Start MySQL: `cd devops && docker-compose up` (exposes MySQL on `localhost:3307`, db name `wealth_management`)
- Create `backend/.env` with:
  ```
  DBUSER=<db username>
  DBPASS=<db password>
  GOLD_URL=<url to scrape gold price>
  STOCK_URL=<url to scrape stock price (bursa company profile)>
  KLSE_STOCK_BASE_URL=<url to scrape klse index price>
  ```
- DB schema migrations run automatically on server startup (`golang-migrate`, files in `devops/database/migrations`).
  There is no separate migrate CLI invocation in this repo — just start the server.

### Backend (`cd backend` first — required, since the module uses `internal/`)

- Run web server: `go run cmd/server/main.go` (port 8080; runs DB migrations on boot)
- Run scraper: `go run cmd/scraper/main.go` (does NOT run migrations; currently only calls
  `fund.ScrapeFundNavAndIncomeDist`)
- Build: `go build ./...`
- Test: `go test ./...` (no tests exist yet in the codebase)
- Vet/format: `go vet ./...`, `gofmt -l .`

### Frontend (`cd frontend` first)

- Dev server: `npm start` (`ng serve`)
- Build: `npm run build`
- Test: `npm test` (`ng test`, via vitest)

### Run everything

- `./start-all.sh` — brings up docker DB, then runs `ng serve` and the Go server together, tears both down on Ctrl+C.
- `start-scraper.sh` — brings up docker DB, then runs the scraper binary.

## Architecture

### Backend: package-by-feature

```
backend/
  cmd/
    server/main.go     # Gin HTTP API entry point; runs DB migrations, sets up CORS for http://localhost:4200
    scraper/main.go     # standalone scraper entry point (no HTTP server, no migrations)
  internal/
    fund/               # feature: Unit Trust Funds
    gold/                # feature: Gold holdings
    stock/               # feature: Stock holdings
    admin/               # health/ping route
    platform/            # shared infra, not feature-specific
      database/          # MySQL connection + migration runner
      config/             # .env loading via godotenv
      decimal/             # apd.Decimal parsing helpers (money uses cockroachdb/apd, not float)
      scrape/              # chromedp-based HTML fetch helper shared by scrapers
```

Each feature package (`fund`, `gold`, `stock`) follows the same internal shape: `model.go` (structs), `repository.go`
(raw SQL against `*sql.DB`), `handler.go` (Gin handlers), `route.go`/`routes.go` (registers routes on the shared
`*gin.Engine`), `scraper.go` (feature-specific scraping/business logic). Routes are wired centrally in
`internal/app/routes.go`.

Money/quantity values use `github.com/cockroachdb/apd/v3` (arbitrary-precision decimal) throughout models and
calculations — never plain `float64` for prices/units/amounts.

### Fund scraping: provider strategy pattern

`internal/fund/provider/provider.go` defines the `FundDataProvider` interface ( `FetchNavByDate`,
`FetchIncomeDistribution`). Each fund house gets its own implementation under `internal/fund/provider/` (e.g.
`aham.go`). `fund.ScrapeFundNavAndIncomeDist` (in `internal/fund/scraper.go`) looks up the right provider by the fund's
`Provider` column and:

1. fetches latest NAV and inserts into price history,
2. fetches income distributions since the last processed date,
3. computes reinvested units (dividend payout ÷ NAV on payment date) and inserts a `REINVESTED` transaction per
   distribution.

This is incremental — it assumes prior transactions were computed correctly and only pulls/computes forward from the
last known state (see `getIncomeDistPullStartDate`). When adding a new fund house provider, implement `FundDataProvider`
and register it in the `scraperByProvider` map in `scraper.go`.

### Fund transaction amounts: netInvestmentAmount vs totalAmount

`Txn` (`internal/fund/model.go`) has two derived money fields that are easy to conflate:

- `netInvestmentAmount` — value of the units transacted, always `unit × unitPrice`. Represents "how much of this
  actually became fund units," regardless of transaction type.
- `totalAmount` — actual cash that moved for the transaction: `netInvestmentAmount + salesCharge` on `BUY` (a fee is
  paid on top), `netInvestmentAmount − salesCharge` on `SELL` (a redemption fee reduces proceeds), or `0` on
  `REINVESTED` (a dividend reinvestment has no fresh cash in/out — the "cash" is the fund's own distribution,
  redirected into more units).

Both are computed together in `Txn.CalculateTxnTotals`, which branches on `TxnType`. `REINVESTED` transactions built
directly in `scraper.go` (not via the handler) set both fields inline rather than calling `CalculateTxnTotals`, so keep
the two in sync if the formula changes. `netInvestmentAmount` is what feeds "avg unit price" style calculations
(unaffected by fees); `totalAmount` is what feeds "cash invested" calculations (excludes reinvestment).

### Scraping infra

`internal/platform/scrape.GetHtmlStringFromUrl` drives headless Chrome via chromedp to fetch a page's rendered HTML
(used for sources that need JS execution or have bot/challenge detection), including basic detection of
challenge/captcha pages in the response. Feature scrapers (`gold`, `stock`, fund providers) build on this rather than
issuing raw HTTP requests, since target sites often block naive scraping.

### Database

MySQL, schema managed via numbered `golang-migrate` migration pairs (`N_description.up.sql` / `.down.sql`) in
`devops/database/migrations`. Migrations are applied automatically by `database.InitDbConnection(true)` when the server
starts (the scraper binary calls `InitDbConnection(false)` and does not run migrations). Table names are lowercased
(`--lower_case_table_names=1` in docker-compose).

### Frontend

Angular (standalone components, no NgModules), Material for UI components, Tailwind for utility styling. Feature folders
under `frontend/src/app/` (`gold-mgmt`, `stock-mgmt`, `funds-mgmt`) each hold a `*.service.ts` (HTTP calls to the Go
backend), `*.model.ts` (TS interfaces mirroring backend structs), and page components. `interceptors/api.interceptor.ts`
handles outgoing API request concerns centrally.

## Notes

- README.md's TODO list is the closest thing to a living roadmap/backlog for this project — check it for planned work
  and known design gaps (e.g. fund scraper still needs a recompute job, per-fund-house providers beyond AHAM, portfolio
  overview aggregation).
- On Windows, clone into a native WSL path (`~/...`), not a `/mnt/c/...` mount — I/O is much slower on the mount point.
- The scraper depends on a locally installed Chrome/Chromium (`chromedp` drives an existing browser, it doesn't bundle
  one) — if scraper runs fail with an exec/launch error, check that a browser is installed before debugging the
  scraping logic itself.
