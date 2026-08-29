ALTER TABLE fund_txn RENAME COLUMN gross_total_price TO net_investment_amount;
ALTER TABLE fund_txn RENAME COLUMN net_total_price TO total_amount;
