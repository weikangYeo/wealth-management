package fund

import "database/sql"

type repository struct {
	// this is to hold the ref of sql connection that created during app start, to pool connections.
	db *sql.DB
}

func newRepository(db *sql.DB) *repository {
	return &repository{db: db}
}

func (repo *repository) getAllFunds() ([]Fund, error) {
	rows, err := repo.db.Query("SELECT fund_code, name, url FROM fund_info")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	funds := make([]Fund, 0)
	for rows.Next() {
		var fund Fund
		if err := rows.Scan(&fund.FundCode, &fund.Name, &fund.URL); err != nil {
			return nil, err
		}
		funds = append(funds, fund)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return funds, nil
}
func (repo *repository) insertFund(fund Fund) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare("INSERT INTO fund_info(fund_code, name, url) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(fund.FundCode, fund.Name, fund.URL)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (repo *repository) getFundTxnByFundCode(fundCode string) ([]Txn, error) {
	rows, err := repo.db.Query("select id, fund_code,txn_date, unit, unit_price, sales_charge, gross_total_price, net_total_price, txn_type ,remark from fund_txn where fund_code=?",
		fundCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var fundTxns []Txn
	if rows.Next() {
		var fundTxn Txn
		if err := rows.Scan(&fundTxn.ID, &fundTxn.FundCode, &fundTxn.TxnType,
			&fundTxn.Unit, &fundTxn.UnitPrice, &fundTxn.SalesCharge,
			&fundTxn.GrossTotalPrice, &fundTxn.NetTotalPrice, &fundTxn.TxnType,
			&fundTxn.Remark); err != nil {
			return nil, err
		}
		fundTxns = append(fundTxns, fundTxn)
	}
	return fundTxns, nil
}
