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
	"github.com/streadway/amqp"
)

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

var db *sql.DB
var rabbitConn *amqp.Connection
var rabbitCh *amqp.Channel

func main() {
	var err error

	// Database setup
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@db:5432/shopdb?sslmode=disable"
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

	// RabbitMQ setup
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	for range 120 {
		var err error
		rabbitConn, err = amqp.Dial(rabbitMQURL)
		if err == nil {
			log.Println("Connected to RabbitMQ")
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	if rabbitConn == nil {
		log.Fatal("Failed to connect to RabbitMQ after retries")
	}

	rabbitCh, err = rabbitConn.Channel()
	if err != nil {
		log.Fatal("Failed to open RabbitMQ channel:", err)
	}

	// Declare queue
	_, err = rabbitCh.QueueDeclare(
		"email_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to declare queue:", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/products", handleProducts)
	mux.HandleFunc("POST /api/products", handleProducts)
	mux.HandleFunc("POST /api/products/{id}/buy", handleProductActions)

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

		// Publish event to RabbitMQ
		if rabbitCh != nil {
			body, err := json.Marshal(product)
			if err != nil {
				log.Printf("Failed to marshal product event: %v", err)
			} else {
				err = rabbitCh.Publish(
					//"",            // exchange
					//"email_queue", // routing key
					"product_events", // exchange (instead of "")
					"",               // routing key (empty for fanout)
					false,            // mandatory
					false,            // immediate,
					amqp.Publishing{
						ContentType:  "application/json",
						Body:         body,
						DeliveryMode: 2, //2 is persistent mode
					})
				if err != nil {
					log.Printf("Failed to publish message: %v", err)
				} else {
					log.Printf("Published event for product %d", product.ID)
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
	}
}
