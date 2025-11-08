const { createApp } = Vue;

const API_URL = '/api';

createApp({
    data() {
        return {
            games: [],
            newGame: {
                title: '',
                description: ''
            },
            loading: true,
            error: null
        }
    },
    mounted() {
        this.fetchGames();
    },
    methods: {
        async fetchGames() {
            try {
                const response = await fetch(`${API_URL}/games`);
                if (!response.ok) throw new Error('Failed to fetch games');
                this.games = await response.json();
                this.loading = false;
            } catch (err) {
                this.error = err.message;
                this.loading = false;
            }
        },
        async addGame() {
            try {
                const response = await fetch(`${API_URL}/games`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(this.newGame)
                });
                
                if (!response.ok) throw new Error('Failed to add game');
                
                // Reset form
                this.newGame.title = '';
                this.newGame.description = '';
                
                // Refresh list
                await this.fetchGames();
            } catch (err) {
                alert('Error adding game: ' + err.message);
            }
        },
        async addStar(gameId) {
            try {
                const response = await fetch(`${API_URL}/games/${gameId}/star`, {
                    method: 'POST'
                });
                
                if (!response.ok) throw new Error('Failed to add star');
                
                // Update the game's star count locally
                const game = this.games.find(g => g.id === gameId);
                if (game) {
                    game.stars++;
                }
            } catch (err) {
                alert('Error adding star: ' + err.message);
            }
        }
    }
}).mount('#app');