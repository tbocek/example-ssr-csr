<script>
const API_URL = '/api';

let games = $state([]);
let newGame = $state({ title: '', description: '' });
let loading = $state(true);
let error = $state(null);

// Fetch games on mount
async function fetchGames() {
    try {
        const response = await fetch(`${API_URL}/games`);
        if (!response.ok) throw new Error('Failed to fetch games');
        games = await response.json();
        loading = false;
    } catch (err) {
        error = err.message;
        loading = false;
    }
}

// Add game
async function addGame(event) {
    event.preventDefault();
    try {
        const response = await fetch(`${API_URL}/games`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(newGame)
        });
        
        if (!response.ok) throw new Error('Failed to add game');
        
        // Reset form
        newGame.title = '';
        newGame.description = '';
        
        // Refresh list
        await fetchGames();
    } catch (err) {
        alert('Error adding game: ' + err.message);
    }
}

// Add star
async function addStar(gameId) {
    try {
        const response = await fetch(`${API_URL}/games/${gameId}/star`, {
            method: 'POST'
        });
        
        if (!response.ok) throw new Error('Failed to add star');
        
        // Update locally
        const game = games.find(g => g.id === gameId);
        if (game) {
            game.stars++;
        }
    } catch (err) {
        alert('Error adding star: ' + err.message);
    }
}

fetchGames();

</script>

<style>
    :global(body) {
        font-family: Arial, sans-serif;
        max-width: 800px;
        margin: 50px auto;
        padding: 20px;
    }
    h1 { color: #333; }
    table {
        width: 100%;
        border-collapse: collapse;
        margin: 20px 0;
    }
    th, td {
        padding: 12px;
        text-align: left;
        border-bottom: 1px solid #ddd;
    }
    th {
        background-color: #4CAF50;
        color: white;
    }
    form {
        background: #f4f4f4;
        padding: 20px;
        border-radius: 5px;
        margin: 20px 0;
    }
    input[type="text"], textarea {
        width: calc(100% - 20px);
        padding: 10px;
        margin: 5px 0;
        border: 1px solid #ddd;
        border-radius: 3px;
    }
    textarea {
        resize: vertical;
        min-height: 60px;
    }
    input[type="submit"] {
        background-color: #4CAF50;
        color: white;
        padding: 10px 20px;
        border: none;
        border-radius: 3px;
        cursor: pointer;
        margin: 5px 0;
    }
    input[type="submit"]:hover {
        background-color: #45a049;
    }
    .star-btn {
        background: none;
        border: none;
        cursor: pointer;
        padding: 5px;
    }
    .star-btn:hover {
        transform: scale(1.2);
    }
    .star-count {
        display: inline-flex;
        align-items: center;
        gap: 5px;
    }
    .loading {
        text-align: center;
        color: #666;
    }
</style>

<svelte:head>
    <title>Game Management</title>
    <link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 10 10'><circle cx='5' cy='5' r='4' fill='%23007bff'/></svg>">
</svelte:head>

<h1>Game Management - PrevelteKit Demo</h1>

<h2>Add New Game</h2>
<form onsubmit={addGame}>
    <input type="text" bind:value={newGame.title} placeholder="Game Title" required />
    <textarea bind:value={newGame.description} placeholder="Description" required></textarea>
    <input type="submit" value="Add Game" />
</form>

<h2>Game List</h2>
{#if loading}
    <div class="loading">Loading games...</div>
{:else if error}
    <div style="color: red;">Error: {error}</div>
{:else if games.length > 0}
    <table>
        <thead>
            <tr>
                <th>ID</th>
                <th>Title</th>
                <th>Description</th>
                <th>Stars</th>
            </tr>
        </thead>
        <tbody>
            {#each games as game (game.id)}
                <tr>
                    <td>{game.id}</td>
                    <td>{game.title}</td>
                    <td>{game.description}</td>
                    <td>
                        <div class="star-count">
                            <button class="star-btn" onclick={() => addStar(game.id)} aria-label="star button">
                                <svg width="24" height="24" viewBox="0 0 24 24" fill="#FFD700" stroke="#FFA500" stroke-width="1">
                                    <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
                                </svg>
                            </button>
                            <span>{game.stars}</span>
                        </div>
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>
{:else}
    <p>No games found. Add one above!</p>
{/if}