CREATE TABLE IF NOT EXISTS stock_price
(
    stock_name      VARCHAR(32),
    price_date      datetime,
    last_done_price DECIMAL(10, 4),
    PRIMARY KEY (stock_name, price_date),
    FOREIGN KEY (stock_name) REFERENCES stock (stock_name)
)