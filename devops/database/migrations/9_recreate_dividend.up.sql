DROP TABLE IF EXISTS stock_dividend;

CREATE TABLE IF NOT EXISTS stock_dividend
(
    stock_name        VARCHAR(32)    NOT NULL,
    ex_date           DATE           NOT NULL,
    payment_date      DATE           NOT NULL,
    stock_unit        DECIMAL(10, 2) NOT NULL,
    dividend_per_unit DECIMAL(10, 2) NOT NULL,
    tax               DECIMAL(10, 2) NOT NULL,
    gross_amount      DECIMAL(10, 2) NOT NULL,
    net_amount        DECIMAL(10, 2) NOT NULL,
    remark            VARCHAR(255)   NOT NULL,
    PRIMARY KEY (stock_name, ex_date),
    FOREIGN KEY (stock_name) REFERENCES stock (stock_name)
);