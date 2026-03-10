CREATE TABLE IF NOT EXISTS games (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    stars INTEGER DEFAULT 0
);

-- Install PGMQ extension
CREATE EXTENSION IF NOT EXISTS pgmq;

-- Create email queue (if not exists)
SELECT pgmq.create('email_queue');
