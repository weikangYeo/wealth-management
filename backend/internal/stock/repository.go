package stock

import "database/sql"

type repository struct {
	db *sql.DB
}

func newStockRepository(db *sql.DB) *repository {
	return &repository{db: db}
}

func (r repository) getAllStocks() ([]Stock, error) {
	rows, err := r.db.Query("SELECT stock_code, display_name FROM stock order by stock_code")
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

func (r repository) createStock(stock Stock) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare("INSERT INTO stock (stock_code, display_name) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	if _, err := stmt.Exec(stock.StockCode, stock.DisplayName); err != nil {
		return err
	}
	return tx.Commit()
}

func (r repository) getStockTxnByStockCode(stockCode string) ([]Txn, error) {
	rows, err := r.db.Query("SELECT * from stock_txn where stock_code = ? order by txn_date desc", stockCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	txns := []Txn{}
	for rows.Next() {
		var txn Txn
		if err := rows.Scan(&txn.ID, &txn.StockCode, &txn.TxnDate, &txn.Unit, &txn.UnitPrice,
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
	stmt, err := tx.Prepare("INSERT INTO stock_txn (id, stock_code, txn_date, unit, unit_price, broker_fee, total_price, txn_type, remark) values (?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	if _, err := stmt.Exec(stockTxn.ID, stockTxn.StockCode, stockTxn.TxnDate, stockTxn.Unit, stockTxn.UnitPrice, stockTxn.BrokerFee, stockTxn.TotalPrice, stockTxn.TxnType, stockTxn.Remark); err != nil {
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
		if err := rows.Scan(&dividend.StockCode, &dividend.TxnDate, &dividend.Amount); err != nil {
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
	if _, err := stmt.Exec(dividend.StockCode, dividend.TxnDate, dividend.Amount); err != nil {
		return err
	}
	return tx.Commit()
}
