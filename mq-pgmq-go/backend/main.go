package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	// Database setup
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@db:5432/gamedb?sslmode=disable"
	}

	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Wait for database
	for range 120 {
		if err := db.Ping(); err == nil {
			log.Println("Connected to database")
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/games", handleGames)
	mux.HandleFunc("POST /api/games", handleGames)
	mux.HandleFunc("POST /api/games/{id}/star", handleGameActions)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
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

		body, err := json.Marshal(game)
		if err != nil {
			log.Printf("Failed to marshal game event: %v", err)
		} else {
			if err := publishToPGMQ(db, body); err != nil {
				log.Printf("Failed to publish to PGMQ: %v", err)
			} else {
				log.Printf("Published event for game %d", game.ID)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(game)
	}
}

func publishToPGMQ(db *sql.DB, gameData []byte) error {
	// Send message to queue using SQL
	var msgID int64
	err := db.QueryRow("SELECT pgmq.send($1, $2)", "email_queue", string(gameData)).Scan(&msgID)
	if err != nil {
		return fmt.Errorf("failed to send message to PGMQ: %v", err)
	}

	log.Printf("Published event to PGMQ queue (message ID: %d)", msgID)
	return nil
}
