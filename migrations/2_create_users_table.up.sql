BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    token TEXT,
    balance INT NOT NULL DEFAULT 1000
);

CREATE TABLE IF NOT EXISTS purchases (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    item TEXT NOT NULL,
    price INT NOT NULL,
    purchase_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS inventory (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    item_name VARCHAR(100) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    UNIQUE (user_id, item_name),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    from_user TEXT,
    to_user TEXT NOT NULL,
    amount INT NOT NULL,
    FOREIGN KEY (from_user) REFERENCES users(username) ON DELETE SET NULL,
    FOREIGN KEY (to_user) REFERENCES users(username) ON DELETE CASCADE
);

COMMIT;
