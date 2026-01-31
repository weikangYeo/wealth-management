# Wealth Management

## TODO
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
  - [ ] Import/Export
    - [ ] Upsert
  - [ ] Item by Item CRUD
- [ ] Watch list
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
### File structure for Gin 
Directory
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