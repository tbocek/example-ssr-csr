CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL DEFAULT 0,
    stock INTEGER NOT NULL DEFAULT 0
);

-- Install PGMQ extension
CREATE EXTENSION IF NOT EXISTS pgmq;

-- Create email queue (if not exists)
SELECT pgmq.create('email_queue');
