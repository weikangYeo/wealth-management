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
        - [ ] Stock
        - [ ] Cash
- [ ] CRUD
    - [X] Import/Export
        - [X] Upsert Gold Txn
    - [ ] Item by Item CRUD
- [ ] Watch list
- [ ] Revamp UI 
- [ ] "Fun" Part - Web Scrapper
    - [ ] Funds Info & Price
    - [ ] Stock info & Price
    - [ ] Gold info & Price
    - [ ] News that might relate
    - [ ] The ETL process of web scraped data and how to process it
- [ ] Based on scrapped info, feed to LLM to provide suggestion
    - [ ] compare LLM rules and custom defined (by dev) rules
- [ ] Setup DB and schema migration (golang-migrate/migrate or Goose or GORM)

## Note for future self
- Current/Next item to work with
  - Gold Management UI 
    - list to show all current txn 
      - current total cost, grams, market value (placeholder) 
    - UI of bulk import 

### File structure for Gin

Directory

prvious, package by layers
```
    ├── go.mod
    ├── go.sum
    ├── main.go
    ├── handlers/
    │   ├── ping_handler.go
    │   ├── stock_handler.go
    │   ├── fund_handler.go
    │   └── dashboard_handler.go
    └── routes/
        ├── routes.go              # The main router setup file
        ├── ping_routes.go
        ├── stock_routes.go
        ├── fund_routes.go
        └── dashboard_routes.go
```

tobe, package by feature
```
backend/
    ├── cmd/
    │   └── server/
    │       └── main.go             # Entry point: Wires features together
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
```