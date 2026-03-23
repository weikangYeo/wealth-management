create table if not exists gold_price_history
(
    date      date primary key not null,
    buy_price decimal(10, 2)
);