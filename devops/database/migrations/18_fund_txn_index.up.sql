ALTER TABLE fund_txn
    DROP INDEX fund_code,
    ADD INDEX idx_fund_txn_code_date_type (fund_code, txn_date, txn_type);
