CREATE TABLE gold_txn
(
    id           varchar(36) primary key not null,
    bank         varchar(10),
    txn_date     date,
    gram         decimal(10, 2),
    unit_price   decimal(10, 2),
    total_price  decimal(12, 2),
    txn_type     varchar(4),
    entry_source varchar(10)
);

