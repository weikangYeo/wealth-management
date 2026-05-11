CREATE TABLE fund_info
(
    fund_code varchar(32) primary key not null,
    name      varchar(255),
    url       varchar(255)
);

CREATE TABLE fund_price_history
(
    fund_code       varchar(32),
    price_date      datetime,
    last_done_price DECIMAL(14, 4),
    PRIMARY KEY (fund_code, price_date),
    FOREIGN KEY (fund_code) REFERENCES fund_info (fund_code)
);

CREATE TABLE IF NOT EXISTS fund_txn
(
    id                VARCHAR(36) PRIMARY KEY NOT NULL,
    fund_code         VARCHAR(32)             NOT NULL,
    txn_date          DATE                    NOT NULL,
    unit              DECIMAL(14, 4)          NOT NULL,
    unit_price        DECIMAL(14, 4)          NOT NULL,
    sales_charge      DECIMAL(14, 4)          NOT NULL,
    gross_total_price DECIMAL(14, 4)          NOT NULL,
    net_total_price   DECIMAL(14, 4)          NOT NULL,
    txn_type          VARCHAR(16)             NOT NULL,
    remark            VARCHAR(100),
    FOREIGN KEY (fund_code) REFERENCES fund_info (fund_code)
);