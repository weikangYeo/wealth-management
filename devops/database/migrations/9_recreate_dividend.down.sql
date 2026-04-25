CREATE TABLE IF NOT EXISTS stock_dividend
(
    stock_name VARCHAR(32)    NOT NULL,
    txn_date   DATE           NOT NULL,
    amount     DECIMAL(10, 2) NOT NULL,
    PRIMARY KEY (stock_name, txn_date),
    FOREIGN KEY (stock_name) REFERENCES stock (stock_name)
);