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
    CREATE TABLE IF NOT EXISTS products (
      id SERIAL PRIMARY KEY,
      name VARCHAR(255) NOT NULL,
      description TEXT NOT NULL,
      price NUMERIC(10, 2) NOT NULL DEFAULT 0,
      stock INTEGER NOT NULL DEFAULT 0
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

// Helper function to render product row
function renderProductRow(product) {
  return `
    <tr id="product-${product.id}">
      <td>${product.id}</td>
      <td>${escapeHtml(product.name)}</td>
      <td>${escapeHtml(product.description)}</td>
      <td>CHF ${Number(product.price).toFixed(2)}</td>
      <td>${product.stock}</td>
      <td>
        <button ${product.stock === 0 ? 'disabled' : ''}
                hx-post="/api/products/${product.id}/buy"
                hx-target="#product-${product.id}"
                hx-swap="outerHTML"
                hx-status:409="swap:outerHTML">${product.stock === 0 ? 'Sold out' : 'Buy'}</button>
      </td>
    </tr>`;
}

// GET all products
app.get('/api/products', async (req, res) => {
  try {
    const result = await pool.query('SELECT id, name, description, price, stock FROM products ORDER BY id');
    
    if (result.rows.length === 0) {
      return res.send('<p>No products found. Add one above!</p>');
    }

    let html = `<table>
      <thead>
        <tr>
          <th>ID</th>
          <th>Name</th>
          <th>Description</th>
          <th>Price</th>
          <th>Stock</th>
          <th></th>
        </tr>
      </thead>
      <tbody>`;

    result.rows.forEach(product => {
      html += renderProductRow(product);
    });

    html += '</tbody></table>';
    res.send(html);
  } catch (err) {
    console.error(err);
    res.status(500).send('Database error');
  }
});

// POST new product
app.post('/api/products', async (req, res) => {
  try {
    const { name, description, price, stock } = req.body;

    if (!name || !description || !price || !stock) {
      return res.status(400).send('Name, description, price and stock required');
    }

    await pool.query(
      'INSERT INTO products (name, description, price, stock) VALUES ($1, $2, $3, $4)',
      [name, description, price, stock]
    );

    // Return updated product list
    const result = await pool.query('SELECT id, name, description, price, stock FROM products ORDER BY id');
    
    let html = `<table>
      <thead>
        <tr>
          <th>ID</th>
          <th>Name</th>
          <th>Description</th>
          <th>Price</th>
          <th>Stock</th>
          <th></th>
        </tr>
      </thead>
      <tbody>`;

    result.rows.forEach(product => {
      html += renderProductRow(product);
    });

    html += '</tbody></table>';
    res.send(html);
  } catch (err) {
    console.error(err);
    res.status(500).send('Database error');
  }
});

// POST buy a product
app.post('/api/products/:id/buy', async (req, res) => {
  try {
    const productId = parseInt(req.params.id);

    if (isNaN(productId)) {
      return res.status(400).send('Invalid product ID');
    }

    const result = await pool.query(
      'UPDATE products SET stock = stock - 1 WHERE id = $1 AND stock > 0 RETURNING id, name, description, price, stock',
      [productId]
    );

    if (result.rows.length === 0) {
      // Sold out in the meantime: send the current row so the page isn't left showing stale stock
      const current = await pool.query('SELECT id, name, description, price, stock FROM products WHERE id = $1', [productId]);
      if (current.rows.length === 0) {
        return res.status(404).send('Product not found');
      }
      return res.status(409).send(renderProductRow(current.rows[0]));
    }

    res.send(renderProductRow(result.rows[0]));
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