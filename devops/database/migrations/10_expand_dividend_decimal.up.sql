ALTER TABLE stock_dividend
    MODIFY COLUMN stock_unit        DECIMAL(14, 4) NOT NULL,
    MODIFY COLUMN dividend_per_unit DECIMAL(14, 4) NOT NULL,
    MODIFY COLUMN tax               DECIMAL(14, 4) NOT NULL,
    MODIFY COLUMN gross_amount      DECIMAL(14, 4) NOT NULL,
    MODIFY COLUMN net_amount        DECIMAL(14, 4) NOT NULL;
