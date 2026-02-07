package repository

import (
	"database/sql"
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
