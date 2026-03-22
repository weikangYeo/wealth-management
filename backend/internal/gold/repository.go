package gold

import (
	"database/sql"
)

// repository skip interface design here for now since repo still small (YNGNI)
type repository struct {
	// this is to hold the ref of sql connection that created during app start, to pool connections.
	db *sql.DB
}

// newGoldRepository use constructor so the db connection stay private and won't be modified,
// else has to make it as Public to create FundRepo struct
func newGoldRepository(db *sql.DB) repository {
	return repository{db: db}
}

func (repo *repository) getAllTxn() ([]Txn, error) {
	rows, err := repo.db.Query("SELECT * FROM gold_txn")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	goldTxns := make([]Txn, 0)
	for rows.Next() {
		var gold Txn
		if err := rows.Scan(&gold.ID, &gold.Bank, &gold.TxnDate, &gold.Gram, &gold.UnitPrice, &gold.TotalPrice, &gold.TxnType, &gold.EntrySource); err != nil {
			return nil, err
		}
		goldTxns = append(goldTxns, gold)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return goldTxns, nil

}

func (repo *repository) replaceAllTxnByEntrySource(entrySource string, goldTxns []Txn) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	// when there is any error occur and return later, rollback would always call
	// in the event of `tx.commit()`, it been flushed and nothing else would be rolled back (no-opt)
	defer tx.Rollback()
	_, err = tx.Exec("DELETE FROM gold_txn WHERE entry_source=?", entrySource)
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO gold_txn VALUES (?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, goldTxn := range goldTxns {
		_, err = stmt.Exec(
			goldTxn.ID,
			goldTxn.Bank,
			goldTxn.TxnDate,
			goldTxn.Gram.String(),
			goldTxn.UnitPrice.String(),
			goldTxn.TotalPrice.String(),
			goldTxn.TxnType,
			goldTxn.EntrySource)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (repo *repository) insertOrUpdatePriceHistory(priceHistory PriceHistory) error {
	tx, err := repo.db.Begin()
	defer tx.Rollback()
	stmt, err := tx.Prepare("INSERT INTO gold_price_history VALUES (?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		priceHistory.date,
		priceHistory.buyPrice)
	if err != nil {
		return err
	}
	return tx.Commit()
}
