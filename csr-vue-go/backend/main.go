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

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
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
	mux.HandleFunc("GET /api/products", handleProducts)
	mux.HandleFunc("POST /api/products", handleProducts)
	mux.HandleFunc("POST /api/products/{id}/buy", handleProductActions)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func initDB() error {
	query := `
	CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		price NUMERIC(10, 2) NOT NULL DEFAULT 0,
		stock INTEGER NOT NULL DEFAULT 0
	);`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	return err
}

func handleProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, name, description, price, stock FROM products ORDER BY id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		products := []Product{}
		for rows.Next() {
			var product Product
			if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			products = append(products, product)
		}

		json.NewEncoder(w).Encode(products)

	case "POST":
		var newProduct Product
		if err := json.NewDecoder(r.Body).Decode(&newProduct); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := db.QueryRow(
			"INSERT INTO products (name, description, price, stock) VALUES ($1, $2, $3, $4) RETURNING id",
			newProduct.Name, newProduct.Description, newProduct.Price, newProduct.Stock,
		).Scan(&newProduct.ID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newProduct)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleProductActions(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/products/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "buy" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	productID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		var product Product
		err := db.QueryRow(
			"UPDATE products SET stock = stock - 1 WHERE id = $1 AND stock > 0 RETURNING id, name, description, price, stock",
			productID,
		).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock)

		if err == sql.ErrNoRows {
			http.Error(w, "Product not found or sold out", http.StatusConflict)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
