package stock

import "database/sql"

type repository struct {
	db *sql.DB
}

func newStockRepository(db *sql.DB) *repository {
	return &repository{db: db}
}

func (r repository) getAllStocks() ([]Stock, error) {
	rows, err := r.db.Query("SELECT stock_name, display_name FROM stock order by stock_name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stocks := []Stock{}
	for rows.Next() {
		var stock Stock
		if err := rows.Scan(&stock.StockName, &stock.DisplayName); err != nil {
			return nil, err
		}
		stocks = append(stocks, stock)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stocks, nil
}

func (r repository) getStockByStockName(code string) (Stock, error) {
	var stock Stock
	err := r.db.QueryRow("SELECT stock_name, display_name, bursa_stock_id FROM stock where stock_name=? ", code).Scan(&stock.StockName, &stock.DisplayName, &stock.BursaStockId)
	if err != nil {
		return Stock{}, err
	}
	return stock, nil
}

func (r repository) createStock(stock Stock) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare("INSERT INTO stock (stock_name, display_name, bursa_stock_id) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	if _, err := stmt.Exec(stock.StockName, stock.DisplayName, stock.BursaStockId); err != nil {
		return err
	}
	return tx.Commit()
}

func (r repository) getStockTxnByStockName(stockName string) ([]Txn, error) {
	rows, err := r.db.Query("SELECT * from stock_txn where stock_name = ? order by txn_date desc", stockName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	txns := []Txn{}
	for rows.Next() {
		var txn Txn
		if err := rows.Scan(&txn.ID, &txn.StockName, &txn.TxnDate, &txn.Unit, &txn.UnitPrice,
			&txn.BrokerFee, &txn.TotalPrice, &txn.TxnType, &txn.Remark); err != nil {
			return nil, err
		}
		txns = append(txns, txn)
	}
	return txns, nil
}

func (r repository) createStockTxn(stockTxn Txn) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare("INSERT INTO stock_txn (id, stock_name, txn_date, unit, unit_price, broker_fee, total_price, txn_type, remark) values (?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	if _, err := stmt.Exec(stockTxn.ID, stockTxn.StockName, stockTxn.TxnDate, stockTxn.Unit, stockTxn.UnitPrice, stockTxn.BrokerFee, stockTxn.TotalPrice, stockTxn.TxnType, stockTxn.Remark); err != nil {
		return err
	}
	return tx.Commit()
}

func (r repository) getAllDividend() ([]Dividend, error) {
	rows, err := r.db.Query("SELECT * FROM dividend")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dividends := []Dividend{}
	for rows.Next() {
		var dividend Dividend
		if err := rows.Scan(&dividend.StockName, &dividend.TxnDate, &dividend.Amount); err != nil {
			return nil, err
		}
		dividends = append(dividends, dividend)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dividends, nil

}

func (r repository) createDividend(dividend Dividend) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare("INSERT INTO stock_dividend VALUES (?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	if _, err := stmt.Exec(dividend.StockName, dividend.TxnDate, dividend.Amount); err != nil {
		return err
	}
	return tx.Commit()
}
