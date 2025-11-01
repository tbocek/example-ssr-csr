package main

import (
	"context"
	"database/sql"
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
	testDB             *sql.DB
	postgresContainer  *testcontainerspostgres.PostgresContainer
	testContainerCtx   context.Context
)

func TestMain(m *testing.M) {
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
	_, err := testDB.Exec("DELETE FROM game_statistics")
	if err != nil {
		t.Fatalf("Failed to clear game_statistics: %v", err)
	}
	_, err = testDB.Exec("DELETE FROM games")
	if err != nil {
		t.Fatalf("Failed to clear games: %v", err)
	}
}

func TestIntegration_TransferStarsWithTransaction_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	clearTestData(t)
	db = testDB // Use test database

	// Arrange: Insert test games
	var fromGameID, toGameID int
	err := testDB.QueryRow(
		"INSERT INTO games (title, description, stars) VALUES ($1, $2, $3) RETURNING id",
		"Source Game", "Has stars", 10,
	).Scan(&fromGameID)
	if err != nil {
		t.Fatalf("Failed to insert source game: %v", err)
	}

	err = testDB.QueryRow(
		"INSERT INTO games (title, description, stars) VALUES ($1, $2, $3) RETURNING id",
		"Target Game", "Needs stars", 5,
	).Scan(&toGameID)
	if err != nil {
		t.Fatalf("Failed to insert target game: %v", err)
	}

	// Act: Transfer 3 stars
	err = transferStarsWithTransaction(fromGameID, toGameID, 3)
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	// Assert: Verify final states
	var fromStars, toStars int
	err = testDB.QueryRow("SELECT stars FROM games WHERE id = $1", fromGameID).Scan(&fromStars)
	if err != nil {
		t.Fatalf("Failed to get source game stars: %v", err)
	}

	err = testDB.QueryRow("SELECT stars FROM games WHERE id = $1", toGameID).Scan(&toStars)
	if err != nil {
		t.Fatalf("Failed to get target game stars: %v", err)
	}

	if fromStars != 7 {
		t.Errorf("Expected source game to have 7 stars, got %d", fromStars)
	}
	if toStars != 8 {
		t.Errorf("Expected target game to have 8 stars, got %d", toStars)
	}
}

func TestIntegration_TransferStarsWithTransaction_RollbackOnInsufficientStars(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	clearTestData(t)
	db = testDB

	// Arrange: Insert test games
	var fromGameID, toGameID int
	err := testDB.QueryRow(
		"INSERT INTO games (title, description, stars) VALUES ($1, $2, $3) RETURNING id",
		"Source Game", "Not enough stars", 2,
	).Scan(&fromGameID)
	if err != nil {
		t.Fatalf("Failed to insert source game: %v", err)
	}

	err = testDB.QueryRow(
		"INSERT INTO games (title, description, stars) VALUES ($1, $2, $3) RETURNING id",
		"Target Game", "Waiting", 5,
	).Scan(&toGameID)
	if err != nil {
		t.Fatalf("Failed to insert target game: %v", err)
	}

	// Act: Try to transfer more stars than available
	err = transferStarsWithTransaction(fromGameID, toGameID, 10)

	// Assert: Should fail
	if err == nil {
		t.Error("Expected error for insufficient stars, got nil")
	}

	// Verify both games unchanged (rollback worked)
	var fromStars, toStars int
	testDB.QueryRow("SELECT stars FROM games WHERE id = $1", fromGameID).Scan(&fromStars)
	testDB.QueryRow("SELECT stars FROM games WHERE id = $1", toGameID).Scan(&toStars)

	if fromStars != 2 {
		t.Errorf("Expected source game unchanged at 2 stars, got %d", fromStars)
	}
	if toStars != 5 {
		t.Errorf("Expected target game unchanged at 5 stars, got %d", toStars)
	}
}

