const express = require('express');
const { Pool } = require('pg');

const app = express();
const port = 8080;

// Database connection
const pool = new Pool({
  connectionString: process.env.DATABASE_URL || 'postgres://postgres:password@localhost:5432/postgres?sslmode=disable'
});

// Middleware
app.use(express.urlencoded({ extended: true }));
app.use(express.json());

// Wait for database and initialize schema
async function initDB() {
  let retries = 120;
  while (retries > 0) {
    try {
      await pool.query('SELECT 1');
      console.log('Connected to database');
      break;
    } catch (err) {
      retries--;
      await new Promise(resolve => setTimeout(resolve, 250));
    }
  }

  if (retries === 0) {
    throw new Error('Failed to connect to database');
  }

  await pool.query(`
    CREATE TABLE IF NOT EXISTS games (
      id SERIAL PRIMARY KEY,
      title VARCHAR(255) NOT NULL,
      description TEXT NOT NULL,
      stars INTEGER DEFAULT 0
    )
  `);
}

// Helper function to escape HTML
function escapeHtml(text) {
  const map = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#039;'
  };
  return text.replace(/[&<>"']/g, m => map[m]);
}

// Helper function to render game row
function renderGameRow(game) {
  return `
    <tr id="game-${game.id}">
      <td>${game.id}</td>
      <td>${escapeHtml(game.title)}</td>
      <td>${escapeHtml(game.description)}</td>
      <td>
        <div class="star-count">
          <button class="star-btn" 
                  hx-post="/api/games/${game.id}/star" 
                  hx-target="#game-${game.id}"
                  hx-swap="outerHTML">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="#FFD700" stroke="#FFA500" stroke-width="1">
              <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
            </svg>
          </button>
          <span>${game.stars}</span>
        </div>
      </td>
    </tr>`;
}

// GET all games
app.get('/api/games', async (req, res) => {
  try {
    const result = await pool.query('SELECT id, title, description, stars FROM games ORDER BY id');
    
    if (result.rows.length === 0) {
      return res.send('<p>No games found. Add one above!</p>');
    }

    let html = `<table>
      <thead>
        <tr>
          <th>ID</th>
          <th>Title</th>
          <th>Description</th>
          <th>Stars</th>
        </tr>
      </thead>
      <tbody>`;

    result.rows.forEach(game => {
      html += renderGameRow(game);
    });

    html += '</tbody></table>';
    res.send(html);
  } catch (err) {
    console.error(err);
    res.status(500).send('Database error');
  }
});

// POST new game
app.post('/api/games', async (req, res) => {
  try {
    const { title, description } = req.body;

    if (!title || !description) {
      return res.status(400).send('Title and description required');
    }

    await pool.query(
      'INSERT INTO games (title, description, stars) VALUES ($1, $2, 0)',
      [title, description]
    );

    // Return updated game list
    const result = await pool.query('SELECT id, title, description, stars FROM games ORDER BY id');
    
    let html = `<table>
      <thead>
        <tr>
          <th>ID</th>
          <th>Title</th>
          <th>Description</th>
          <th>Stars</th>
        </tr>
      </thead>
      <tbody>`;

    result.rows.forEach(game => {
      html += renderGameRow(game);
    });

    html += '</tbody></table>';
    res.send(html);
  } catch (err) {
    console.error(err);
    res.status(500).send('Database error');
  }
});

// POST star a game
app.post('/api/games/:id/star', async (req, res) => {
  try {
    const gameId = parseInt(req.params.id);

    if (isNaN(gameId)) {
      return res.status(400).send('Invalid game ID');
    }

    const result = await pool.query(
      'UPDATE games SET stars = stars + 1 WHERE id = $1 RETURNING id, title, description, stars',
      [gameId]
    );

    if (result.rows.length === 0) {
      return res.status(404).send('Game not found');
    }

    res.send(renderGameRow(result.rows[0]));
  } catch (err) {
    console.error(err);
    res.status(500).send('Database error');
  }
});

// Start server
initDB().then(() => {
  app.listen(port, () => {
    console.log(`Server running on port ${port}`);
  });
}).catch(err => {
  console.error('Failed to initialize database:', err);
  process.exit(1);
});