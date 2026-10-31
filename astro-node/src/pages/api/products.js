import { getProducts, createProduct } from '../../db.js';

export async function GET() {
  const products = await getProducts();
  return new Response(JSON.stringify(products), {
    headers: { 'Content-Type': 'application/json' }
  });
}

export async function POST({ request }) {
  const formData = await request.formData();
  const name = formData.get('name');
  const description = formData.get('description');
  const price = formData.get('price');
  const stock = formData.get('stock');
  
  if (!name || !description || !price || !stock) {
    return new Response('Missing fields', { status: 400 });
  }
  
  await createProduct(name, description, price, stock);
  return new Response(null, { status: 201 });
}