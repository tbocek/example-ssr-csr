package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	testcontainerspostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDB            *sql.DB
	postgresContainer *testcontainerspostgres.PostgresContainer
	testContainerCtx  context.Context
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(m.Run())
	}

	var err error
	testContainerCtx = context.Background()

	// Start PostgreSQL container
	postgresContainer, err = testcontainerspostgres.Run(
		testContainerCtx,
		"postgres:18-alpine",
		testcontainerspostgres.WithDatabase("testdb"),
		testcontainerspostgres.WithUsername("testuser"),
		testcontainerspostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		fmt.Printf("Failed to start postgres container: %v\n", err)
		panic(err)
	}

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(testContainerCtx, "sslmode=disable")
	if err != nil {
		fmt.Printf("Failed to get connection string: %v\n", err)
		panic(err)
	}

	// Connect to database
	testDB, err = sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		panic(err)
	}

	// Run migrations
	if err := setupTestDB(testDB); err != nil {
		fmt.Printf("Failed to setup test database: %v\n", err)
		panic(err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	testDB.Close()
	if err := postgresContainer.Terminate(testContainerCtx); err != nil {
		fmt.Printf("Failed to terminate container: %v\n", err)
	}

	os.Exit(code)
}

func setupTestDB(database *sql.DB) error {
	driver, err := postgres.WithInstance(database, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return err
	}

	// Force to clean state
	_, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return err
	}

	// If database has migrations, drop them
	if err == nil || dirty {
		if err := m.Drop(); err != nil {
			return err
		}
	}

	// Run migrations fresh
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func clearTestData(t *testing.T) {
	_, err := testDB.Exec("DELETE FROM product_statistics")
	if err != nil {
		t.Fatalf("Failed to clear product_statistics: %v", err)
	}
	_, err = testDB.Exec("DELETE FROM products")
	if err != nil {
		t.Fatalf("Failed to clear products: %v", err)
	}
}

func TestIntegration_TransferStockWithTransaction_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	clearTestData(t)
	db = testDB // Use test database

	// Arrange: Insert test products
	var fromProductID, toProductID int
	err := testDB.QueryRow(
		"INSERT INTO products (name, description, stock) VALUES ($1, $2, $3) RETURNING id",
		"Source Product", "Has stock", 10,
	).Scan(&fromProductID)
	if err != nil {
		t.Fatalf("Failed to insert source product: %v", err)
	}

	err = testDB.QueryRow(
		"INSERT INTO products (name, description, stock) VALUES ($1, $2, $3) RETURNING id",
		"Target Product", "Needs stock", 5,
	).Scan(&toProductID)
	if err != nil {
		t.Fatalf("Failed to insert target product: %v", err)
	}

	// Act: Transfer 3 units
	err = transferStockWithTransaction(fromProductID, toProductID, 3)
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	// Assert: Verify final states
	var fromStock, toStock int
	err = testDB.QueryRow("SELECT stock FROM products WHERE id = $1", fromProductID).Scan(&fromStock)
	if err != nil {
		t.Fatalf("Failed to get source product stock: %v", err)
	}

	err = testDB.QueryRow("SELECT stock FROM products WHERE id = $1", toProductID).Scan(&toStock)
	if err != nil {
		t.Fatalf("Failed to get target product stock: %v", err)
	}

	if fromStock != 7 {
		t.Errorf("Expected source product to have 7 units, got %d", fromStock)
	}
	if toStock != 8 {
		t.Errorf("Expected target product to have 8 units, got %d", toStock)
	}
}

func TestIntegration_TransferStockWithTransaction_RollbackOnInsufficientStock(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	clearTestData(t)
	db = testDB

	// Arrange: Insert test products
	var fromProductID, toProductID int
	err := testDB.QueryRow(
		"INSERT INTO products (name, description, stock) VALUES ($1, $2, $3) RETURNING id",
		"Source Product", "Not enough stock", 2,
	).Scan(&fromProductID)
	if err != nil {
		t.Fatalf("Failed to insert source product: %v", err)
	}

	err = testDB.QueryRow(
		"INSERT INTO products (name, description, stock) VALUES ($1, $2, $3) RETURNING id",
		"Target Product", "Waiting", 5,
	).Scan(&toProductID)
	if err != nil {
		t.Fatalf("Failed to insert target product: %v", err)
	}

	// Act: Try to transfer more stock than available
	err = transferStockWithTransaction(fromProductID, toProductID, 10)

	// Assert: Should fail
	if err == nil {
		t.Error("Expected error for insufficient stock, got nil")
	}

	// Verify both products unchanged (rollback worked)
	var fromStock, toStock int
	testDB.QueryRow("SELECT stock FROM products WHERE id = $1", fromProductID).Scan(&fromStock)
	testDB.QueryRow("SELECT stock FROM products WHERE id = $1", toProductID).Scan(&toStock)

	if fromStock != 2 {
		t.Errorf("Expected source product unchanged at 2 units, got %d", fromStock)
	}
	if toStock != 5 {
		t.Errorf("Expected target product unchanged at 5 units, got %d", toStock)
	}
}

