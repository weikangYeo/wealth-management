CREATE TABLE gold_txn
(
    id           VARCHAR(36) PRIMARY KEY NOT NULL,
    bank         VARCHAR(10),
    txn_date     DATE,
    gram         DECIMAL(10, 2),
    unit_price   DECIMAL(10, 2),
    total_price  DECIMAL(12, 2),
    txn_type     VARCHAR(4),
    entry_source VARCHAR(10)
);

