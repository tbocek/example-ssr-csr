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

	_ "github.com/lib/pq"
)

type Game struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Stars       int    `json:"stars"`
}

type GameStatistics struct {
	ID          int       `json:"id"`
	GameID      int       `json:"game_id"`
	TotalStars  int       `json:"total_stars"`
	LastUpdated time.Time `json:"last_updated"`
}

var db *sql.DB

func main() {
	var err error

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@db:5432/gamedb?sslmode=disable"
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

	// Load sample data
	if err := loadSampleData(); err != nil {
		log.Fatal("Failed to load sample data:", err)
	}

	mux := http.NewServeMux()
	
	// Original endpoints
	mux.HandleFunc("GET /api/games", handleGames)
	mux.HandleFunc("POST /api/games", handleGames)
	mux.HandleFunc("POST /api/games/{id}/star", handleGameActions)
	
	// Transaction demo endpoints
	mux.HandleFunc("POST /api/demo/with-transaction/{id}", handleWithTransaction)
	mux.HandleFunc("POST /api/demo/without-transaction/{id}", handleWithoutTransaction)
	mux.HandleFunc("POST /api/demo/transfer", handleTransferStars)
	mux.HandleFunc("POST /api/demo/transfer-no-tx", handleTransferStarsNoTx)
	mux.HandleFunc("GET /api/demo/game/{id}", handleGetGameDetails)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func initDB() error {
	// Create games table
	gamesTable := `
	CREATE TABLE IF NOT EXISTS games (
		id SERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		stars INTEGER DEFAULT 0
	);`

	// Create game_statistics table
	statsTable := `
	CREATE TABLE IF NOT EXISTS game_statistics (
		id SERIAL PRIMARY KEY,
		game_id INTEGER UNIQUE NOT NULL REFERENCES games(id) ON DELETE CASCADE,
		total_stars INTEGER DEFAULT 0,
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(gamesTable); err != nil {
		return err
	}

	if _, err := db.Exec(statsTable); err != nil {
		return err
	}

	log.Println("Database schema initialized")
	return nil
}

func loadSampleData() error {
	// Check if data already exists
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM games").Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		log.Println("Sample data already exists")
		return nil
	}

	// Insert sample games
	games := []struct {
		title       string
		description string
		stars       int
	}{
		{"The Legend of Zelda", "Epic adventure game", 5},
		{"Super Mario Bros", "Classic platformer", 3},
		{"Metroid", "Space exploration", 4},
	}

	for _, game := range games {
		_, err := db.Exec(
			"INSERT INTO games (title, description, stars) VALUES ($1, $2, $3)",
			game.title, game.description, game.stars,
		)
		if err != nil {
			return err
		}
	}

	log.Println("Sample games loaded")
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(game)
	}
}

// ============= TRANSACTION DEMO ENDPOINTS =============

// WITH Transaction - demonstrates rollback on failure
func handleWithTransaction(w http.ResponseWriter, r *http.Request) {
	gameIDStr := strings.TrimPrefix(r.URL.Path, "/api/demo/with-transaction/")
	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	// Get stars before
	var starsBefore int
	db.QueryRow("SELECT stars FROM games WHERE id = $1", gameID).Scan(&starsBefore)

	// Call transaction function
	err = addStarWithTransaction(gameID)

	var starsAfter int
	db.QueryRow("SELECT stars FROM games WHERE id = $1", gameID).Scan(&starsAfter)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":        err.Error(),
			"stars_before": starsBefore,
			"stars_after":  starsAfter,
			"rolled_back":  starsBefore == starsAfter,
			"message":      "Transaction rolled back! Stars unchanged.",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      true,
			"stars_before": starsBefore,
			"stars_after":  starsAfter,
			"message":      "Transaction committed successfully!",
		})
	}
}

// WITHOUT Transaction - demonstrates data corruption
func handleWithoutTransaction(w http.ResponseWriter, r *http.Request) {
	gameIDStr := strings.TrimPrefix(r.URL.Path, "/api/demo/without-transaction/")
	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	// Get stars before
	var starsBefore int
	db.QueryRow("SELECT stars FROM games WHERE id = $1", gameID).Scan(&starsBefore)

	// Call non-transaction function
	err = addStarWithoutTransaction(gameID)

	var starsAfter int
	db.QueryRow("SELECT stars FROM games WHERE id = $1", gameID).Scan(&starsAfter)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":        err.Error(),
			"stars_before": starsBefore,
			"stars_after":  starsAfter,
			"rolled_back":  false,
			"message":      "Error occurred but no transaction to rollback",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      true,
			"stars_before": starsBefore,
			"stars_after":  starsAfter,
			"message":      "Star added to game 1 successfully (simple operation, no transaction needed)",
		})
	}
}

// Transfer stars between games with transaction
func handleTransferStars(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromID int `json:"from_id"`
		ToID   int `json:"to_id"`
		Stars  int `json:"stars"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get stars before
	var fromStarsBefore, toStarsBefore int
	db.QueryRow("SELECT stars FROM games WHERE id = $1", req.FromID).Scan(&fromStarsBefore)
	db.QueryRow("SELECT stars FROM games WHERE id = $1", req.ToID).Scan(&toStarsBefore)

	// Execute transfer
	err := transferStarsWithTransaction(req.FromID, req.ToID, req.Stars)

	// Get stars after
	var fromStarsAfter, toStarsAfter int
	db.QueryRow("SELECT stars FROM games WHERE id = $1", req.FromID).Scan(&fromStarsAfter)
	db.QueryRow("SELECT stars FROM games WHERE id = $1", req.ToID).Scan(&toStarsAfter)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
			"from_game": map[string]int{
				"before": fromStarsBefore,
				"after":  fromStarsAfter,
			},
			"to_game": map[string]int{
				"before": toStarsBefore,
				"after":  toStarsAfter,
			},
			"rolled_back": fromStarsBefore == fromStarsAfter && toStarsBefore == toStarsAfter,
			"message":     "Transaction rolled back! Both games unchanged.",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"from_game": map[string]int{
				"before": fromStarsBefore,
				"after":  fromStarsAfter,
			},
			"to_game": map[string]int{
				"before": toStarsBefore,
				"after":  toStarsAfter,
			},
			"message": "Transfer successful!",
		})
	}
}

