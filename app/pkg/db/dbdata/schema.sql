CREATE TABLE IF NOT EXISTS users(
    id SERIAL PRIMARY KEY,
    name VARCHAR(32),
    password VARCHAR(32),
    balance DECIMAL(6, 0),
    items JSONB,
    CONSTRAINT idx_users_name UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS items(
    id SERIAL PRIMARY KEY,
    name VARCHAR(32),
    price DECIMAL(5, 0),
    CONSTRAINT idx_items_name UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS operations(
    id SERIAL PRIMARY KEY,
    sender VARCHAR(32),
    reciver VARCHAR(32),
    amount DECIMAL(5, 0)
);

CREATE INDEX IF NOT EXISTS idx_sender ON operations USING HASH(sender);
CREATE INDEX IF NOT EXISTS idx_reciver ON operations USING HASH(reciver);

