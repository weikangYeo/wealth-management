CREATE TABLE IF NOT EXISTS gold_price_history
(
    date      DATE PRIMARY KEY NOT NULL,
    buy_price DECIMAL(10, 2)
);