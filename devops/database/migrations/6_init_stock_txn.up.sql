CREATE TABLE IF NOT EXISTS stock
(
    stock_code   VARCHAR(32) PRIMARY KEY NOT NULL,
    display_name VARCHAR(32)
);

CREATE TABLE IF NOT EXISTS stock_txn
(
    id          VARCHAR(36) PRIMARY KEY NOT NULL,
    stock_code  VARCHAR(32)             NOT NULL,
    txn_date    DATE                    NOT NULL,
    unit        DECIMAL(10, 2)          NOT NULL,
    unit_price  DECIMAL(10, 2)          NOT NULL,
    broker_fee  DECIMAL(10, 2)          NOT NULL,
    total_price DECIMAL(10, 2)          NOT NULL,
    txn_type    VARCHAR(16)             NOT NULL,
    remark      VARCHAR(100),
    FOREIGN KEY (stock_code) REFERENCES stock (stock_code)
);

CREATE TABLE IF NOT EXISTS stock_dividend
(
    stock_code VARCHAR(32)    NOT NULL,
    txn_date   DATE           NOT NULL,
    amount     DECIMAL(10, 2) NOT NULL,
    PRIMARY KEY (stock_code, txn_date),
    FOREIGN KEY (stock_code) REFERENCES stock (stock_code)
);