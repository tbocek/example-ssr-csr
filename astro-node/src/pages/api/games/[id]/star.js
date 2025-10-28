import { incrementStar } from '../../../../db.js';

export async function POST({ params }) {
  const id = parseInt(params.id);
  
  if (isNaN(id)) {
    return new Response('Invalid ID', { status: 400 });
  }
  
  const game = await incrementStar(id);
  
  if (!game) {
    return new Response('Not found', { status: 404 });
  }
  
  return new Response(JSON.stringify(game), {
    headers: { 'Content-Type': 'application/json' }
  });
}