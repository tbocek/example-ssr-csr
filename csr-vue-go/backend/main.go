package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Game struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Stars       int    `json:"stars"`
}

var db *sql.DB

func main() {
	var err error

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"
	}

	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Wait for database to be ready
	for range 120 {
		if err := db.Ping(); err == nil {
			log.Println("Connected to database")
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to connect to database after retries:", err)
	}

	// Initialize database schema
	if err := initDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/games", handleGames)
	mux.HandleFunc("POST /api/games", handleGames)
	mux.HandleFunc("POST /api/games/{id}/star", handleGameActions)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func initDB() error {
	query := `
	CREATE TABLE IF NOT EXISTS games (
		id SERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		stars INTEGER DEFAULT 0
	);`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	return err
}

func handleGames(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, title, description, stars FROM games ORDER BY id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		games := []Game{}
		for rows.Next() {
			var game Game
			if err := rows.Scan(&game.ID, &game.Title, &game.Description, &game.Stars); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			games = append(games, game)
		}

		json.NewEncoder(w).Encode(games)

	case "POST":
		var newGame Game
		if err := json.NewDecoder(r.Body).Decode(&newGame); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := db.QueryRow(
			"INSERT INTO games (title, description, stars) VALUES ($1, $2, 0) RETURNING id",
			newGame.Title, newGame.Description,
		).Scan(&newGame.ID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		newGame.Stars = 0
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newGame)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGameActions(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/games/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "star" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	gameID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		var game Game
		err := db.QueryRow(
			"UPDATE games SET stars = stars + 1 WHERE id = $1 RETURNING id, title, description, stars",
			gameID,
		).Scan(&game.ID, &game.Title, &game.Description, &game.Stars)

		if err == sql.ErrNoRows {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(game)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
