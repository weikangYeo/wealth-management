package repository

import (
	"database/sql"
	"log"
	"wealth-management/domains"
)

// GoldRepository skip interface design here for now since repo still small (YNGNI)
type GoldRepository struct {
	// this is to hold the ref of sql connection that created during app start, to pool connections.
	db *sql.DB
}

// NewGoldRepository use constructor so the db connection stay private and won't be modified,
// else has to make it as Public to create FundRepo struct
func NewGoldRepository(db *sql.DB) GoldRepository {
	return GoldRepository{db: db}
}

func (repo *GoldRepository) GetAll() ([]domains.GoldTxn, error) {
	rows, err := repo.db.Query("SELECT * FROM gold_txn")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	goldTxns := make([]domains.GoldTxn, 0)
	for rows.Next() {
		var gold domains.GoldTxn
		if err := rows.Scan(&gold.ID, &gold.Bank, &gold.TxnDate, &gold.Gram, &gold.UnitPrice, &gold.TotalPrice, &gold.TxnType); err != nil {
			return nil, err
		}
		goldTxns = append(goldTxns, gold)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return goldTxns, nil

}

func (repo *GoldRepository) ReplaceAllByEntrySource(entrySource string, goldTxns []domains.GoldTxn) error {
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
		log.Printf("Debug - GoldTxn %#v\n", goldTxn)
		log.Printf("Debug - goldTxn.Gram.String() %s\n", goldTxn.Gram.String())
		log.Printf("Debug - goldTxn.UnitPrice.String() %s\n", goldTxn.UnitPrice.String())
		log.Printf("Debug - goldTxn.TotalPrice.String() %s\n", goldTxn.TotalPrice.String())
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