func TestIntegration_TransferStockWithTransaction_RollbackOnMaxExceeded(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	clearTestData(t)
	db = testDB

	// Arrange: Target product near maximum
	var fromProductID, toProductID int
	err := testDB.QueryRow(
		"INSERT INTO products (name, description, stock) VALUES ($1, $2, $3) RETURNING id",
		"Source Product", "Donor", 20,
	).Scan(&fromProductID)
	if err != nil {
		t.Fatalf("Failed to insert source product: %v", err)
	}

	err = testDB.QueryRow(
		"INSERT INTO products (name, description, stock) VALUES ($1, $2, $3) RETURNING id",
		"Target Product", "Near max", 95,
	).Scan(&toProductID)
	if err != nil {
		t.Fatalf("Failed to insert target product: %v", err)
	}

	// Act: Try to transfer stock that would exceed 100
	err = transferStockWithTransaction(fromProductID, toProductID, 10)

	// Assert: Should fail with business rule violation
	if err == nil {
		t.Error("Expected error for exceeding maximum, got nil")
	}

	// Verify rollback
	var fromStock, toStock int
	testDB.QueryRow("SELECT stock FROM products WHERE id = $1", fromProductID).Scan(&fromStock)
	testDB.QueryRow("SELECT stock FROM products WHERE id = $1", toProductID).Scan(&toStock)

	if fromStock != 20 {
		t.Errorf("Expected source product unchanged at 20 units, got %d", fromStock)
	}
	if toStock != 95 {
		t.Errorf("Expected target product unchanged at 95 units, got %d", toStock)
	}
}

func TestIntegration_TransferStockWithoutTransaction_DataCorruption(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	clearTestData(t)
	db = testDB

	// Arrange
	var fromProductID, toProductID int
	err := testDB.QueryRow(
		"INSERT INTO products (name, description, stock) VALUES ($1, $2, $3) RETURNING id",
		"Source Product", "Will lose stock", 10,
	).Scan(&fromProductID)
	if err != nil {
		t.Fatalf("Failed to insert source product: %v", err)
	}

	err = testDB.QueryRow(
		"INSERT INTO products (name, description, stock) VALUES ($1, $2, $3) RETURNING id",
		"Target Product", "Won't receive stock", 5,
	).Scan(&toProductID)
	if err != nil {
		t.Fatalf("Failed to insert target product: %v", err)
	}

	// Act: Call function without transaction (always fails after first operation)
	err = transferStockWithoutTransaction(fromProductID, toProductID, 3)

	// Assert: Should fail
	if err == nil {
		t.Error("Expected error from simulated failure")
	}

	// Verify data corruption: stock deducted but never added
	var fromStock, toStock int
	testDB.QueryRow("SELECT stock FROM products WHERE id = $1", fromProductID).Scan(&fromStock)
	testDB.QueryRow("SELECT stock FROM products WHERE id = $1", toProductID).Scan(&toStock)

	// This demonstrates the problem
	if fromStock != 7 {
		t.Errorf("Expected source product to have 7 units (deducted), got %d", fromStock)
	}
	if toStock != 5 {
		t.Errorf("Expected target product unchanged at 5 units (never added), got %d", toStock)
	}

	// 3 units disappeared! Total before: 15, Total after: 12
	total := fromStock + toStock
	if total != 12 {
		t.Errorf("Data corruption: expected 12 total stock, got %d", total)
	}
}

func TestIntegration_MoveStockWithTransaction_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	clearTestData(t)
	db = testDB

	// Arrange: Create product 1 and target product
	_, err := testDB.Exec(
		"INSERT INTO products (id, name, description, stock) VALUES ($1, $2, $3, $4)",
		1, "Product 1", "Receives stock", 0,
	)
	if err != nil {
		t.Fatalf("Failed to insert product 1: %v", err)
	}

	var targetProductID int
	err = testDB.QueryRow(
		"INSERT INTO products (name, description, stock) VALUES ($1, $2, $3) RETURNING id",
		"Target Product", "Loses unit", 5,
	).Scan(&targetProductID)
	if err != nil {
		t.Fatalf("Failed to insert target product: %v", err)
	}

	// Act
	err = moveStockWithTransaction(targetProductID)
	if err != nil {
		t.Fatalf("Add unit failed: %v", err)
	}

	// Assert
	var product1Stock, targetStock int
	testDB.QueryRow("SELECT stock FROM products WHERE id = 1").Scan(&product1Stock)
	testDB.QueryRow("SELECT stock FROM products WHERE id = $1", targetProductID).Scan(&targetStock)

	if product1Stock != 1 {
		t.Errorf("Expected product 1 to have 1 unit, got %d", product1Stock)
	}
	if targetStock != 4 {
		t.Errorf("Expected target product to have 4 units, got %d", targetStock)
	}
}
