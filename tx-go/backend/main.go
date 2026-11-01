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

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

type ProductStatistics struct {
	ID          int       `json:"id"`
	ProductID   int       `json:"product_id"`
	TotalSold   int       `json:"total_sold"`
	LastUpdated time.Time `json:"last_updated"`
}

var db *sql.DB

func main() {
	var err error

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@db:5432/shopdb?sslmode=disable"
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

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	mux := http.NewServeMux()

	// Original endpoints
	mux.HandleFunc("GET /api/products", handleProducts)
	mux.HandleFunc("POST /api/products", handleProducts)
	mux.HandleFunc("POST /api/products/{id}/buy", handleProductActions)

	// Transaction demo endpoints
	mux.HandleFunc("POST /api/demo/with-transaction/{id}", handleWithTransaction)
	mux.HandleFunc("POST /api/demo/without-transaction/{id}", handleWithoutTransaction)
	mux.HandleFunc("POST /api/demo/transfer", handleTransferStock)
	mux.HandleFunc("POST /api/demo/transfer-no-tx", handleTransferStockNoTx)
	mux.HandleFunc("GET /api/demo/product/{id}", handleGetProductDetails)

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

		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(product)
	}
}

// ============= TRANSACTION DEMO ENDPOINTS =============

// WITH Transaction - demonstrates rollback on failure
func handleWithTransaction(w http.ResponseWriter, r *http.Request) {
	productIDStr := strings.TrimPrefix(r.URL.Path, "/api/demo/with-transaction/")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Get stock before
	var stockBefore int
	db.QueryRow("SELECT stock FROM products WHERE id = $1", productID).Scan(&stockBefore)

	// Call transaction function
	err = moveStockWithTransaction(productID)

	var stockAfter int
	db.QueryRow("SELECT stock FROM products WHERE id = $1", productID).Scan(&stockAfter)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":        err.Error(),
			"stock_before": stockBefore,
			"stock_after":  stockAfter,
			"rolled_back":  stockBefore == stockAfter,
			"message":      "Transaction rolled back! Stock unchanged.",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      true,
			"stock_before": stockBefore,
			"stock_after":  stockAfter,
			"message":      "Transaction committed successfully!",
		})
	}
}

// WITHOUT Transaction - demonstrates data corruption
func handleWithoutTransaction(w http.ResponseWriter, r *http.Request) {
	productIDStr := strings.TrimPrefix(r.URL.Path, "/api/demo/without-transaction/")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Get stock before (this demo always changes product 1, whatever the ID)
	var stockBefore int
	db.QueryRow("SELECT stock FROM products WHERE id = $1", 1).Scan(&stockBefore)

	// Call NON-transaction function
	err = moveStockWithoutTransaction(productID)

	var stockAfter int
	db.QueryRow("SELECT stock FROM products WHERE id = $1", 1).Scan(&stockAfter)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":        err.Error(),
			"stock_before": stockBefore,
			"stock_after":  stockAfter,
			"rolled_back":  false,
			"message":      "NO ROLLBACK! Unit was added to product 1 even though error occurred. No way to undo!",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":      true,
			"stock_before": stockBefore,
			"stock_after":  stockAfter,
			"message":      "Unit added to product 1 successfully (simple operation, no transaction needed)",
		})
	}
}

