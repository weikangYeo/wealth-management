ALTER TABLE stock_dividend
    MODIFY COLUMN stock_unit        DECIMAL(10, 2) NOT NULL,
    MODIFY COLUMN dividend_per_unit DECIMAL(10, 2) NOT NULL,
    MODIFY COLUMN tax               DECIMAL(10, 2) NOT NULL,
    MODIFY COLUMN gross_amount      DECIMAL(10, 2) NOT NULL,
    MODIFY COLUMN net_amount        DECIMAL(10, 2) NOT NULL;
