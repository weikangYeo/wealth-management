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
