package stock

import "database/sql"

type repository struct {
	db *sql.DB
}

func newStockRepository(db *sql.DB) repository {
	return repository{db: db}
}

func (r repository) getAllStocks() ([]Stock, error) {
	rows, err := r.db.Query("SELECT * FROM stock")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stocks := []Stock{}
	for rows.Next() {
		var stock Stock
		if err := rows.Scan(&stock.StockCode, &stock.DisplayName); err != nil {
			return nil, err
		}
		stocks = append(stocks, stock)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stocks, nil
}