// Get game details
func handleGetGameDetails(w http.ResponseWriter, r *http.Request) {
	gameIDStr := strings.TrimPrefix(r.URL.Path, "/api/demo/game/")
	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	var game Game
	err = db.QueryRow(
		"SELECT id, title, description, stars FROM games WHERE id = $1",
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
}

// Transfer stars WITHOUT transaction - demonstrates data corruption
func handleTransferStarsNoTx(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromID int `json:"from_id"`
		ToID   int `json:"to_id"`
		Stars  int `json:"stars"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get stars before
	var fromStarsBefore, toStarsBefore int
	db.QueryRow("SELECT stars FROM games WHERE id = $1", req.FromID).Scan(&fromStarsBefore)
	db.QueryRow("SELECT stars FROM games WHERE id = $1", req.ToID).Scan(&toStarsBefore)

	// Execute transfer WITHOUT transaction
	err := transferStarsWithoutTransaction(req.FromID, req.ToID, req.Stars)

	// Get stars after
	var fromStarsAfter, toStarsAfter int
	db.QueryRow("SELECT stars FROM games WHERE id = $1", req.FromID).Scan(&fromStarsAfter)
	db.QueryRow("SELECT stars FROM games WHERE id = $1", req.ToID).Scan(&toStarsAfter)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
		"from_game": map[string]int{
			"before": fromStarsBefore,
			"after":  fromStarsAfter,
		},
		"to_game": map[string]int{
			"before": toStarsBefore,
			"after":  toStarsAfter,
		},
		"rolled_back": false,
		"message":     "NO ROLLBACK! First operation committed, second operation never executed. Data is inconsistent!",
	})
}

// ============= TRANSACTION LOGIC =============

// WITH Transaction - proper implementation
// Deducts from target game (TO), adds to game 1 (FROM)
func addStarWithTransaction(gameID int) error {
	const fromGameID = 1 // Game 1 receives the star
	toGameID := gameID   // Target game loses the star
	
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Defer rollback - will only execute if commit is not called
	defer tx.Rollback()

	// Operation 1: Deduct star from TO game
	var toStars int
	err = tx.QueryRow("SELECT stars FROM games WHERE id = $1", toGameID).Scan(&toStars)
	if err != nil {
		log.Println("Operation 1 failed (select to):", err)
		return err
	}
	
	if toStars < 1 {
		return fmt.Errorf("target game has no stars to transfer")
	}
	
	_, err = tx.Exec("UPDATE games SET stars = stars - 1 WHERE id = $1", toGameID)
	if err != nil {
		log.Println("Operation 1 failed (deduct from to):", err)
		return err
	}
	log.Printf("Operation 1: Deducted 1 star from game %d\n", toGameID)

	// Simulate failure here
	// Uncomment to test rollback:
	// return fmt.Errorf("Simulated failure! Network error!")

	// Operation 2: Add star to FROM game
	_, err = tx.Exec("UPDATE games SET stars = stars + 1 WHERE id = $1", fromGameID)
	if err != nil {
		log.Println("Operation 2 failed (add to from):", err)
		return err // Rollback happens automatically
	}
	log.Printf("Operation 2: Added 1 star to game %d\n", fromGameID)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Println("Commit failed:", err)
		return err
	}

	log.Println("Transaction committed successfully")
	return nil
}

// WITHOUT Transaction - demonstrates the problem
// Just adds to game 1 (FROM), no deduction
func addStarWithoutTransaction(gameID int) error {
	const fromGameID = 1 // Game 1 receives the star
	
	// Operation 1: Add star to FROM game (NO TRANSACTION)
	_, err := db.Exec("UPDATE games SET stars = stars + 1 WHERE id = $1", fromGameID)
	if err != nil {
		log.Println("Operation 1 failed:", err)
		return err
	}
	log.Printf("Operation 1: Added 1 star to game %d (SAVED TO DB)\n", fromGameID)

	return nil
}

// Transfer stars with transaction
func transferStarsWithTransaction(fromID, toID, stars int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Operation 1: Check and deduct stars from source game
	var fromStarsBefore int
	err = tx.QueryRow("SELECT stars FROM games WHERE id = $1", fromID).Scan(&fromStarsBefore)
	if err != nil {
		return err
	}

	if fromStarsBefore < stars {
		return fmt.Errorf("insufficient stars: game has %d but trying to transfer %d", fromStarsBefore, stars)
	}

	_, err = tx.Exec("UPDATE games SET stars = stars - $1 WHERE id = $2", stars, fromID)
	if err != nil {
		return err
	}

	log.Printf("Deducted %d stars from game %d\n", stars, fromID)

	// Operation 2: Check and add stars to target game
	var toStarsBefore int
	err = tx.QueryRow("SELECT stars FROM games WHERE id = $1", toID).Scan(&toStarsBefore)
	if err != nil {
		return err
	}

	if toStarsBefore+stars > 100 {
		log.Printf("Business rule violation: Target game has %d stars, adding %d would exceed 100\n", toStarsBefore, stars)
		return fmt.Errorf("target game would exceed 100 stars (%d + %d = %d)", toStarsBefore, stars, toStarsBefore+stars)
	}

	_, err = tx.Exec("UPDATE games SET stars = stars + $1 WHERE id = $2", stars, toID)
	if err != nil {
		return err
	}

	log.Printf("Added %d stars to game %d\n", stars, toID)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("Transfer transaction committed")
	return nil
}

// Transfer stars WITHOUT transaction - demonstrates data corruption
func transferStarsWithoutTransaction(fromID, toID, stars int) error {
	// Operation 1: Deduct stars from source game (NO TRANSACTION)
	_, err := db.Exec("UPDATE games SET stars = stars - $1 WHERE id = $2", stars, fromID)
	if err != nil {
		log.Println("Operation 1 failed:", err)
		return err
	}
	log.Printf("Operation 1: Deducted %d stars from game %d (SAVED TO DB)\n", stars, fromID)

	// Simulate failure BEFORE operation 2
	log.Println("Simulated failure! Network error before adding stars to target")
	return fmt.Errorf("network error before operation 2")

	// Operation 2: Add stars to target game (NEVER EXECUTES)
	// This is unreachable due to return above
	_, err = db.Exec("UPDATE games SET stars = stars + $1 WHERE id = $2", stars, toID)
	
	// Without transaction: Operation 1 already committed
	// Stars deducted from source but never added to target
	// Stars disappeared! Data corruption!
	
	return err
}