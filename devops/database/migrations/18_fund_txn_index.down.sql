ALTER TABLE fund_txn
    DROP INDEX idx_fund_txn_code_date_type,
    ADD INDEX fund_code (fund_code);
