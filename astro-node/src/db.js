import pkg from 'pg';
const { Pool } = pkg;

const pool = new Pool({
  connectionString: process.env.DATABASE_URL || 'postgres://postgres:password@localhost:5432/postgres'
});

export async function initDB() {
  await pool.query(`
    CREATE TABLE IF NOT EXISTS games (
      id SERIAL PRIMARY KEY,
      title VARCHAR(255) NOT NULL,
      description TEXT NOT NULL,
      stars INTEGER DEFAULT 0
    )
  `);
}

export async function getGames() {
  const result = await pool.query('SELECT * FROM games ORDER BY id');
  return result.rows;
}

export async function createGame(title, description) {
  await pool.query(
    'INSERT INTO games (title, description, stars) VALUES ($1, $2, 0)',
    [title, description]
  );
}

export async function incrementStar(id) {
  const result = await pool.query(
    'UPDATE games SET stars = stars + 1 WHERE id = $1 RETURNING *',
    [id]
  );
  return result.rows[0];
}