func TestIntegration_TransferStarsWithTransaction_RollbackOnMaxExceeded(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	clearTestData(t)
	db = testDB

	// Arrange: Target game near maximum
	var fromGameID, toGameID int
	err := testDB.QueryRow(
		"INSERT INTO games (title, description, stars) VALUES ($1, $2, $3) RETURNING id",
		"Source Game", "Donor", 20,
	).Scan(&fromGameID)
	if err != nil {
		t.Fatalf("Failed to insert source game: %v", err)
	}

	err = testDB.QueryRow(
		"INSERT INTO games (title, description, stars) VALUES ($1, $2, $3) RETURNING id",
		"Target Game", "Near max", 95,
	).Scan(&toGameID)
	if err != nil {
		t.Fatalf("Failed to insert target game: %v", err)
	}

	// Act: Try to transfer stars that would exceed 100
	err = transferStarsWithTransaction(fromGameID, toGameID, 10)

	// Assert: Should fail with business rule violation
	if err == nil {
		t.Error("Expected error for exceeding maximum, got nil")
	}

	// Verify rollback
	var fromStars, toStars int
	testDB.QueryRow("SELECT stars FROM games WHERE id = $1", fromGameID).Scan(&fromStars)
	testDB.QueryRow("SELECT stars FROM games WHERE id = $1", toGameID).Scan(&toStars)

	if fromStars != 20 {
		t.Errorf("Expected source game unchanged at 20 stars, got %d", fromStars)
	}
	if toStars != 95 {
		t.Errorf("Expected target game unchanged at 95 stars, got %d", toStars)
	}
}

func TestIntegration_TransferStarsWithoutTransaction_DataCorruption(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	clearTestData(t)
	db = testDB

	// Arrange
	var fromGameID, toGameID int
	err := testDB.QueryRow(
		"INSERT INTO games (title, description, stars) VALUES ($1, $2, $3) RETURNING id",
		"Source Game", "Will lose stars", 10,
	).Scan(&fromGameID)
	if err != nil {
		t.Fatalf("Failed to insert source game: %v", err)
	}

	err = testDB.QueryRow(
		"INSERT INTO games (title, description, stars) VALUES ($1, $2, $3) RETURNING id",
		"Target Game", "Won't receive stars", 5,
	).Scan(&toGameID)
	if err != nil {
		t.Fatalf("Failed to insert target game: %v", err)
	}

	// Act: Call function without transaction (always fails after first operation)
	err = transferStarsWithoutTransaction(fromGameID, toGameID, 3)

	// Assert: Should fail
	if err == nil {
		t.Error("Expected error from simulated failure")
	}

	// Verify data corruption: stars deducted but never added
	var fromStars, toStars int
	testDB.QueryRow("SELECT stars FROM games WHERE id = $1", fromGameID).Scan(&fromStars)
	testDB.QueryRow("SELECT stars FROM games WHERE id = $1", toGameID).Scan(&toStars)

	// This demonstrates the problem
	if fromStars != 7 {
		t.Errorf("Expected source game to have 7 stars (deducted), got %d", fromStars)
	}
	if toStars != 5 {
		t.Errorf("Expected target game unchanged at 5 stars (never added), got %d", toStars)
	}

	// 3 stars disappeared! Total before: 15, Total after: 12
	total := fromStars + toStars
	if total != 12 {
		t.Errorf("Data corruption: expected 12 total stars, got %d", total)
	}
}

func TestIntegration_AddStarWithTransaction_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	clearTestData(t)
	db = testDB

	// Arrange: Create game 1 and target game
	_, err := testDB.Exec(
		"INSERT INTO games (id, title, description, stars) VALUES ($1, $2, $3, $4)",
		1, "Game 1", "Receives stars", 0,
	)
	if err != nil {
		t.Fatalf("Failed to insert game 1: %v", err)
	}

	var targetGameID int
	err = testDB.QueryRow(
		"INSERT INTO games (title, description, stars) VALUES ($1, $2, $3) RETURNING id",
		"Target Game", "Loses star", 5,
	).Scan(&targetGameID)
	if err != nil {
		t.Fatalf("Failed to insert target game: %v", err)
	}

	// Act
	err = addStarWithTransaction(targetGameID)
	if err != nil {
		t.Fatalf("Add star failed: %v", err)
	}

	// Assert
	var game1Stars, targetStars int
	testDB.QueryRow("SELECT stars FROM games WHERE id = 1").Scan(&game1Stars)
	testDB.QueryRow("SELECT stars FROM games WHERE id = $1", targetGameID).Scan(&targetStars)

	if game1Stars != 1 {
		t.Errorf("Expected game 1 to have 1 star, got %d", game1Stars)
	}
	if targetStars != 4 {
		t.Errorf("Expected target game to have 4 stars, got %d", targetStars)
	}
}