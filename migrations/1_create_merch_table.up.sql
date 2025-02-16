BEGIN;

CREATE TABLE IF NOT EXISTS merch (
    id SERIAL PRIMARY KEY,
    item_name TEXT UNIQUE NOT NULL,
    price INT NOT NULL CHECK (price > 0)
);

INSERT INTO merch (item_name, price) VALUES
    ('t-shirt', 80),
    ('cup', 20),
    ('book', 50),
    ('pen', 10),
    ('powerbank', 200),
    ('hoody', 300),
    ('umbrella', 200),
    ('socks', 10),
    ('wallet', 50),
    ('pink-hoody', 500)
ON CONFLICT (item_name) DO NOTHING;

COMMIT;
