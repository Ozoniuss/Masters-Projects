CREATE TABLE IF NOT EXISTS orders
(
    id SERIAL PRIMARY KEY,
    content text NOT NULL,
    status text NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
