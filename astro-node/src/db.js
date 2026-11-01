import pkg from 'pg';
const { Pool } = pkg;

const pool = new Pool({
  connectionString: process.env.DATABASE_URL || 'postgres://postgres:password@localhost:5432/postgres'
});

let initialized;

// Runs once, before the first query from either the page or the API routes
function initDB() {
  initialized ??= pool.query(`
    CREATE TABLE IF NOT EXISTS products (
      id SERIAL PRIMARY KEY,
      name VARCHAR(255) NOT NULL,
      description TEXT NOT NULL,
      price NUMERIC(10, 2) NOT NULL DEFAULT 0,
      stock INTEGER NOT NULL DEFAULT 0
    )
  `).catch(err => {
    initialized = undefined;
    throw err;
  });
  return initialized;
}

export async function getProducts() {
  await initDB();
  const result = await pool.query('SELECT * FROM products ORDER BY id');
  return result.rows;
}

export async function createProduct(name, description, price, stock) {
  await initDB();
  await pool.query(
    'INSERT INTO products (name, description, price, stock) VALUES ($1, $2, $3, $4)',
    [name, description, price, stock]
  );
}

export async function buyProduct(id) {
  await initDB();
  const result = await pool.query(
    'UPDATE products SET stock = stock - 1 WHERE id = $1 AND stock > 0 RETURNING *',
    [id]
  );
  return result.rows[0];
}