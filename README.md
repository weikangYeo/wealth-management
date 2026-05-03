# Wealth Management

## TODO

- [ ] Makefile
- [ ] Dashboard
- [ ] Drag & Drop Widget
- [ ] Widget
    - [ ] P&L
    - [ ] annualized return
    - [ ] calendar return
    - [ ] Investment type
        - [ ] Gold
        - [ ] Funds
        - [ ] Stock (CARD BASE)
            - [ ] Total Dividends, Unrealized Profit, Realized Profit, annualized returned, average DY
        - [ ] Cash
- [ ] Modules
    - [X] Gold
        - [X] Import/Export
            - [X] Upsert Gold Txn
        - [X] View and aggregated info
        - [ ] Scrape gold price by banks
    - [ ] Item by Item CRUD
    - [ ] Stock
        - [ ] Portfolio overview, group by sector (30 % bank, 20 % Tech etc)
        - [ ] Aggregated all stock info
          - [x] total dividends
          - [ ] DY
        - [x] Scrape Stock Price
            - [x] Scrape Stock Price from bursa
            - [x] Scrape Stock Price from klse screener
            - [x] set Scrape Stock Price from klse screener
        - [X] Scrape Dividend info, so dont have to manual insert dividend every time
            - [X] add withholding tax if it is REIT
        - [X] Dividend Input fields (Stepper component)
        - [ ] Dividend Graph over the years
        - [ ] Capital Grow over the years/anuallized return per year since this stock is purchased ?
        - [ ] GET Stock API to include DY or other aggregated info I interested the most
        - [ ] Design a mechanism to calculate the aggregated info, on demand vs job vs app start vs etc...
- [ ] Onboard Data (to this system) and back up
- [ ] Watch list
- [ ] Revamp UI
- [ ] "Fun" Part - Web Scrapper
    - [ ] Funds Info & Price
    - [ ] Stock info
    - [X] Stock info & Price
    - [X] Gold Price
    - [ ] News that might relate
    - [ ] The ETL process of web scraped data and how to process it
- [ ] Based on scrapped info, feed to LLM to provide suggestion
    - [ ] compare LLM rules and custom defined (by dev) rules
- [X] Setup DB and schema migration (golang-migrate/migrate or Goose or GORM)

## Note for future self

- Current/Next item to work with
    - Stock Detail - get stock summary
    - Stock metrics calculation

### File structure

this repo is using "Package by feature" way to seperate packages.

```
backend/
    ├── cmd/
    │   └── scrapper/
    │       └── main.go             # Entry point for scrapper
    │   └── server/
    │       └── main.go             # Entry point for web BE
    ├── internal/                   # Code not meant to be imported by other projects
    │   ├── gold/                   # FEATURE: Gold Management
    │   │   ├── handler.go          # HTTP endpoints (was gold_handler.go)
    │   │   ├── repository.go       # DB operations (was gold_repo.go)
    │   │   ├── model.go            # Data structures (was domains/gold.go)
    │   │   ├── scrapper.go         # Scraping logic (was gold_price_scrapper.go)
    │   │   └── routes.go           # Feature-specific route registration
    │   │
    │   ├── system/                 # FEATURE: System/Health
    │   │   └── ping.go             # (was ping_route.go)
    │   │
    │   └── platform/               # SHARED TOOLS/INFRASTRUCTURE
    │       ├── scraper/
    │       │   └── cookies.go      # Shared scraper utils (was cookies-harvestor.go)
    │       └── database/           # DB connection/init logic
    ├── go.mod
    └── go.sum
```

## Start project

- run `docker-compose up` in `/devops` folder
- create a file call `.env` in `backend` directory, which following content

```
DBUSER=<replace with your username>
DBPASS=<replace with your password>
GOLD_URL=<Url to scrape gold price> // now it only support 1 local
STOCK_URL=<Url to scrape stock price> // now it only support bursa company profile
```

- `go run cmd/scrapper/main.go ` to start scrapper logic
- `go` run cmd/server/main.go` to start web app logic 