// Transfer stock between products with transaction
func handleTransferStock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromID int `json:"from_id"`
		ToID   int `json:"to_id"`
		Stock  int `json:"stock"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Stock < 1 {
		http.Error(w, "stock must be positive", http.StatusBadRequest)
		return
	}

	// Get stock before
	var fromStockBefore, toStockBefore int
	db.QueryRow("SELECT stock FROM products WHERE id = $1", req.FromID).Scan(&fromStockBefore)
	db.QueryRow("SELECT stock FROM products WHERE id = $1", req.ToID).Scan(&toStockBefore)

	// Execute transfer
	err := transferStockWithTransaction(req.FromID, req.ToID, req.Stock)

	// Get stock after
	var fromStockAfter, toStockAfter int
	db.QueryRow("SELECT stock FROM products WHERE id = $1", req.FromID).Scan(&fromStockAfter)
	db.QueryRow("SELECT stock FROM products WHERE id = $1", req.ToID).Scan(&toStockAfter)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
			"from_product": map[string]int{
				"before": fromStockBefore,
				"after":  fromStockAfter,
			},
			"to_product": map[string]int{
				"before": toStockBefore,
				"after":  toStockAfter,
			},
			"rolled_back": fromStockBefore == fromStockAfter && toStockBefore == toStockAfter,
			"message":     "Transaction rolled back! Both products unchanged.",
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"from_product": map[string]int{
				"before": fromStockBefore,
				"after":  fromStockAfter,
			},
			"to_product": map[string]int{
				"before": toStockBefore,
				"after":  toStockAfter,
			},
			"message": "Transfer successful!",
		})
	}
}

// Get product details
func handleGetProductDetails(w http.ResponseWriter, r *http.Request) {
	productIDStr := strings.TrimPrefix(r.URL.Path, "/api/demo/product/")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var product Product
	err = db.QueryRow(
		"SELECT id, name, description, price, stock FROM products WHERE id = $1",
		productID,
	).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock)

	if err == sql.ErrNoRows {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// Transfer stock WITHOUT transaction - demonstrates data corruption
func handleTransferStockNoTx(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FromID int `json:"from_id"`
		ToID   int `json:"to_id"`
		Stock  int `json:"stock"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Stock < 1 {
		http.Error(w, "stock must be positive", http.StatusBadRequest)
		return
	}

	// Get stock before
	var fromStockBefore, toStockBefore int
	db.QueryRow("SELECT stock FROM products WHERE id = $1", req.FromID).Scan(&fromStockBefore)
	db.QueryRow("SELECT stock FROM products WHERE id = $1", req.ToID).Scan(&toStockBefore)

	// Execute transfer WITHOUT transaction
	err := transferStockWithoutTransaction(req.FromID, req.ToID, req.Stock)

	// Get stock after
	var fromStockAfter, toStockAfter int
	db.QueryRow("SELECT stock FROM products WHERE id = $1", req.FromID).Scan(&fromStockAfter)
	db.QueryRow("SELECT stock FROM products WHERE id = $1", req.ToID).Scan(&toStockAfter)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
		"from_product": map[string]int{
			"before": fromStockBefore,
			"after":  fromStockAfter,
		},
		"to_product": map[string]int{
			"before": toStockBefore,
			"after":  toStockAfter,
		},
		"rolled_back": false,
		"message":     "NO ROLLBACK! First operation committed, second operation never executed. Data is inconsistent!",
	})
}

// ============= TRANSACTION LOGIC =============

// WITH Transaction - proper implementation
// Deducts from target product (TO), adds to product 1 (FROM)
func moveStockWithTransaction(productID int) error {
	const fromProductID = 1  // Product 1 receives the unit
	toProductID := productID // Target product loses the unit

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Defer rollback - will only execute if commit is not called
	defer tx.Rollback()

	// Operation 1: Deduct unit from TO product
	var toStock int
	err = tx.QueryRow("SELECT stock FROM products WHERE id = $1 FOR UPDATE", toProductID).Scan(&toStock)
	if err != nil {
		log.Println("Operation 1 failed (select to):", err)
		return err
	}

	if toStock < 1 {
		return fmt.Errorf("target product has no stock to transfer")
	}

	_, err = tx.Exec("UPDATE products SET stock = stock - 1 WHERE id = $1", toProductID)
	if err != nil {
		log.Println("Operation 1 failed (deduct from to):", err)
		return err
	}
	log.Printf("Operation 1: Deducted 1 unit from product %d\n", toProductID)

	// Simulate failure here
	// Uncomment to test rollback:
	// return fmt.Errorf("Simulated failure! Network error!")

	// Operation 2: Add unit to FROM product
	_, err = tx.Exec("UPDATE products SET stock = stock + 1 WHERE id = $1", fromProductID)
	if err != nil {
		log.Println("Operation 2 failed (add to from):", err)
		return err // Rollback happens automatically
	}
	log.Printf("Operation 2: Added 1 unit to product %d\n", fromProductID)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Println("Commit failed:", err)
		return err
	}

	log.Println("Transaction committed successfully")
	return nil
}

