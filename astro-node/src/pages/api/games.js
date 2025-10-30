import { getGames, createGame } from '../../db.js';

export async function GET() {
  const games = await getGames();
  return new Response(JSON.stringify(games), {
    headers: { 'Content-Type': 'application/json' }
  });
}

export async function POST({ request }) {
  const formData = await request.formData();
  const title = formData.get('title');
  const description = formData.get('description');
  
  if (!title || !description) {
    return new Response('Missing fields', { status: 400 });
  }
  
  await createGame(title, description);
  return new Response(null, { status: 201 });
}