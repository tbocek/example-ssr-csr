import { buyProduct } from '../../../../db.js';

export async function POST({ params }) {
  const id = parseInt(params.id);
  
  if (isNaN(id)) {
    return new Response('Invalid ID', { status: 400 });
  }
  
  const product = await buyProduct(id);
  
  if (!product) {
    return new Response('Not found or sold out', { status: 409 });
  }
  
  return new Response(JSON.stringify(product), {
    headers: { 'Content-Type': 'application/json' }
  });
}