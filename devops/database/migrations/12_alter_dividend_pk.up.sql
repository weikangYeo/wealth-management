ALTER TABLE stock_dividend
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (stock_name, ex_date, dividend_per_unit);