// WITHOUT Transaction - demonstrates the problem
// Just adds to product 1 (FROM), no deduction
func moveStockWithoutTransaction(productID int) error {
	const fromProductID = 1 // Product 1 receives the unit

	// Operation 1: Add unit to FROM product (NO TRANSACTION)
	_, err := db.Exec("UPDATE products SET stock = stock + 1 WHERE id = $1", fromProductID)
	if err != nil {
		log.Println("Operation 1 failed:", err)
		return err
	}
	log.Printf("Operation 1: Added 1 unit to product %d (SAVED TO DB)\n", fromProductID)

	return nil
}

// Transfer stock with transaction
func transferStockWithTransaction(fromID, toID, stock int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Lock both products in ID order, so that opposite transfers can't deadlock
	_, err = tx.Exec("SELECT id FROM products WHERE id IN ($1, $2) ORDER BY id FOR UPDATE", fromID, toID)
	if err != nil {
		return err
	}

	// Operation 1: Check and deduct stock from source product
	var fromStockBefore int
	err = tx.QueryRow("SELECT stock FROM products WHERE id = $1", fromID).Scan(&fromStockBefore)
	if err != nil {
		return err
	}

	if fromStockBefore < stock {
		return fmt.Errorf("insufficient stock: product has %d but trying to transfer %d", fromStockBefore, stock)
	}

	_, err = tx.Exec("UPDATE products SET stock = stock - $1 WHERE id = $2", stock, fromID)
	if err != nil {
		return err
	}

	log.Printf("Deducted %d units from product %d\n", stock, fromID)

	// Operation 2: Check and add stock to target product
	var toStockBefore int
	err = tx.QueryRow("SELECT stock FROM products WHERE id = $1", toID).Scan(&toStockBefore)
	if err != nil {
		return err
	}

	if toStockBefore+stock > 100 {
		log.Printf("Business rule violation: Target product has %d units, adding %d would exceed 100\n", toStockBefore, stock)
		return fmt.Errorf("target product would exceed 100 units (%d + %d = %d)", toStockBefore, stock, toStockBefore+stock)
	}

	_, err = tx.Exec("UPDATE products SET stock = stock + $1 WHERE id = $2", stock, toID)
	if err != nil {
		return err
	}

	log.Printf("Added %d units to product %d\n", stock, toID)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("Transfer transaction committed")
	return nil
}

// Transfer stock WITHOUT transaction - demonstrates data corruption
func transferStockWithoutTransaction(fromID, toID, stock int) error {
	// Operation 1: Deduct stock from source product (NO TRANSACTION)
	_, err := db.Exec("UPDATE products SET stock = stock - $1 WHERE id = $2", stock, fromID)
	if err != nil {
		log.Println("Operation 1 failed:", err)
		return err
	}
	log.Printf("Operation 1: Deducted %d units from product %d (SAVED TO DB)\n", stock, fromID)

	// Simulate failure BEFORE operation 2
	log.Println("Simulated failure! Network error before adding stock to target")
	return fmt.Errorf("network error before operation 2")

	// Operation 2: Add stock to target product (NEVER EXECUTES)
	// This is unreachable due to return above
	_, err = db.Exec("UPDATE products SET stock = stock + $1 WHERE id = $2", stock, toID)

	// Without transaction: Operation 1 already committed
	// Stock deducted from source but never added to target
	// Stock disappeared! Data corruption!

	return err
}
