package fund

import (
	"database/sql"
	"time"
	"wealth-management/internal/platform/decimal"

	"github.com/cockroachdb/apd/v3"
)

type repository struct {
	// this is to hold the ref of sql connection that created during app start, to pool connections.
	db *sql.DB
}

func newRepository(db *sql.DB) *repository {
	return &repository{db: db}
}

func (repo *repository) getAllFunds() ([]Fund, error) {
	rows, err := repo.db.Query("SELECT fund_code, name, fsm_url, scrape_param_value, COALESCE(provider, '') FROM fund_info")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	funds := make([]Fund, 0)
	for rows.Next() {
		var fund Fund
		if err := rows.Scan(&fund.FundCode, &fund.Name, &fund.FsmURL, &fund.ScrapeParamValue, &fund.Provider); err != nil {
			return nil, err
		}
		funds = append(funds, fund)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return funds, nil
}

func (repo *repository) getFundInfo(fundCode string) (Fund, error) {
	var fund Fund
	err := repo.db.QueryRow("SELECT fund_code, name, fsm_url FROM fund_info where fund_code=? ", fundCode).
		Scan(&fund.FundCode, &fund.Name, &fund.FsmURL)
	if err != nil {
		return fund, err
	}
	return fund, nil
}

func (repo *repository) insertFund(fund Fund) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare("INSERT INTO fund_info(fund_code, name, fsm_url, scrape_param_value, provider) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(fund.FundCode, fund.Name, fund.FsmURL, fund.ScrapeParamValue, fund.Provider)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *repository) getFundTxnByFundCode(fundCode string) ([]Txn, error) {
	rows, err := repo.db.Query(`select id, fund_code,txn_date, unit, unit_price, sales_charge,
    	net_investment_amount, total_amount, txn_type ,remark
		from fund_txn where fund_code=?`,
		fundCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fundTxns := make([]Txn, 0)
	for rows.Next() {
		var fundTxn Txn
		if err := rows.Scan(&fundTxn.ID, &fundTxn.FundCode, &fundTxn.TxnDate,
			&fundTxn.Unit, &fundTxn.UnitPrice, &fundTxn.SalesCharge,
			&fundTxn.NetInvestmentAmount, &fundTxn.TotalAmount, &fundTxn.TxnType,
			&fundTxn.Remark); err != nil {
			return nil, err
		}
		fundTxns = append(fundTxns, fundTxn)
	}
	return fundTxns, nil
}

func (repo *repository) insertFundTxn(fundTxn Txn) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`insert into fund_txn(id, fund_code,txn_date, unit,
                     unit_price, sales_charge, net_investment_amount, total_amount,
                     txn_type ,remark) values (?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(fundTxn.ID, fundTxn.FundCode, fundTxn.TxnDate, fundTxn.Unit, fundTxn.UnitPrice,
		fundTxn.SalesCharge, fundTxn.NetInvestmentAmount, fundTxn.TotalAmount, fundTxn.TxnType, fundTxn.Remark)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *repository) insertIgnoreFundPriceHistory(priceHistory PriceHistory) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare("insert ignore into fund_price_history(fund_code, price_date, nav) values (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(priceHistory.FundCode, priceHistory.PriceDate, priceHistory.Nav)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *repository) getOldestFundTxn(fundCode string) (Txn, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return Txn{}, err
	}
	defer tx.Rollback()
	var fundTxn Txn
	err = tx.QueryRow("SELECT id, fund_code, txn_date, unit, unit_price, sales_charge, net_investment_amount, total_amount, txn_type, remark FROM fund_txn WHERE fund_code=? ORDER BY txn_date ASC LIMIT 1",
		fundCode).Scan(&fundTxn.ID, &fundTxn.FundCode, &fundTxn.TxnDate,
		&fundTxn.Unit, &fundTxn.UnitPrice, &fundTxn.SalesCharge, &fundTxn.NetInvestmentAmount,
		&fundTxn.TotalAmount, &fundTxn.TxnType, &fundTxn.Remark)
	if err != nil {
		return Txn{}, err
	}
	return fundTxn, nil
}

func (repo *repository) getLatestReinvestmentTxn(fundCode string) (Txn, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return Txn{}, err
	}
	defer tx.Rollback()
	var fundTxn Txn
	err = tx.QueryRow("SELECT id, fund_code, txn_date, unit, unit_price, sales_charge, net_investment_amount, total_amount, txn_type, remark FROM fund_txn WHERE fund_code=? AND txn_type =? ORDER BY txn_date ASC LIMIT 1",
		fundCode, "REINVESTED").Scan(&fundTxn.ID, &fundTxn.FundCode, &fundTxn.TxnDate,
		&fundTxn.Unit, &fundTxn.UnitPrice, &fundTxn.SalesCharge, &fundTxn.NetInvestmentAmount,
		&fundTxn.TotalAmount, &fundTxn.TxnType, &fundTxn.Remark)
	if err != nil {
		return Txn{}, err
	}
	return fundTxn, nil
}

func (repo *repository) getTotalUnitToDate(fundCode string, txnDate time.Time) (*apd.Decimal, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// todo use sum(unit and -unit) base on txn_type (BUY, SELL, REINVESTED)
	var totalStr string
	err = tx.QueryRow(`SELECT COALESCE(SUM(CASE WHEN txn_type = 'SELL' THEN -unit ELSE unit END), '0')
			FROM fund_txn WHERE fund_code = ? and txn_date <= ?`, fundCode, txnDate).Scan(&totalStr)
	if err != nil {
		return nil, err
	}
	total, err := decimal.ToDecimal(totalStr)
	if err != nil {
		return nil, err
	}
	return total, nil
}
