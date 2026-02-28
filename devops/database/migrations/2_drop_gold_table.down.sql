CREATE TABLE IF NOT EXISTS gold_txn
(
    id          BIGINT PRIMARY KEY NOT NULL,
    bank        VARCHAR(10),
    txn_date    DATE,
    gram        DECIMAL(10, 2),
    unit_price  DECIMAL(10, 2),
    total_price DECIMAL(12, 2),
    txn_type    VARCHAR(4)
);
